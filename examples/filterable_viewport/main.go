package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/examples/common"
	"github.com/robinovitch61/bubbleo/filterable_viewport"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
)

type appKeys struct {
	quit               key.Binding
	toggleWrapTextKey  key.Binding
	toggleSelectionKey key.Binding
}

var appKeyMap = appKeys{
	quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
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

var keyMap = filterable_viewport.DefaultKeyMap()
var styles = viewport.DefaultStyles()

type model struct {
	// fv is the filterable container for the lines
	fv *filterable_viewport.Model[viewport.Item]

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
		// quit is always available
		if key.Matches(msg, appKeyMap.quit) {
			return m, tea.Quit
		}

		if !m.ready {
			// if the viewport not ready, only handle quitting
			return m, nil
		}
		// if the filterable viewport is focused, handle its messages
		if m.fv.FilterFocused() {
			m.fv, cmd = m.fv.Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, appKeyMap.toggleWrapTextKey):
			m.fv.Viewport.SetWrapText(!m.fv.Viewport.GetWrapText())
			return m, nil
		case key.Matches(msg, appKeyMap.toggleSelectionKey):
			m.fv.Viewport.SetSelectionEnabled(!m.fv.Viewport.GetSelectionEnabled())
			return m, nil
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
			m.fv = filterable_viewport.New[viewport.Item](
				viewportWidth, 
				viewportHeight,
				filterable_viewport.WithKeyMap[viewport.Item](keyMap),
				filterable_viewport.WithStyles[viewport.Item](styles),
			)
			m.fv.SetContent(m.lines)
			m.fv.Viewport.SetSelectionEnabled(false)
			m.fv.Viewport.SetWrapText(true)
			m.ready = true
		} else {
			m.fv.SetWidth(viewportWidth)
			m.fv.SetHeight(viewportHeight)
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
		return "\n  Initializing..."
	}
	var header = strings.Join(getHeader(
		m.fv.Viewport.GetWrapText(),
		m.fv.Viewport.GetSelectionEnabled(),
		keyMap.ViewportKeyMap,
		[]key.Binding{
			keyMap.FilterKey,
			keyMap.RegexFilterKey,
			keyMap.ApplyFilterKey,
			keyMap.CancelFilterKey,
			keyMap.ToggleMatchesOnlyKey,
		},
	), "\n")
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m.fv.View()),
	)
}

func getHeader(wrapped, selectionEnabled bool, viewportKeyMap viewport.KeyMap, bindings []key.Binding) []string {
	var header []string
	header = append(header, lipgloss.NewStyle().Bold(true).Render("A Supercharged Filterable Viewport"+fmt.Sprintf(" (%s to quit)", appKeyMap.quit.Help().Key)))
	header = append(header, "- Wrapping enabled: "+fmt.Sprint(wrapped)+fmt.Sprintf(" (%s to toggle)", appKeyMap.toggleWrapTextKey.Help().Key))
	header = append(header, "- Selection enabled: "+fmt.Sprint(selectionEnabled)+fmt.Sprintf(" (%s to toggle)", appKeyMap.toggleSelectionKey.Help().Key))
	header = append(header, getShortHelp([]key.Binding{
		viewportKeyMap.PageDown,
		viewportKeyMap.PageUp,
		viewportKeyMap.HalfPageUp,
		viewportKeyMap.HalfPageDown,
		viewportKeyMap.Up,
		viewportKeyMap.Down,
		viewportKeyMap.Left,
		viewportKeyMap.Right,
		viewportKeyMap.Top,
		viewportKeyMap.Bottom,
	}))
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
