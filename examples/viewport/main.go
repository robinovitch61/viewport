package main

// An example program demonstrating the viewport component

import (
	"fmt"
	"os"
	"strings"

	"github.com/robinovitch61/bubbleo/examples/common"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/viewport"
)

var keyMap = viewport.DefaultKeyMap()
var styles = viewport.DefaultStyles()

type model struct {
	// viewport is the container for the lines
	viewport *viewport.Model[viewport.Item]

	// lines contains the lines to be displayed in the viewport
	lines []viewport.Item

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
		// 2 for border, 5 for content above viewport
		viewportWidth, viewportHeight := msg.Width-2, msg.Height-5-2
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New[viewport.Item](
				viewportWidth,
				viewportHeight,
				viewport.WithKeyMap[viewport.Item](keyMap),
				viewport.WithStyles[viewport.Item](styles),
			)
			m.viewport.SetContent(m.lines)
			m.viewport.SetSelectionEnabled(false)
			m.viewport.SetStringToHighlight("surf")
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
	header = append(header, "- Text to highlight: 'surf'")
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
	lines := strings.Split(common.ExampleContent, "\n")
	renderableLines := make([]viewport.Item, len(lines))
	for i, line := range lines {
		renderableLines[i] = viewport.Item{LineBuffer: linebuffer.New(line)}
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
