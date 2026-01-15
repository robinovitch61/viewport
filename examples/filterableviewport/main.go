package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/examples/text"
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

var viewportKeyMap = viewport.DefaultKeyMap()
var filterableViewportKeyMap = filterableviewport.DefaultKeyMap()
var styles = filterableviewport.DefaultStyles()

type model struct {
	// fv is the filterable container for the objects
	fv *filterableviewport.Model[object]

	// objects contains the objects to be displayed in the viewport
	objects []object

	// ready indicates whether the model has been initialized
	ready bool

	// width and height of the viewport
	viewportWidth, viewportHeight int
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
		// if the filterable viewport is capturing input, forward messages to it
		if m.fv.IsCapturingInput() {
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
		}

	case tea.WindowSizeMsg:
		// 2 for border, 5 for content above viewport
		m.viewportWidth, m.viewportHeight = msg.Width-2, msg.Height-5-2
		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			vp := viewport.New[object](
				m.viewportWidth,
				m.viewportHeight,
				viewport.WithKeyMap[object](viewportKeyMap),
				viewport.WithStyles[object](viewport.DefaultStyles()),
			)
			m.fv = filterableviewport.New[object](
				vp,
				filterableviewport.WithKeyMap[object](filterableViewportKeyMap),
				filterableviewport.WithStyles[object](styles),
				filterableviewport.WithPrefixText[object]("Filter:"),
				filterableviewport.WithEmptyText[object]("No Current Filter"),
				filterableviewport.WithMatchingItemsOnly[object](false),
				filterableviewport.WithCanToggleMatchingItemsOnly[object](true),
				filterableviewport.WithVerticalPad[object](10),
				filterableviewport.WithHorizontalPad[object](50),
			)
			m.fv.SetObjects(m.objects)
			m.fv.SetSelectionEnabled(false)
			m.fv.SetWrapText(true)
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
		return "Initializing filterable viewport..."
	}
	var header = strings.Join(getHeader(
		m.fv.GetWrapText(),
		m.fv.GetSelectionEnabled(),
		viewportKeyMap,
		[]key.Binding{
			filterableViewportKeyMap.FilterKey,
			filterableViewportKeyMap.RegexFilterKey,
			filterableViewportKeyMap.ApplyFilterKey,
			filterableViewportKeyMap.CancelFilterKey,
			filterableViewportKeyMap.ToggleMatchingItemsOnlyKey,
		},
		m.viewportWidth,
		m.viewportHeight,
	), "\n")
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m.fv.View()),
	)
}

func getHeader(wrapped, selectionEnabled bool, viewportKeyMap viewport.KeyMap, bindings []key.Binding, vpWidth, vpHeight int) []string {
	var header []string
	suffix := fmt.Sprintf(" (%s to quit) [viewport is %d by %d]", appKeyMap.quit.Help().Key, vpWidth, vpHeight)
	header = append(header, lipgloss.NewStyle().Bold(true).Render("A Supercharged Filterable Viewport"+suffix))
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
	lines := strings.Split(text.ExampleContent, "\n")
	objects := make([]object, len(lines))
	for i, line := range lines {
		objects[i] = object{item: item.NewItem(line)}
	}

	p := tea.NewProgram(
		model{objects: objects},
		tea.WithAltScreen(), // use the full size of the terminal in its "alternate screen buffer"
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
