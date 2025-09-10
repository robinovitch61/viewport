package main

// An example program demonstrating the viewport component

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/examples/text"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/item"
)

type object struct {
	item item.Item
}

func (o object) GetItem() item.Item {
	return o.item
}

var keyMap = viewport.DefaultKeyMap()
var styles = viewport.DefaultStyles()

type model struct {
	// viewport is the container for the lines
	viewport *viewport.Model[object]

	// lines contains the lines to be displayed in the viewport
	lines []object

	// ready indicates whether the model has been initialized
	ready bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}
		if k := msg.String(); k == "w" {
			m.viewport.SetWrapText(!m.viewport.GetWrapText())
		}
		if k := msg.String(); k == "s" {
			m.viewport.SetSelectionEnabled(!m.viewport.GetSelectionEnabled())
		}

	case tea.WindowSizeMsg:
		// 2 for border, 4 for content above viewport
		viewportWidth, viewportHeight := msg.Width-2, msg.Height-4-2
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New[object](
				viewportWidth,
				viewportHeight,
				viewport.WithKeyMap[object](keyMap),
				viewport.WithStyles[object](styles),
			)
			m.viewport.SetObjects(m.lines)
			m.viewport.SetSelectionEnabled(false)
			m.viewport.SetWrapText(true)
			m.ready = true
		} else {
			m.viewport.SetWidth(viewportWidth)
			m.viewport.SetHeight(viewportHeight)
		}
	}

	// Handle keyboard events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Initializing viewport..."
	}
	var header = strings.Join(getHeader(
		m.viewport.GetWrapText(),
		m.viewport.GetSelectionEnabled(),
		[]key.Binding{
			keyMap.PageDown,
			keyMap.PageUp,
			keyMap.HalfPageUp,
			keyMap.HalfPageDown,
			keyMap.Up,
			keyMap.Down,
			keyMap.Left,
			keyMap.Right,
			keyMap.Top,
			keyMap.Bottom,
		},
	), "\n")
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m.viewport.View()),
	)
}

func getHeader(wrapped, selectionEnabled bool, bindings []key.Binding) []string {
	var header []string
	header = append(header, lipgloss.NewStyle().Bold(true).Render("A Supercharged Viewport (q/ctrl+c/esc to quit)"))
	header = append(header, "- Wrapping enabled: "+fmt.Sprint(wrapped)+" (w to toggle)")
	header = append(header, "- Selection enabled: "+fmt.Sprint(selectionEnabled)+" (s to toggle)")
	header = append(header, getShortHelp(bindings))
	return header
}

func getShortHelp(bindings []key.Binding) string {
	var output string
	for _, km := range bindings {
		output += km.Help().Key + " " + km.Help().Desc + "  "
	}
	output = strings.TrimSpace(output)
	return output
}

func main() {
	lines := strings.Split(text.ExampleContent, "\n")
	renderableLines := make([]object, len(lines))
	for i, line := range lines {
		renderableLines[i] = object{item: item.NewItem(line)}
	}

	p := tea.NewProgram(
		model{lines: renderableLines},
		tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
