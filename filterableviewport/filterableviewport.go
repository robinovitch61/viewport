package filterableviewport

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/item"
)

type filterMode int

const (
	filterModeOff filterMode = iota
	filterModeEditing
	filterModeApplied
)

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

// WithVerticalPad sets the number of lines of context to keep above/below the focused match (scrolloff)
func WithVerticalPad[T viewport.Object](verticalPad int) Option[T] {
	return func(m *Model[T]) {
		m.verticalPad = verticalPad
	}
}

// WithHorizontalPad sets the number of columns of context to keep left/right of the focused match (panoff)
func WithHorizontalPad[T viewport.Object](horizontalPad int) Option[T] {
	return func(m *Model[T]) {
		m.horizontalPad = horizontalPad
	}
}

// WithMaxMatchLimit sets the maximum number of matches when searching.
// When this limit is exceeded, match highlighting and navigation are disabled
// and all items are shown regardless of matchingItemsOnly setting.
// Set to 0 for unlimited matches. Default is 30000.
func WithMaxMatchLimit[T viewport.Object](maxMatchLimit int) Option[T] {
	return func(m *Model[T]) {
		m.maxMatchLimit = maxMatchLimit
	}
}

// Model is the state and logic for a filterable viewport
type Model[T viewport.Object] struct {
	vp *viewport.Model[T]

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
	allMatches                 []viewport.Highlight
	numMatchingItems           int
	focusedMatchIdx            int
	previousFocusedMatchIdx    int
	totalMatchesOnAllItems     int
	itemIdxToFilteredIdx       map[int]int
	matchWidthsByMatchIdx      map[int]item.WidthRange
	lastFilterValue            string
	maxMatchLimit              int // 0 = unlimited
	matchLimitExceeded         bool

	verticalPad   int
	horizontalPad int
}

// New creates a new filterable viewport model with default configuration
func New[T viewport.Object](vp *viewport.Model[T], opts ...Option[T]) *Model[T] {
	ti := textinput.New()
	ti.CharLimit = 0
	ti.Prompt = ""

	defaultKeyMap := DefaultKeyMap()
	defaultStyles := DefaultStyles()

	m := &Model[T]{
		vp:                         vp,
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
		allMatches:                 []viewport.Highlight{},
		numMatchingItems:           0,
		focusedMatchIdx:            -1,
		previousFocusedMatchIdx:    -1,
		totalMatchesOnAllItems:     0,
		itemIdxToFilteredIdx:       make(map[int]int),
		matchWidthsByMatchIdx:      make(map[int]item.WidthRange),
		lastFilterValue:            "",
		maxMatchLimit:              30000, // reasonable default
		matchLimitExceeded:         false,
		verticalPad:                0,
		horizontalPad:              0,
	}
	m.SetHeight(vp.GetHeight())

	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}

	// set initial pre-footer line
	m.vp.SetPreFooterLine(m.renderFilterLine())

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

	if m.vp.IsCapturingInput() {
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.FilterKey):
			if m.filterMode != filterModeEditing {
				m.isRegexMode = false
				// remove (?i) prefix when switching to non-regex mode
				if newValue, found := strings.CutPrefix(m.filterTextInput.Value(), "(?i)"); found {
					m.filterTextInput.SetValue(newValue)
					m.filterTextInput.SetCursor(len(newValue))
				}
				m.filterTextInput.Focus()
				m.filterMode = filterModeEditing
				m.updateMatchingItems()
				m.ensureCurrentMatchInView()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.RegexFilterKey):
			if m.filterMode != filterModeEditing {
				m.isRegexMode = true
				m.filterTextInput.Focus()
				m.filterMode = filterModeEditing
				m.updateMatchingItems()
				m.ensureCurrentMatchInView()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.CaseInsensitiveFilterKey):
			if m.filterMode != filterModeEditing {
				m.isRegexMode = true
				currentValue := m.filterTextInput.Value()
				if currentValue == "" {
					m.filterTextInput.SetValue("(?i)")
					m.filterTextInput.SetCursor(4)
				} else if !strings.HasPrefix(currentValue, "(?i)") {
					// add the (?i) prefix if not already present when toggling case-insensitive mode
					newValue := "(?i)" + currentValue
					m.filterTextInput.SetValue(newValue)
					m.filterTextInput.SetCursor(len(newValue))
				}
				// already has (?i) prefix
				m.filterTextInput.Focus()
				m.filterMode = filterModeEditing
				m.updateMatchingItems()
				m.ensureCurrentMatchInView()
				return m, textinput.Blink
			}
		case key.Matches(msg, m.keyMap.ApplyFilterKey):
			if m.filterMode == filterModeEditing {
				m.filterTextInput.Blur()
				m.filterMode = filterModeApplied
				m.updateMatchingItems()
				m.ensureCurrentMatchInView()
				return m, nil
			}
		case key.Matches(msg, m.keyMap.ToggleMatchingItemsOnlyKey):
			if m.filterMode != filterModeEditing && m.canToggleMatchingItemsOnly {
				m.matchingItemsOnly = !m.matchingItemsOnly
				m.updateMatchingItems()
				m.ensureCurrentMatchInView()
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
			m.updateMatchingItems()
			m.ensureCurrentMatchInView()
			m.updateHighlighting()
			return m, nil
		}
	}

	if m.filterMode != filterModeEditing {
		m.vp, cmd = m.vp.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.filterTextInput, cmd = m.filterTextInput.Update(msg)
		m.updateMatchingItems()
		m.ensureCurrentMatchInView()
		m.updateHighlighting()
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the filterable viewport model as a string
func (m *Model[T]) View() string {
	return m.vp.View()
}

// GetWidth returns the width of the filterable viewport
func (m *Model[T]) GetWidth() int {
	return m.vp.GetWidth()
}

// SetWidth updates the width of both the viewport and textinput
func (m *Model[T]) SetWidth(width int) {
	m.vp.SetWidth(width)
}

// GetHeight returns the height of the filterable viewport
func (m *Model[T]) GetHeight() int {
	return m.vp.GetHeight()
}

// SetHeight updates the height of the filterable viewport
func (m *Model[T]) SetHeight(height int) {
	m.vp.SetHeight(height)
}

// SetObjects sets the viewport objects
func (m *Model[T]) SetObjects(objects []T) {
	if objects == nil {
		objects = []T{}
	}
	m.objects = objects
	m.updateMatchingItems()
}

// AppendObjects appends objects to the viewport's existing objects
func (m *Model[T]) AppendObjects(objects []T) {
	if objects == nil {
		return
	}
	startIdx := len(m.objects)
	m.objects = append(m.objects, objects...)

	// if filter active and not at limit, do incremental update
	if m.filterMode != filterModeOff &&
		m.filterTextInput.Value() != "" &&
		!m.matchLimitExceeded {
		m.appendMatchesForNewObjects(startIdx, objects)
	} else if m.matchLimitExceeded {
		// already at limit, just update viewport with all objects
		m.vp.SetObjects(m.objects)
	} else {
		m.updateMatchingItems()
	}
}

// FilterFocused returns true if the filter text input is focused
func (m *Model[T]) FilterFocused() bool {
	return m.filterTextInput.Focused()
}

// IsCapturingInput returns true when the filterableviewport or its underlying
// viewport is capturing input (e.g., filter entry, filename entry). Callers
// should check this before processing their own key bindings.
func (m *Model[T]) IsCapturingInput() bool {
	return m.filterTextInput.Focused() || m.vp.IsCapturingInput()
}

// GetWrapText returns whether text wrapping is enabled in the viewport
func (m *Model[T]) GetWrapText() bool {
	return m.vp.GetWrapText()
}

// SetWrapText sets whether text wrapping is enabled in the viewport
func (m *Model[T]) SetWrapText(wrapText bool) {
	m.vp.SetWrapText(wrapText)
}

// GetSelectionEnabled returns whether selection is enabled in the viewport
func (m *Model[T]) GetSelectionEnabled() bool {
	return m.vp.GetSelectionEnabled()
}

// SetSelectionEnabled sets whether selection is enabled in the viewport
func (m *Model[T]) SetSelectionEnabled(selectionEnabled bool) {
	m.vp.SetSelectionEnabled(selectionEnabled)
}

// updateMatchingItems recalculates the matching items and updates match tracking
func (m *Model[T]) updateMatchingItems() {
	matchingObjects := m.getMatchingObjectsAndUpdateMatches()
	m.updateFocusedMatchHighlight()

	if !m.matchLimitExceeded {
		m.numMatchingItems = len(matchingObjects)
	}

	// when match limit exceeded, show all objects
	if m.showMatchesOnly() {
		m.vp.SetObjects(matchingObjects)
	} else {
		m.vp.SetObjects(m.objects)
	}

	// update the pre-footer line with the current filter state
	m.vp.SetPreFooterLine(m.renderFilterLine())
}

// updateHighlighting updates the viewport's highlighting based on the filter
func (m *Model[T]) updateHighlighting() {
	filterText := m.filterTextInput.Value()
	if filterText == "" {
		m.vp.SetHighlights(nil)
		return
	}

	if m.isRegexMode {
		_, err := regexp.Compile(filterText)
		if err != nil {
			m.vp.SetHighlights(nil)
			return
		}
	}
	m.updateFocusedMatchHighlight()
}

// updateFocusedMatchHighlight sets a specific highlight for the currently focused match
func (m *Model[T]) updateFocusedMatchHighlight() {
	if m.focusedMatchIdx < 0 || m.focusedMatchIdx >= len(m.allMatches) {
		m.vp.SetHighlights(nil)
		return
	}

	// if only focus changed, update only the affected highlights
	if m.previousFocusedMatchIdx >= 0 && m.previousFocusedMatchIdx < len(m.allMatches) &&
		m.focusedMatchIdx != m.previousFocusedMatchIdx &&
		len(m.allMatches) > 0 {
		currentHighlights := m.vp.GetHighlights()
		if len(currentHighlights) == len(m.allMatches) {
			if m.previousFocusedMatchIdx < len(currentHighlights) {
				currentHighlights[m.previousFocusedMatchIdx].ItemHighlight.Style = m.styles.Match.Unfocused
			}
			if m.focusedMatchIdx < len(currentHighlights) {
				currentHighlights[m.focusedMatchIdx].ItemHighlight.Style = m.styles.Match.Focused
			}
			m.vp.SetHighlights(currentHighlights)
			m.previousFocusedMatchIdx = m.focusedMatchIdx
			return
		}
	}

	// otherwise, rebuild all highlights
	highlights := make([]viewport.Highlight, len(m.allMatches))
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
		highlight := viewport.Highlight{
			ItemIndex: itemIdx,
			ItemHighlight: item.Highlight{
				Style:                    style,
				ByteRangeUnstyledContent: match.ItemHighlight.ByteRangeUnstyledContent,
			},
		}
		highlights[matchIdx] = highlight
	}

	m.vp.SetHighlights(highlights)
	m.previousFocusedMatchIdx = m.focusedMatchIdx
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
				matchingItemsOnlyText(m.showMatchesOnly()),
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
	filterValue := m.filterTextInput.Value()
	filterChanged := filterValue != m.lastFilterValue
	m.lastFilterValue = filterValue

	m.allMatches = []viewport.Highlight{}
	prevFocusedMatchIdx := m.focusedMatchIdx
	m.focusedMatchIdx = -1
	m.totalMatchesOnAllItems = 0
	m.itemIdxToFilteredIdx = make(map[int]int)
	m.matchLimitExceeded = false

	if m.filterMode == filterModeOff || filterValue == "" {
		return m.objects
	}

	contentNoAnsiStrings := make([]string, len(m.objects))
	for i := range m.objects {
		contentNoAnsiStrings[i] = m.objects[i].GetItem().ContentNoAnsi()
	}

	var highlights []viewport.Highlight
	var regex *regexp.Regexp
	var err error
	if m.isRegexMode {
		regex, err = regexp.Compile(filterValue)
		if err != nil {
			return []T{}
		}
	}

	matchIdx := 0
	totalMatchCount := 0
	maxReached := false
	itemsWithMatchesSet := make(map[int]bool)

	for itemIdx := range contentNoAnsiStrings {
		matches := m.extractMatches(m.objects[itemIdx], filterValue, regex)

		if len(matches) > 0 {
			itemsWithMatchesSet[itemIdx] = true
		}

		if m.maxMatchLimit > 0 && totalMatchCount+len(matches) > m.maxMatchLimit {
			maxReached = true
			break
		}

		totalMatchCount += len(matches)

		newHighlights := m.buildHighlightsFromMatches(itemIdx, matches, matchIdx)
		matchIdx += len(matches)
		highlights = append(highlights, newHighlights...)
	}

	m.matchLimitExceeded = maxReached

	if maxReached {
		// clear match state and return all objects - no highlighting or navigation when limit exceeded
		m.allMatches = []viewport.Highlight{}
		m.focusedMatchIdx = -1
		m.totalMatchesOnAllItems = totalMatchCount
		// count of items with matches up to the limit
		m.numMatchingItems = len(itemsWithMatchesSet)
		return m.objects
	}

	filteredObjects := make([]T, 0, len(m.objects))
	itemsWithMatches := make(map[int]bool)

	for _, highlight := range highlights {
		itemIdx := highlight.ItemIndex
		if !itemsWithMatches[itemIdx] {
			filteredObjects = append(filteredObjects, m.objects[itemIdx])
			m.itemIdxToFilteredIdx[itemIdx] = len(filteredObjects) - 1
			itemsWithMatches[itemIdx] = true
		}
		m.allMatches = append(m.allMatches, highlight)
	}

	m.totalMatchesOnAllItems = len(m.allMatches)

	if filterChanged {
		if m.totalMatchesOnAllItems > 0 {
			m.focusedMatchIdx = 0
		} else {
			m.focusedMatchIdx = -1
		}
	} else {
		if prevFocusedMatchIdx >= 0 && prevFocusedMatchIdx < len(m.allMatches) {
			m.focusedMatchIdx = prevFocusedMatchIdx
		} else if m.totalMatchesOnAllItems > 0 {
			m.focusedMatchIdx = 0
		} else {
			m.focusedMatchIdx = -1
		}
	}

	return filteredObjects
}

// appendMatchesForNewObjects processes only newly appended objects for matches
// and incrementally updates match state without rescanning existing objects
func (m *Model[T]) appendMatchesForNewObjects(startIdx int, newObjects []T) {
	filterValue := m.filterTextInput.Value()

	var regex *regexp.Regexp
	var err error
	if m.isRegexMode {
		regex, err = regexp.Compile(filterValue)
		if err != nil {
			// invalid regex, fallback to full update
			m.updateMatchingItems()
			return
		}
	}

	matchIdx := len(m.allMatches)
	totalMatchCount := m.totalMatchesOnAllItems
	prevNumMatchingItems := m.numMatchingItems
	itemsWithMatchesSet := make(map[int]bool)
	var newHighlights []viewport.Highlight

	for i, obj := range newObjects {
		itemIdx := startIdx + i
		matches := m.extractMatches(obj, filterValue, regex)

		if len(matches) > 0 {
			itemsWithMatchesSet[itemIdx] = true
		}

		if m.maxMatchLimit > 0 && totalMatchCount+len(matches) > m.maxMatchLimit {
			// transition to match limit exceeded
			m.matchLimitExceeded = true
			m.allMatches = []viewport.Highlight{}
			m.focusedMatchIdx = -1
			m.totalMatchesOnAllItems = totalMatchCount
			m.numMatchingItems = prevNumMatchingItems + len(itemsWithMatchesSet)
			m.vp.SetObjects(m.objects)
			m.updateFocusedMatchHighlight()
			// update the pre-footer line with the current filter state
			m.vp.SetPreFooterLine(m.renderFilterLine())
			return
		}

		totalMatchCount += len(matches)

		highlights := m.buildHighlightsFromMatches(itemIdx, matches, matchIdx)
		matchIdx += len(matches)
		newHighlights = append(newHighlights, highlights...)
	}

	// append new matches to existing
	m.allMatches = append(m.allMatches, newHighlights...)
	m.totalMatchesOnAllItems = totalMatchCount
	m.numMatchingItems = prevNumMatchingItems + len(itemsWithMatchesSet)

	// update viewport objects
	if m.showMatchesOnly() {
		// build filtered objects list including new matching items
		filteredObjects := make([]T, 0, m.numMatchingItems)
		itemsWithMatches := make(map[int]bool)

		for _, highlight := range m.allMatches {
			itemIdx := highlight.ItemIndex
			if !itemsWithMatches[itemIdx] {
				filteredObjects = append(filteredObjects, m.objects[itemIdx])
				m.itemIdxToFilteredIdx[itemIdx] = len(filteredObjects) - 1
				itemsWithMatches[itemIdx] = true
			}
		}
		m.vp.SetObjects(filteredObjects)
	} else {
		// already updated by append to m.objects
		m.vp.SetObjects(m.objects)
	}

	m.updateFocusedMatchHighlight()
	// update the pre-footer line with the current filter state
	m.vp.SetPreFooterLine(m.renderFilterLine())
}

// extractMatches extracts matches from an object using the current filter settings
func (m *Model[T]) extractMatches(obj T, filterValue string, regex *regexp.Regexp) []item.Match {
	if m.isRegexMode && regex != nil {
		return obj.GetItem().ExtractRegexMatches(regex)
	}
	return obj.GetItem().ExtractExactMatches(filterValue)
}

// buildHighlightsFromMatches creates viewport highlights from item matches
func (m *Model[T]) buildHighlightsFromMatches(itemIdx int, matches []item.Match, startMatchIdx int) []viewport.Highlight {
	highlights := make([]viewport.Highlight, 0, len(matches))
	matchIdx := startMatchIdx

	for i := range matches {
		m.matchWidthsByMatchIdx[matchIdx] = matches[i].WidthRange
		matchIdx++

		highlight := viewport.Highlight{
			ItemIndex: itemIdx,
			ItemHighlight: item.Highlight{
				Style:                    m.styles.Match.Unfocused,
				ByteRangeUnstyledContent: matches[i].ByteRange,
			},
		}
		highlights = append(highlights, highlight)
	}

	return highlights
}

func (m *Model[T]) showMatchesOnly() bool {
	return m.matchingItemsOnly && !m.matchLimitExceeded
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
	if m.matchLimitExceeded {
		return fmt.Sprintf("(%d+ matches on %d+ items)", m.maxMatchLimit, m.numMatchingItems)
	}
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
	m.setSelectionToCurrentMatch()
	m.updateFocusedMatchHighlight()
	m.vp.SetPreFooterLine(m.renderFilterLine())
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
	m.setSelectionToCurrentMatch()
	m.updateFocusedMatchHighlight()
	m.vp.SetPreFooterLine(m.renderFilterLine())
}

func (m *Model[T]) getFocusedMatch() *viewport.Highlight {
	if m.focusedMatchIdx < 0 || m.focusedMatchIdx >= len(m.allMatches) {
		return nil
	}
	return &m.allMatches[m.focusedMatchIdx]
}

func (m *Model[T]) ensureCurrentMatchInView() {
	currentMatch := m.getFocusedMatch()
	if currentMatch == nil {
		return
	}
	widthRange := m.matchWidthsByMatchIdx[m.focusedMatchIdx]
	m.vp.EnsureItemInView(currentMatch.ItemIndex, widthRange.Start, widthRange.End, m.verticalPad, m.horizontalPad)
}

func (m *Model[T]) setSelectionToCurrentMatch() {
	currentMatch := m.getFocusedMatch()
	if currentMatch == nil {
		return
	}
	if m.vp.GetSelectionEnabled() && m.vp.GetSelectedItemIdx() != currentMatch.ItemIndex {
		m.vp.SetSelectedItemIdx(currentMatch.ItemIndex)
	}
}
