package filterable_viewport

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/bubbleo/viewport"
)

type filterMode int

const (
	filterModeOff filterMode = iota
	filterModeEditing
	filterModeApplied
)

type Model[T viewport.Renderable] struct {
	Viewport *viewport.Model[T]

	keyMap        KeyMap
	textInput     textinput.Model
	filterMode    filterMode
	items         []T
	matchingItems int
}

// New creates a new filterable viewport model
func New[T viewport.Renderable](width, height int, km KeyMap, styles viewport.Styles) *Model[T] {
	ti := textinput.New()
	ti.CharLimit = 0

	viewportHeight := height - 1 // -1 for filter line
	vp := viewport.New[T](width, viewportHeight, km.ViewportKeyMap, styles)

	return &Model[T]{
		Viewport:   vp,
		keyMap:     km,
		filterMode: filterModeOff,
		textInput:  ti,
	}
}

func (m *Model[T]) Init() tea.Cmd {
	return nil
}

func (m *Model[T]) Update(msg tea.Msg) (*Model[T], tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.FilterKey):
			if m.filterMode != filterModeEditing {
				m.textInput.Focus()
				m.filterMode = filterModeEditing
				m.matchingItems = matchingItems(m.filterMode, m.items, m.textInput.Value())
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.ApplyFilterKey):
			if m.filterMode == filterModeEditing {
				m.textInput.Blur()
				m.filterMode = filterModeApplied
				m.matchingItems = matchingItems(m.filterMode, m.items, m.textInput.Value())
				return m, nil
			}
		case key.Matches(msg, m.keyMap.CancelFilterKey):
			m.filterMode = filterModeOff
			m.textInput.Blur()
			m.textInput.SetValue("")
			m.matchingItems = matchingItems(m.filterMode, m.items, m.textInput.Value())
			return m, nil
		}

		if m.filterMode == filterModeEditing {
			m.textInput, cmd = m.textInput.Update(msg)
			m.matchingItems = matchingItems(m.filterMode, m.items, m.textInput.Value())
			cmds = append(cmds, cmd)
		}
	}

	if m.filterMode != filterModeEditing {
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model[T]) View() string {
	filterLine := m.renderFilterLine()
	viewportView := m.Viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left, filterLine, viewportView)
}

// SetContent sets the content and updates total item count
func (m *Model[T]) SetContent(items []T) {
	m.Viewport.SetContent(items)
	m.items = items
	m.matchingItems = matchingItems(m.filterMode, m.items, m.textInput.Value())
}

// SetWidth updates the width of both the viewport and textinput
func (m *Model[T]) SetWidth(width int) {
	m.Viewport.SetWidth(width)
	m.textInput.SetWidth(width)
}

// SetHeight updates the height, accounting for the filter line
func (m *Model[T]) SetHeight(height int) {
	m.Viewport.SetHeight(height - 1) // -1 for filter line
}

// FilterFocused returns true if the filter text input is focused
func (m *Model[T]) FilterFocused() bool {
	return m.textInput.Focused()
}

func (m *Model[T]) renderFilterLine() string {
	switch m.filterMode {
	case filterModeOff:
		return "No filter"
	case filterModeEditing:
		if m.textInput.Value() == "" {
			return m.textInput.View() + " " + matchCountText(m.matchingItems, len(m.items))
		}
		return m.textInput.View() + " " + matchCountText(m.matchingItems, len(m.items))
	case filterModeApplied:
		if m.textInput.Value() == "" {
			return "No filter"
		}
		return m.textInput.View() + " " + matchCountText(m.matchingItems, len(m.items))
	default:
		panic(fmt.Sprintf("invalid filter mode: %d", m.filterMode))
	}
}

func matchingItems[T viewport.Renderable](
	mode filterMode,
	items []T,
	filter string,
) int {
	if mode == filterModeOff || filter == "" {
		return len(items)
	}

	count := 0
	for i := range items {
		if strings.Contains(items[i].Render().Content(), filter) {
			count++
		}
	}
	return count
}

func matchCountText(matching, total int) string {
	if matching == 0 {
		return "(no matches)"
	}
	return fmt.Sprintf("(%d / %d items match)", matching, total)
}
