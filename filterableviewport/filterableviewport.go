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
	"github.com/robinovitch61/bubbleo/viewport/item"
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
type Option[T viewport.Object] func(*Model[T])

// WithKeyMap sets the key mapping for the viewport
func WithKeyMap[T viewport.Object](keyMap KeyMap) Option[T] {
	return func(m *Model[T]) {
		m.keyMap = keyMap
	}
}

// WithStyles sets the styles for the filterable viewport
func WithStyles[T viewport.Object](styles Styles) Option[T] {
	return func(m *Model[T]) {
		m.styles = styles
		m.filterTextInput.Cursor.Style = styles.CursorStyle
	}
}

// WithPrefixText sets the prefix text for the filter line
func WithPrefixText[T viewport.Object](prefix string) Option[T] {
	return func(m *Model[T]) {
		m.prefixText = prefix
	}
}

// WithEmptyText sets the text to display when the filter is empty
func WithEmptyText[T viewport.Object](whenEmpty string) Option[T] {
	return func(m *Model[T]) {
		m.emptyText = whenEmpty
	}
}

// WithMatchingItemsOnly sets whether to show only the matching items
func WithMatchingItemsOnly[T viewport.Object](matchingItemsOnly bool) Option[T] {
	return func(m *Model[T]) {
		m.matchingItemsOnly = matchingItemsOnly
	}
}

// WithCanToggleMatchingItemsOnly sets whether this viewport can toggle matching items only mode
func WithCanToggleMatchingItemsOnly[T viewport.Object](canToggleMatchingItemsOnly bool) Option[T] {
	return func(m *Model[T]) {
		m.canToggleMatchingItemsOnly = canToggleMatchingItemsOnly
	}
}

// Model is the state and logic for a filterable viewport
type Model[T viewport.Object] struct {
	Viewport *viewport.Model[T]

	height          int
	keyMap          KeyMap
	filterTextInput textinput.Model
	filterMode      filterMode
	prefixText      string
	emptyText       string
	objects         []T
	isRegexMode     bool
	styles          Styles

	matchingItemsOnly          bool
	canToggleMatchingItemsOnly bool
	allMatches                 []Match
	numMatchingItems           int
	focusedMatchIdx            int
	previousFocusedMatchIdx    int
	totalMatchesOnAllItems     int
	itemIdxToFilteredIdx       map[int]int
}

// New creates a new filterable viewport model with default configuration
func New[T viewport.Object](vp *viewport.Model[T], opts ...Option[T]) *Model[T] {
	ti := textinput.New()
	ti.CharLimit = 0
	ti.Prompt = ""

	defaultKeyMap := DefaultKeyMap()
	defaultStyles := DefaultStyles()

	m := &Model[T]{
		Viewport:                   vp,
		height:                     0, // set below in SetHeight
		keyMap:                     defaultKeyMap,
		filterTextInput:            ti,
		filterMode:                 filterModeOff,
		prefixText:                 "",
		emptyText:                  "No Filter",
		objects:                    []T{},
		isRegexMode:                false,
		styles:                     defaultStyles,
		matchingItemsOnly:          false,
		canToggleMatchingItemsOnly: true,
		allMatches:                 []Match{},
		numMatchingItems:           0,
		focusedMatchIdx:            -1,
		previousFocusedMatchIdx:    -1,
		totalMatchesOnAllItems:     0,
		itemIdxToFilteredIdx:       make(map[int]int),
	}
	m.SetHeight(vp.GetHeight())

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
		case key.Matches(msg, m.keyMap.ToggleMatchingItemsOnlyKey):
			if m.filterMode != filterModeEditing && m.canToggleMatchingItemsOnly {
				m.matchingItemsOnly = !m.matchingItemsOnly
				m.updateMatchingItems()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.NextMatchKey):
			if m.filterMode != filterModeEditing && m.filterMode != filterModeOff && len(m.allMatches) > 0 {
				m.navigateToNextMatch()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.PrevMatchKey):
			if m.filterMode != filterModeEditing && m.filterMode != filterModeOff && len(m.allMatches) > 0 {
				m.navigateToPrevMatch()
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
	matchingItems := m.getMatchingObjectsAndUpdateMatches()
	m.ensureCurrentMatchInView()
	m.updateFocusedMatchHighlight()
	m.numMatchingItems = len(matchingItems)
	if m.matchingItemsOnly {
		m.Viewport.SetObjects(matchingItems)
	} else {
		m.Viewport.SetObjects(m.objects)
	}
}

// updateHighlighting updates the viewport's highlighting based on the filter
func (m *Model[T]) updateHighlighting() {
	filterText := m.filterTextInput.Value()
	if filterText == "" {
		m.Viewport.SetHighlights(nil)
		return
	}

	if m.isRegexMode {
		_, err := regexp.Compile(filterText)
		if err != nil {
			m.Viewport.SetHighlights(nil)
			return
		}
	}
	m.updateFocusedMatchHighlight()
}

// updateFocusedMatchHighlight sets a specific highlight for the currently focused match
func (m *Model[T]) updateFocusedMatchHighlight() {
	if m.focusedMatchIdx < 0 || m.focusedMatchIdx >= len(m.allMatches) {
		m.Viewport.SetHighlights(nil)
		return
	}

	// if only focus changed, update only the affected highlights
	if m.previousFocusedMatchIdx >= 0 && m.previousFocusedMatchIdx < len(m.allMatches) &&
		m.focusedMatchIdx != m.previousFocusedMatchIdx &&
		len(m.allMatches) > 0 {
		currentHighlights := m.Viewport.GetHighlights()
		if len(currentHighlights) == len(m.allMatches) {
			if m.previousFocusedMatchIdx < len(currentHighlights) {
				currentHighlights[m.previousFocusedMatchIdx].Style = m.styles.Match.Unfocused
			}
			if m.focusedMatchIdx < len(currentHighlights) {
				currentHighlights[m.focusedMatchIdx].Style = m.styles.Match.Focused
			}
			m.Viewport.SetHighlights(currentHighlights)
			m.previousFocusedMatchIdx = m.focusedMatchIdx
			return
		}
	}

	// otherwise, rebuild all highlights
	var highlights []item.Highlight
	for matchIdx, match := range m.allMatches {
		itemIdx := match.ItemIndex
		if m.matchingItemsOnly {
			if filteredIdx, ok := m.itemIdxToFilteredIdx[itemIdx]; ok {
				itemIdx = filteredIdx
			} else {
				panic("focused match item index not found in filtered items")
			}
		}
		style := m.styles.Match.Unfocused
		if matchIdx == m.focusedMatchIdx {
			style = m.styles.Match.Focused
		}
		highlight := item.Highlight{
			ItemIndex:       itemIdx,
			StartByteOffset: match.Start,
			EndByteOffset:   match.End,
			Style:           style,
		}
		highlights = append(highlights, highlight)
	}

	m.Viewport.SetHighlights(highlights)
	m.previousFocusedMatchIdx = m.focusedMatchIdx
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

// SetObjects sets the viewport objects
func (m *Model[T]) SetObjects(objects []T) {
	if objects == nil {
		objects = []T{}
	}
	m.Viewport.SetObjects(objects)
	m.objects = objects
	m.updateMatchingItems()
}

// SetWidth updates the width of both the viewport and textinput
func (m *Model[T]) SetWidth(width int) {
	m.Viewport.SetWidth(width)
}

// SetHeight updates the height, accounting for the filter line
func (m *Model[T]) SetHeight(height int) {
	m.height = height // TODO LEO: test this or remove height
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
		filterLine = m.emptyText
	case filterModeEditing, filterModeApplied:
		if m.filterTextInput.Value() == "" && m.filterMode == filterModeApplied {
			filterLine = m.emptyText
		} else {
			filterLine = strings.Join(removeEmpty([]string{
				m.getModeIndicator(),
				m.prefixText,
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
	filterItem := item.NewItem(filterLine)
	res, _ := filterItem.Take(0, m.GetWidth(), "...", []item.Highlight{})
	return res
}

func (m *Model[T]) getModeIndicator() string {
	if m.isRegexMode {
		return "[regex]"
	}
	return "[exact]"
}

// getMatchingObjectsAndUpdateMatches filters objects and updates match tracking
func (m *Model[T]) getMatchingObjectsAndUpdateMatches() []T {
	prevFocusedMatchIdx := m.focusedMatchIdx

	m.allMatches = []Match{}
	m.focusedMatchIdx = -1
	m.totalMatchesOnAllItems = 0
	m.itemIdxToFilteredIdx = make(map[int]int)

	filterValue := m.filterTextInput.Value()
	if m.filterMode == filterModeOff || filterValue == "" {
		return m.objects
	}

	filteredObjects := make([]T, 0, len(m.objects))
	if m.isRegexMode {
		regex, err := regexp.Compile(filterValue)
		if err != nil {
			return []T{}
		}
		for i := range m.objects {
			contentNoAnsi := m.objects[i].GetItem().ContentNoAnsi()
			matches := regex.FindAllStringIndex(contentNoAnsi, -1)
			if len(matches) > 0 {
				filteredObjects = append(filteredObjects, m.objects[i])
				m.itemIdxToFilteredIdx[i] = len(filteredObjects) - 1
				for _, match := range matches {
					m.allMatches = append(m.allMatches, Match{
						ItemIndex: i,
						Start:     match[0],
						End:       match[1],
					})
				}
			}
		}
	} else {
		for i := range m.objects {
			contentNoAnsi := m.objects[i].GetItem().ContentNoAnsi()
			start := 0
			hasMatch := false
			for {
				index := strings.Index(contentNoAnsi[start:], filterValue)
				if index == -1 {
					break
				}
				if !hasMatch {
					filteredObjects = append(filteredObjects, m.objects[i])
					hasMatch = true
					m.itemIdxToFilteredIdx[i] = len(filteredObjects) - 1
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

	if prevFocusedMatchIdx >= 0 && prevFocusedMatchIdx < len(m.allMatches) {
		m.focusedMatchIdx = prevFocusedMatchIdx
	} else if m.totalMatchesOnAllItems > 0 {
		m.focusedMatchIdx = 0
	} else {
		m.focusedMatchIdx = -1
	}

	return filteredObjects
}

// matchingItemsOnlyText returns the text to display when showing matching items only
func matchingItemsOnlyText(matchingItemsOnly bool) string {
	if matchingItemsOnly {
		return "showing matches only"
	}
	return ""
}

// removeEmpty removes empty strings from a slice
func removeEmpty(s []string) []string {
	var result []string
	for _, str := range s {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
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
	currentMatch := m.focusedMatchIdx + 1
	if m.focusedMatchIdx < 0 {
		currentMatch = 0
	}
	return fmt.Sprintf("(%d/%d matches on %d items)", currentMatch, m.totalMatchesOnAllItems, m.numMatchingItems)
}

func (m *Model[T]) navigateToNextMatch() {
	if len(m.allMatches) == 0 {
		return
	}

	m.focusedMatchIdx = (m.focusedMatchIdx + 1) % len(m.allMatches)
	m.ensureCurrentMatchInView()
	m.updateFocusedMatchHighlight()
}

func (m *Model[T]) navigateToPrevMatch() {
	if len(m.allMatches) == 0 {
		return
	}

	m.focusedMatchIdx--
	if m.focusedMatchIdx < 0 {
		m.focusedMatchIdx = len(m.allMatches) - 1
	}
	m.ensureCurrentMatchInView()
	m.updateFocusedMatchHighlight()
}

func (m *Model[T]) ensureCurrentMatchInView() {
	if m.focusedMatchIdx < 0 || m.focusedMatchIdx >= len(m.allMatches) {
		return
	}

	currentMatch := m.allMatches[m.focusedMatchIdx]
	m.Viewport.ScrollSoItemIdxInView(currentMatch.ItemIndex)
	if m.Viewport.GetSelectionEnabled() {
		m.Viewport.SetSelectedItemIdx(currentMatch.ItemIndex)
	}

	if !m.Viewport.GetWrapText() {
		m.panToCurrentMatch(currentMatch)
	}
}

func (m *Model[T]) panToCurrentMatch(match Match) {
	// TODO LEO: use widths, not byte offsets here
	matchCenter := match.Start + (match.End-match.Start)/2
	viewportWidth := m.Viewport.GetWidth()
	centeredXOffset := matchCenter - viewportWidth/2
	m.Viewport.SetXOffsetWidth(centeredXOffset)
}
