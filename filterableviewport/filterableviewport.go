package filterableviewport

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport"
)

type filterMode int

const (
	filterModeOff filterMode = iota
	filterModeEditing
	filterModeApplied
)

const (
	filterLineHeight = 1
)

// Option is a functional option for configuring the filterable viewport
type Option[T viewport.Renderable] func(*Model[T])

// WithKeyMap sets the key mapping for the viewport
func WithKeyMap[T viewport.Renderable](keyMap KeyMap) Option[T] {
	return func(m *Model[T]) {
		m.keyMap = keyMap
		m.Viewport.SetKeyMap(keyMap.ViewportKeyMap)
	}
}

// WithStyles sets the styling for the viewport
func WithStyles[T viewport.Renderable](styles viewport.Styles) Option[T] {
	return func(m *Model[T]) {
		m.Viewport.SetStyles(styles)
	}
}

type textState struct {
	val       string
	whenEmpty string
}

// WithText sets the text state (prefix and text when empty) for the filter line
func WithText[T viewport.Renderable](prefix, whenEmpty string) Option[T] {
	return func(m *Model[T]) {
		m.text = textState{
			val:       prefix,
			whenEmpty: whenEmpty,
		}
	}
}

// WithMatchesOnly sets whether to show only the matching items
func WithMatchesOnly[T viewport.Renderable](matchesOnly bool) Option[T] {
	return func(m *Model[T]) {
		m.matchesOnly = matchesOnly
	}
}

// WithCanToggleMatchesOnly sets whether this viewport can toggle matches only mode
func WithCanToggleMatchesOnly[T viewport.Renderable](canToggleMatchesOnly bool) Option[T] {
	return func(m *Model[T]) {
		m.canToggleMatchesOnly = canToggleMatchesOnly
	}
}

// Model is the state and logic for a filterable viewport
type Model[T viewport.Renderable] struct {
	Viewport *viewport.Model[T]

	keyMap               KeyMap
	filterTextInput      textinput.Model
	filterMode           filterMode
	text                 textState
	items                []T
	numMatchingItems     int
	isRegexMode          bool
	matchesOnly          bool
	canToggleMatchesOnly bool
}

// New creates a new filterable viewport model with default configuration
func New[T viewport.Renderable](width, height int, opts ...Option[T]) *Model[T] {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	ti := textinput.New()
	ti.CharLimit = 0
	ti.Prompt = ""

	defaultKeyMap := DefaultKeyMap()
	viewportHeight := height - filterLineHeight
	vp := viewport.New[T](width, viewportHeight,
		viewport.WithKeyMap[T](defaultKeyMap.ViewportKeyMap),
		viewport.WithStyles[T](viewport.DefaultStyles()),
	)

	m := &Model[T]{
		Viewport:             vp,
		keyMap:               defaultKeyMap,
		filterTextInput:      ti,
		filterMode:           filterModeOff,
		text:                 textState{whenEmpty: "No Filter"},
		items:                []T{},
		numMatchingItems:     0,
		isRegexMode:          false,
		matchesOnly:          false,
		canToggleMatchesOnly: true,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}

	return m
}

// Init initializes the filterable viewport model
func (m *Model[T]) Init() tea.Cmd {
	return nil
}

// Update processes messages and updates the model state
func (m *Model[T]) Update(msg tea.Msg) (*Model[T], tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.FilterKey):
			if m.filterMode != filterModeEditing {
				m.isRegexMode = false
				m.filterTextInput.Focus()
				m.filterMode = filterModeEditing
				m.updateMatchingItems()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.RegexFilterKey):
			if m.filterMode != filterModeEditing {
				m.isRegexMode = true
				m.filterTextInput.Focus()
				m.filterMode = filterModeEditing
				m.updateMatchingItems()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.ApplyFilterKey):
			if m.filterMode == filterModeEditing {
				m.filterTextInput.Blur()
				m.filterMode = filterModeApplied
				m.updateMatchingItems()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.CancelFilterKey):
			m.filterMode = filterModeOff
			m.isRegexMode = false
			m.filterTextInput.Blur()
			m.filterTextInput.SetValue("")
			m.updateHighlighting()
			m.updateMatchingItems()
			return m, nil
		case key.Matches(msg, m.keyMap.ToggleMatchesOnlyKey):
			if m.canToggleMatchesOnly {
				m.matchesOnly = !m.matchesOnly
				m.updateMatchingItems()
				return m, nil
			}
		}
	}

	if m.filterMode != filterModeEditing {
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.filterTextInput, cmd = m.filterTextInput.Update(msg)
		m.updateHighlighting()
		m.updateMatchingItems()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the filterable viewport model as a string
func (m *Model[T]) View() string {
	filterLine := m.renderFilterLine()
	viewportView := m.Viewport.View()
	return lipgloss.JoinVertical(lipgloss.Left, filterLine, viewportView)
}

// updateMatchingItems recalculates the matching items
func (m *Model[T]) updateMatchingItems() {
	matchingItems := m.getMatchingItems()
	m.numMatchingItems = len(matchingItems)
	if m.matchesOnly {
		m.Viewport.SetContent(matchingItems)
	} else {
		m.Viewport.SetContent(m.items)
	}
}

// updateHighlighting updates the viewport's highlighting based on the filter
func (m *Model[T]) updateHighlighting() {
	filterText := m.filterTextInput.Value()
	if filterText == "" {
		m.Viewport.SetStringToHighlight("")
		return
	}

	if m.isRegexMode {
		regex, err := regexp.Compile(filterText)
		if err != nil {
			m.Viewport.SetStringToHighlight("")
			return
		}
		m.Viewport.SetRegexToHighlight(regex)
	} else {
		m.Viewport.SetStringToHighlight(filterText)
	}
}

// SetContent sets the content and updates total item count
func (m *Model[T]) SetContent(items []T) {
	if items == nil {
		items = []T{}
	}
	m.Viewport.SetContent(items)
	m.items = items
	m.updateMatchingItems()
}

// SetWidth updates the width of both the viewport and textinput
func (m *Model[T]) SetWidth(width int) {
	m.Viewport.SetWidth(width)
}

// SetHeight updates the height, accounting for the filter line
func (m *Model[T]) SetHeight(height int) {
	m.Viewport.SetHeight(height - filterLineHeight)
}

// FilterFocused returns true if the filter text input is focused
func (m *Model[T]) FilterFocused() bool {
	return m.filterTextInput.Focused()
}

func (m *Model[T]) renderFilterLine() string {
	switch m.filterMode {
	case filterModeOff:
		return m.text.whenEmpty
	case filterModeEditing, filterModeApplied:
		if m.filterTextInput.Value() == "" {
			if m.filterMode == filterModeApplied {
				return m.text.whenEmpty
			}
		}
		return strings.Join(removeEmpty([]string{
			m.getModeIndicator(),
			m.text.val,
			m.filterTextInput.View(),
			matchCountText(m.numMatchingItems, len(m.items)),
			matchesOnlyText(m.matchesOnly),
		}),
			" ",
		)
	default:
		panic(fmt.Sprintf("invalid filter mode: %d", m.filterMode))
	}
}

func (m *Model[T]) getModeIndicator() string {
	if m.isRegexMode {
		return "[regex]"
	}
	return "[exact]"
}

func (m *Model[T]) getMatchingItems() []T {
	filterValue := m.filterTextInput.Value()
	if m.filterMode == filterModeOff || filterValue == "" {
		return m.items
	}

	filteredItems := make([]T, 0, len(m.items))
	if m.isRegexMode {
		regex, err := regexp.Compile(filterValue)
		if err != nil {
			return []T{}
		}
		for i := range m.items {
			if regex.MatchString(m.items[i].Render().Content()) {
				filteredItems = append(filteredItems, m.items[i])
			}
		}
	} else {
		for i := range m.items {
			if strings.Contains(m.items[i].Render().Content(), filterValue) {
				filteredItems = append(filteredItems, m.items[i])
			}
		}
	}
	return filteredItems
}

func matchCountText(matching, total int) string {
	if matching == 0 {
		return "(no matches)"
	}
	return fmt.Sprintf("(%d/%d matches)", matching, total)
}

func matchesOnlyText(matchesOnly bool) string {
	if matchesOnly {
		return "showing matches only"
	}
	return ""
}

func removeEmpty(s []string) []string {
	var result []string
	for _, str := range s {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}
