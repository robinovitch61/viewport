package filterableviewport

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/robinovitch61/bubbleo/viewport/linebuffer"

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

// Match represents a single match in the content
type Match struct {
	ItemIndex int // index of the item containing the match
	Start     int // start position of the match within the item's content
	End       int // end position of the match within the item's content
}

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

// WithMatchingItemsOnly sets whether to show only the matching items
func WithMatchingItemsOnly[T viewport.Renderable](matchingItemsOnly bool) Option[T] {
	return func(m *Model[T]) {
		m.matchingItemsOnly = matchingItemsOnly
	}
}

// WithCanToggleMatchingItemsOnly sets whether this viewport can toggle matching items only mode
func WithCanToggleMatchingItemsOnly[T viewport.Renderable](canToggleMatchingItemsOnly bool) Option[T] {
	return func(m *Model[T]) {
		m.canToggleMatchingItemsOnly = canToggleMatchingItemsOnly
	}
}

// Model is the state and logic for a filterable viewport
type Model[T viewport.Renderable] struct {
	Viewport *viewport.Model[T]

	height          int
	keyMap          KeyMap
	filterTextInput textinput.Model
	filterMode      filterMode
	text            textState
	items           []T
	isRegexMode     bool

	matchingItemsOnly          bool
	canToggleMatchingItemsOnly bool
	allMatches                 []Match
	numMatchingItems           int
	currentMatchIdx            int
	totalMatchesOnAllItems     int
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
	viewportHeight := max(0, height-filterLineHeight)
	vp := viewport.New[T](width, viewportHeight,
		viewport.WithKeyMap[T](defaultKeyMap.ViewportKeyMap),
		viewport.WithStyles[T](viewport.DefaultStyles()),
	)

	m := &Model[T]{
		Viewport:                   vp,
		height:                     height,
		keyMap:                     defaultKeyMap,
		filterTextInput:            ti,
		filterMode:                 filterModeOff,
		text:                       textState{whenEmpty: "No Filter"},
		items:                      []T{},
		isRegexMode:                false,
		matchingItemsOnly:          false,
		canToggleMatchingItemsOnly: true,
		allMatches:                 []Match{},
		numMatchingItems:           0,
		currentMatchIdx:            -1,
		totalMatchesOnAllItems:     0,
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
		case key.Matches(msg, m.keyMap.ToggleMatchingItemsOnlyKey):
			if m.canToggleMatchingItemsOnly {
				m.matchingItemsOnly = !m.matchingItemsOnly
				m.updateMatchingItems()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.NextMatchKey):
			if m.filterMode != filterModeOff && len(m.allMatches) > 0 {
				m.navigateToNextMatch()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.PrevMatchKey):
			if m.filterMode != filterModeOff && len(m.allMatches) > 0 {
				m.navigateToPrevMatch()
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
	if m.height <= 0 {
		return ""
	}
	filterLine := m.renderFilterLine()
	viewportView := m.Viewport.View()
	return lipgloss.JoinVertical(lipgloss.Left, filterLine, viewportView)
}

// updateMatchingItems recalculates the matching items and updates match tracking
func (m *Model[T]) updateMatchingItems() {
	matchingItems := m.getMatchingItems()
	m.numMatchingItems = len(matchingItems)
	m.updateMatches()
	if m.matchingItemsOnly {
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

// GetWidth returns the width of the filterable viewport
func (m *Model[T]) GetWidth() int {
	return m.Viewport.GetWidth()
}

// GetHeight returns the height of the filterable viewport
func (m *Model[T]) GetHeight() int {
	if m.height <= 0 {
		return 0
	}
	return m.Viewport.GetHeight() + filterLineHeight
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
	var filterLine string

	switch m.filterMode {
	case filterModeOff:
		filterLine = m.text.whenEmpty
	case filterModeEditing, filterModeApplied:
		if m.filterTextInput.Value() == "" && m.filterMode == filterModeApplied {
			filterLine = m.text.whenEmpty
		} else {
			filterLine = strings.Join(removeEmpty([]string{
				m.getModeIndicator(),
				m.text.val,
				m.filterTextInput.View(),
				m.getTextAfterFilter(),
				matchingItemsOnlyText(m.matchingItemsOnly),
			}),
				" ",
			)
		}
	default:
		panic(fmt.Sprintf("invalid filter mode: %d", m.filterMode))
	}
	filterLineBuffer := linebuffer.New(filterLine)
	res, _ := filterLineBuffer.Take(0, m.GetWidth(), "...", linebuffer.HighlightData{}, lipgloss.NewStyle())
	return res
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

func matchingItemsOnlyText(matchingItemsOnly bool) string {
	if matchingItemsOnly {
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

// updateMatches recalculates all matches and updates match tracking
func (m *Model[T]) updateMatches() {
	m.allMatches = []Match{}
	m.currentMatchIdx = -1
	m.totalMatchesOnAllItems = 0

	filterValue := m.filterTextInput.Value()
	if m.filterMode == filterModeOff || filterValue == "" {
		return
	}

	if m.isRegexMode {
		regex, err := regexp.Compile(filterValue)
		if err != nil {
			return
		}
		for i := range m.items {
			content := m.items[i].Render().Content()
			matches := regex.FindAllStringIndex(content, -1)
			for _, match := range matches {
				m.allMatches = append(m.allMatches, Match{
					ItemIndex: i,
					Start:     match[0],
					End:       match[1],
				})
			}
		}
	} else {
		for i := range m.items {
			content := m.items[i].Render().Content()
			start := 0
			for {
				index := strings.Index(content[start:], filterValue)
				if index == -1 {
					break
				}
				matchStart := start + index
				matchEnd := matchStart + len(filterValue)
				m.allMatches = append(m.allMatches, Match{
					ItemIndex: i,
					Start:     matchStart,
					End:       matchEnd,
				})
				start = matchEnd
			}
		}
	}

	m.totalMatchesOnAllItems = len(m.allMatches)

	if m.totalMatchesOnAllItems > 0 {
		m.currentMatchIdx = 0
	}
}

// getTextAfterFilter returns the text to display after the filter input
func (m *Model[T]) getTextAfterFilter() string {
	if m.filterTextInput.Value() == "" {
		return "type to filter"
	}
	return m.getMatchCountText()
}

// getMatchCountText returns the formatted match count text
func (m *Model[T]) getMatchCountText() string {
	if m.totalMatchesOnAllItems == 0 {
		return "(no matches)"
	}
	currentMatch := m.currentMatchIdx + 1
	if m.currentMatchIdx < 0 {
		currentMatch = 0
	}
	return fmt.Sprintf("(%d/%d matches on %d items)", currentMatch, m.totalMatchesOnAllItems, m.numMatchingItems)
}

func (m *Model[T]) navigateToNextMatch() {
	if len(m.allMatches) == 0 {
		return
	}

	m.currentMatchIdx = (m.currentMatchIdx + 1) % len(m.allMatches)
	m.scrollToCurrentMatch()
}

func (m *Model[T]) navigateToPrevMatch() {
	if len(m.allMatches) == 0 {
		return
	}

	m.currentMatchIdx--
	if m.currentMatchIdx < 0 {
		m.currentMatchIdx = len(m.allMatches) - 1
	}
	m.scrollToCurrentMatch()
}

func (m *Model[T]) scrollToCurrentMatch() {
	if m.currentMatchIdx < 0 || m.currentMatchIdx >= len(m.allMatches) {
		return
	}

	currentMatch := m.allMatches[m.currentMatchIdx]
	m.Viewport.ScrollSoItemIdxInView(currentMatch.ItemIndex)
	if m.Viewport.GetSelectionEnabled() {
		m.Viewport.SetSelectedItemIdx(currentMatch.ItemIndex)
	}
}
