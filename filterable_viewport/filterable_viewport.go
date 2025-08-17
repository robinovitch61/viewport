package filterable_viewport

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

func WithText[T viewport.Renderable](prefix, whenEmpty string) Option[T] {
	return func(m *Model[T]) {
		m.text = textState{
			val:       prefix,
			whenEmpty: whenEmpty,
		}
	}
}

type Model[T viewport.Renderable] struct {
	Viewport *viewport.Model[T]

	keyMap        KeyMap
	textInput     textinput.Model
	filterMode    filterMode
	text          textState
	items         []T
	matchingItems int
	isRegexMode   bool
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
		Viewport:   vp,
		keyMap:     defaultKeyMap,
		filterMode: filterModeOff,
		text:       textState{whenEmpty: "No Filter"},
		textInput:  ti,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}

	return m
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
				m.isRegexMode = false
				m.textInput.Focus()
				m.filterMode = filterModeEditing
				m.updateMatchingItems()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.RegexFilterKey):
			if m.filterMode != filterModeEditing {
				m.isRegexMode = true
				m.textInput.Focus()
				m.filterMode = filterModeEditing
				m.updateMatchingItems()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.ApplyFilterKey):
			if m.filterMode == filterModeEditing {
				m.textInput.Blur()
				m.filterMode = filterModeApplied
				m.updateMatchingItems()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.CancelFilterKey):
			m.filterMode = filterModeOff
			m.isRegexMode = false
			m.textInput.Blur()
			m.textInput.SetValue("")
			m.updateMatchingItems()
			m.updateHighlighting()
			return m, nil
		}
	}

	if m.filterMode != filterModeEditing {
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
		m.updateMatchingItems()
		m.updateHighlighting()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model[T]) View() string {
	filterLine := m.renderFilterLine()
	viewportView := m.Viewport.View()
	return lipgloss.JoinVertical(lipgloss.Left, filterLine, viewportView)
}

// updateMatchingItems recalculates the matching items count
func (m *Model[T]) updateMatchingItems() {
	m.matchingItems = matchingItems(m.filterMode, m.items, m.textInput.Value(), m.isRegexMode)
}

// updateHighlighting updates the viewport's highlighting based on the filter
func (m *Model[T]) updateHighlighting() {
	filterText := m.textInput.Value()
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
	return m.textInput.Focused()
}

func (m *Model[T]) renderFilterLine() string {
	switch m.filterMode {
	case filterModeOff:
		return m.text.whenEmpty
	case filterModeEditing, filterModeApplied:
		if m.textInput.Value() == "" {
			if m.filterMode == filterModeApplied {
				return m.text.whenEmpty
			}
		}
		return strings.Join(removeEmpty([]string{
			m.getModeIndicator(),
			m.text.val,
			m.textInput.View(),
			matchCountText(m.matchingItems, len(m.items)),
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

func matchingItems[T viewport.Renderable](
	mode filterMode,
	items []T,
	filter string,
	isRegex bool,
) int {
	if mode == filterModeOff || filter == "" {
		return len(items)
	}

	count := 0
	if isRegex {
		regex, err := regexp.Compile(filter)
		if err != nil {
			return 0
		}
		for i := range items {
			if regex.MatchString(items[i].Render().Content()) {
				count++
			}
		}
	} else {
		for i := range items {
			if strings.Contains(items[i].Render().Content(), filter) {
				count++
			}
		}
	}
	return count
}

func matchCountText(matching, total int) string {
	if matching == 0 {
		return "(no matches)"
	}
	return fmt.Sprintf("(%d/%d matches)", matching, total)
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
