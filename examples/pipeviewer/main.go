package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/filterableviewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/item"
)

type object struct {
	item item.Item
}

func (o object) GetItem() item.Item {
	return o.item
}

type appKeys struct {
	quit               key.Binding
	toggleWrapTextKey  key.Binding
	toggleSelectionKey key.Binding
	save               key.Binding
}

var appKeyMap = appKeys{
	quit: key.NewBinding(
		key.WithKeys("ctrl+c", "ctrl+d"),
		key.WithHelp("ctrl+c/ctrl+d", "quit"),
	),
	toggleWrapTextKey: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "toggle wrapping"),
	),
	toggleSelectionKey: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle selection"),
	),
	save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save to file"),
	),
}

var viewportKeyMap = viewport.DefaultKeyMap()
var filterableViewportKeyMap = filterableviewport.DefaultKeyMap()
var styles = filterableviewport.DefaultStyles()

type newLineMsg struct {
	line string
}

type stdinDoneMsg struct{}

type fileSavedMsg struct {
	filename string
	err      error
}

type model struct {
	fv                            *filterableviewport.Model[object]
	objects                       []object
	ready                         bool
	stdinDone                     bool
	saveStatus                    string
	viewportWidth, viewportHeight int
}

func (m model) Init() tea.Cmd {
	return readStdinCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case newLineMsg:
		m.objects = append(m.objects, object{item: item.NewItem(msg.line)})
		if m.ready {
			m.fv.SetObjects(m.objects)
		}
		return m, readStdinCmd()

	case stdinDoneMsg:
		m.stdinDone = true
		return m, nil

	case fileSavedMsg:
		if msg.err != nil {
			m.saveStatus = fmt.Sprintf("Error saving: %v", msg.err)
		} else {
			m.saveStatus = fmt.Sprintf("Saved to %s", msg.filename)
		}
		return m, nil

	case tea.KeyMsg:
		// allow quitting with ctrl+c and ctrl+d even when filter is focused
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "ctrl+d"))) {
			return m, tea.Quit
		}

		if !m.ready {
			return m, nil
		}

		if m.fv.FilterFocused() {
			m.fv, cmd = m.fv.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, appKeyMap.toggleWrapTextKey):
			m.fv.SetWrapText(!m.fv.GetWrapText())
			return m, nil
		case key.Matches(msg, appKeyMap.toggleSelectionKey):
			m.fv.SetSelectionEnabled(!m.fv.GetSelectionEnabled())
			return m, nil
		case key.Matches(msg, appKeyMap.save):
			return m, saveToFileCmd(m.objects)
		}

	case tea.WindowSizeMsg:
		m.viewportWidth, m.viewportHeight = msg.Width, msg.Height
		if !m.ready {
			vp := viewport.New[object](
				m.viewportWidth,
				m.viewportHeight,
				viewport.WithKeyMap[object](viewportKeyMap),
				viewport.WithStyles[object](viewport.DefaultStyles()),
				viewport.WithStickyBottom[object](true),
			)
			m.fv = filterableviewport.New[object](
				vp,
				filterableviewport.WithKeyMap[object](filterableViewportKeyMap),
				filterableviewport.WithStyles[object](styles),
				filterableviewport.WithPrefixText[object]("Filter:"),
				filterableviewport.WithEmptyText[object]("Press / to filter"),
				filterableviewport.WithMatchingItemsOnly[object](false),
				filterableviewport.WithCanToggleMatchingItemsOnly[object](true),
			)
			m.fv.SetObjects(m.objects)
			m.fv.SetWrapText(false)
			m.fv.SetSelectionEnabled(false)
			m.ready = true
		} else {
			m.fv.SetWidth(m.viewportWidth)
			m.fv.SetHeight(m.viewportHeight)
		}
	}

	if m.ready {
		m.fv, cmd = m.fv.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	view := m.fv.View()

	// show save status if present
	if m.saveStatus != "" {
		return view + "\n" + m.saveStatus
	}

	return view
}

var stdinReader *bufio.Reader

func init() {
	stdinReader = bufio.NewReader(os.Stdin)
}

// readStdinCmd reads a single line from stdin and returns it as a message
func readStdinCmd() tea.Cmd {
	return func() tea.Msg {
		line, err := stdinReader.ReadString('\n')
		if err == io.EOF {
			if len(line) > 0 {
				// return the last line without newline, then done
				return newLineMsg{line: strings.TrimSuffix(line, "\n")}
			}
			return stdinDoneMsg{}
		}
		if err != nil {
			return stdinDoneMsg{}
		}
		return newLineMsg{line: strings.TrimSuffix(line, "\n")}
	}
}

// saveToFileCmd saves the current objects to a timestamped file
func saveToFileCmd(objects []object) tea.Cmd {
	return func() tea.Msg {
		// generate timestamp-based filename
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("pipeviewer-%s.txt", timestamp)

		// collect all lines
		var content strings.Builder
		for _, obj := range objects {
			content.WriteString(obj.item.Content())
			content.WriteString("\n")
		}

		// write to file
		err := os.WriteFile(filename, []byte(content.String()), 0600)
		return fileSavedMsg{
			filename: filename,
			err:      err,
		}
	}
}

func main() {
	// open /dev/tty for input since stdin is used for piped data
	tty, err := os.Open("/dev/tty")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := tty.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing tty: %v\n", err)
		}
	}()

	p := tea.NewProgram(
		model{
			objects: []object{},
		},
		tea.WithAltScreen(),
		tea.WithInput(tty),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running program: %v\n", err)
		os.Exit(1)
	}
}
