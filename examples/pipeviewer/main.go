package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/filterableviewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/item"
)

var (
	lineNumberStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

type object struct {
	lineNumber item.SingleItem
	content    item.SingleItem
}

func (o object) GetItem() item.Item {
	return item.NewMulti(o.lineNumber, o.content)
}

type appKeys struct {
	quit                     key.Binding
	toggleShowLineNumbersKey key.Binding
	toggleWrapTextKey        key.Binding
	toggleSelectionKey       key.Binding
}

var appKeyMap = appKeys{
	quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	toggleShowLineNumbersKey: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "toggle line numbers"),
	),
	toggleWrapTextKey: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "toggle wrapping"),
	),
	toggleSelectionKey: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "toggle selection"),
	),
}

var viewportKeyMap = viewport.DefaultKeyMap()
var filterableViewportKeyMap = filterableviewport.DefaultKeyMap()
var styles = filterableviewport.DefaultStyles()

type newLinesMsg struct {
	lines []string
}

type stdinDoneMsg struct{}

type model struct {
	itemNumber                    int
	showLineNumbers               bool
	objects                       []object
	fv                            *filterableviewport.Model[object]
	ready                         bool
	viewportWidth, viewportHeight int
}

func stdinIsPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func (m model) Init() tea.Cmd {
	if stdinIsPipe() {
		return readStdinCmd()
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case newLinesMsg:
		if m.ready && len(msg.lines) > 0 {
			newObjects := make([]object, len(msg.lines))
			for i, line := range msg.lines {
				newObjects[i] = object{
					lineNumber: item.NewItem(""),
					content:    item.NewItem(line),
				}
			}
			m.objects = append(m.objects, newObjects...)
			m.fv.AppendObjects(newObjects)
			m.itemNumber += len(msg.lines)
		}
		if stdinIsPipe() {
			return m, readStdinCmd()
		}
		return m, nil

	case stdinDoneMsg:
		return m, nil

	case tea.KeyMsg:
		// allow quitting with ctrl+c even when filter is focused
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
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
		case key.Matches(msg, appKeyMap.toggleShowLineNumbersKey):
			m.showLineNumbers = !m.showLineNumbers
			for i := range m.objects {
				lineNum := ""
				if m.showLineNumbers {
					lineNum = fmt.Sprintf("%d", i+1) + " "
				}
				m.objects[i].lineNumber = item.NewItem(lineNumberStyle.Render(lineNum))
			}
			m.fv.SetObjects(m.objects)
			return m, nil
		case key.Matches(msg, appKeyMap.toggleWrapTextKey):
			m.fv.SetWrapText(!m.fv.GetWrapText())
			return m, nil
		case key.Matches(msg, appKeyMap.toggleSelectionKey):
			m.fv.SetSelectionEnabled(!m.fv.GetSelectionEnabled())
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.viewportWidth, m.viewportHeight = msg.Width, msg.Height
		if !m.ready {
			// configure file saving
			homeDir, _ := os.UserHomeDir()
			saveDir := filepath.Join(homeDir, ".pipeviewer", "saved")
			saveKey := key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save"))

			vp := viewport.New[object](
				m.viewportWidth,
				m.viewportHeight,
				viewport.WithKeyMap[object](viewportKeyMap),
				viewport.WithStyles[object](viewport.DefaultStyles()),
				viewport.WithStickyBottom[object](true),
				viewport.WithFileSaving[object](saveDir, saveKey),
			)
			m.fv = filterableviewport.New[object](
				vp,
				filterableviewport.WithKeyMap[object](filterableViewportKeyMap),
				filterableviewport.WithStyles[object](styles),
				filterableviewport.WithPrefixText[object]("Filter:"),
				filterableviewport.WithEmptyText[object]("Press / to filter"),
				filterableviewport.WithMatchingItemsOnly[object](false),
				filterableviewport.WithCanToggleMatchingItemsOnly[object](true),
				filterableviewport.WithHorizontalPad[object](50),
				filterableviewport.WithVerticalPad[object](20),
			)
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
	return m.fv.View()
}

var stdinReader *bufio.Reader

func init() {
	stdinReader = bufio.NewReader(os.Stdin)
}

// readStdinCmd reads lines from stdin in batches and returns them as a message
func readStdinCmd() tea.Cmd {
	return func() tea.Msg {
		const maxBatchSize = 500
		lines := make([]string, 0, maxBatchSize)

		for i := 0; i < maxBatchSize; i++ {
			line, err := stdinReader.ReadString('\n')
			if err == io.EOF {
				if len(line) > 0 {
					lines = append(lines, strings.TrimSuffix(line, "\n"))
				}
				if len(lines) > 0 {
					return newLinesMsg{lines: lines}
				}
				return stdinDoneMsg{}
			}
			if err != nil {
				if len(lines) > 0 {
					return newLinesMsg{lines: lines}
				}
				return stdinDoneMsg{}
			}
			lines = append(lines, strings.TrimSuffix(line, "\n"))

			// check if more data is immediately available, otherwise return what we have
			// this prevents blocking when data arrives slowly
			if stdinReader.Buffered() == 0 {
				break
			}
		}

		if len(lines) > 0 {
			return newLinesMsg{lines: lines}
		}
		return stdinDoneMsg{}
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
		model{},
		tea.WithAltScreen(),
		tea.WithInput(tty),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running program: %v\n", err)
		os.Exit(1)
	}
}
