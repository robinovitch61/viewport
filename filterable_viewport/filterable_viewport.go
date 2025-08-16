package filterable_viewport

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/bubbleo/viewport"
)

type FilterMode int

const (
	FilterModeOff FilterMode = iota
	FilterModeEditing
	FilterModeApplied
)

type Model[T viewport.Renderable] struct {
	Viewport *viewport.Model[T]

	keyMap        KeyMap
	textInput     textinput.Model
	filterMode    FilterMode
	totalItems    int
	matchingItems int
}

// New creates a new filterable viewport model
func New[T viewport.Renderable](width, height int, km KeyMap, styles viewport.Styles) *Model[T] {
	ti := textinput.New()
	ti.CharLimit = 0

	viewportHeight := height - 1 // -1 for filter line
	vp := viewport.New[T](width, viewportHeight, km.ViewportKeyMap, styles)

	return &Model[T]{
		Viewport:      vp,
		keyMap:        km,
		filterMode:    FilterModeOff,
		textInput:     ti,
		totalItems:    0,
		matchingItems: 0,
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
			if m.filterMode != FilterModeEditing {
				m.filterMode = FilterModeEditing
				m.textInput.Focus()
				m.updateMatchCounts()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.ApplyFilterKey):
			if m.filterMode == FilterModeEditing {
				m.filterMode = FilterModeApplied
				m.textInput.Blur()
				m.updateMatchCounts()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.CancelFilterKey):
			m.filterMode = FilterModeOff
			m.textInput.SetValue("")
			m.textInput.Blur()
			m.updateMatchCounts()
			return m, nil
		}

		if m.filterMode == FilterModeEditing {
			m.textInput, cmd = m.textInput.Update(msg)
			cmds = append(cmds, cmd)
			m.updateMatchCounts()
		}
	}

	if m.filterMode != FilterModeEditing {
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
	m.totalItems = len(items)
	m.updateMatchCounts()
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
	case FilterModeOff:
		return "No filter"
	case FilterModeEditing:
		if m.textInput.Value() == "" {
			return m.textInput.View() + fmt.Sprintf(" (%d / %d items match)", m.matchingItems, m.totalItems)
		}
		return m.textInput.View() + fmt.Sprintf(" (%d / %d items match)", m.matchingItems, m.totalItems)
	case FilterModeApplied:
		if m.textInput.Value() == "" {
			return "No filter"
		}
		return m.textInput.View() + fmt.Sprintf(" (%d / %d items match)", m.matchingItems, m.totalItems)
	default:
		panic(fmt.Sprintf("invalid filter mode: %d", m.filterMode))
	}
}

func (m *Model[T]) updateMatchCounts() {
	// TODO LEO
}
