package viewport

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Terminology:
// - items: a selectable item in the viewport, rendered as one or more lines of text
// - line: a line of text on one row of terminal cells
// - visible: in the vertical sense, a line is visible if it is within the viewport
// - truncated: in the horizontal sense, a line is truncated if it is too long to fit in the viewport
//
// wrap disabled, wide enough viewport:
//                           item index      line index
// this is the first line    0               0
// this is the second line   1               1
//
// wrap disabled, overflows viewport width:
//                           item index      line index
// this is the first...      0               0
// this is the secon...      1               1
//
// wrap enabled:
//               item index      line index
// this is the   0               0
// first line    0               1
// this is the   1               2
// second line   1               3

var surroundingAnsiRegex = regexp.MustCompile(`(\x1b\[[0-9;]*m.*?\x1b\[0?m)`)

// CompareFn is a function type for comparing two items of type T.
type CompareFn[T any] func(a, b T) bool

// Option is a functional option for configuring the viewport
type Option[T Renderable] func(*Model[T])

// WithKeyMap sets the key mapping for the viewport
func WithKeyMap[T Renderable](keyMap KeyMap) Option[T] {
	return func(m *Model[T]) {
		m.navigation.KeyMap = keyMap
	}
}

// WithStyles sets the styling for the viewport
func WithStyles[T Renderable](styles Styles) Option[T] {
	return func(m *Model[T]) {
		m.display.styles = styles
	}
}

// WithWrapText sets whether the viewport wraps text
func WithWrapText[T Renderable](wrap bool) Option[T] {
	return func(m *Model[T]) {
		m.SetWrapText(wrap)
	}
}

// WithSelectionEnabled sets whether the viewport allows selection
func WithSelectionEnabled[T Renderable](enabled bool) Option[T] {
	return func(m *Model[T]) {
		m.SetSelectionEnabled(enabled)
	}
}

// WithFooterEnabled sets whether the viewport shows the footer
func WithFooterEnabled[T Renderable](enabled bool) Option[T] {
	return func(m *Model[T]) {
		m.SetFooterEnabled(enabled)
	}
}

// Model represents a viewport component
type Model[T Renderable] struct {
	// content manages the content and selection state
	content *contentManager[T]

	// display handles rendering
	display *displayManager

	// navigation manages keyboard input and navigation logic
	navigation *NavigationManager

	// config manages configuration options
	config *configuration
}

// New creates a new viewport model with reasonable defaults
func New[T Renderable](width, height int, opts ...Option[T]) (m *Model[T]) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	m = &Model[T]{}
	m.content = newContentManager[T]()
	m.display = newDisplayManager(width, height, DefaultStyles())
	m.navigation = NewNavigationManager(DefaultKeyMap())
	m.config = newConfiguration()

	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}

	return m
}

// Update processes messages and updates the model
func (m *Model[T]) Update(msg tea.Msg) (*Model[T], tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		navCtx := NavigationContext{
			WrapText:        m.config.wrapText,
			Dimensions:      m.display.bounds,
			NumContentLines: m.getNumContentLinesWithFooterVisible(),
			NumVisibleItems: m.getNumVisibleItems(),
		}
		navResult := m.navigation.ProcessKeyMsg(msg, navCtx)

		switch navResult.Action {
		case ActionUp:
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxUp(navResult.SelectionAmount)
			} else {
				m.scrollUp(navResult.ScrollAmount)
			}

		case ActionDown:
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxDown(navResult.SelectionAmount)
			} else {
				m.scrollDown(navResult.ScrollAmount)
			}

		case ActionLeft:
			if !m.config.wrapText {
				m.viewLeft(navResult.ScrollAmount)
			}

		case ActionRight:
			if !m.config.wrapText {
				m.viewRight(navResult.ScrollAmount)
			}

		case ActionHalfPageUp:
			m.scrollUp(navResult.ScrollAmount)
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxUp(navResult.SelectionAmount)
			}

		case ActionHalfPageDown:
			m.scrollDown(navResult.ScrollAmount)
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxDown(navResult.SelectionAmount)
			}

		case ActionPageUp:
			m.scrollUp(navResult.ScrollAmount)
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxUp(navResult.SelectionAmount)
			}

		case ActionPageDown:
			m.scrollDown(navResult.ScrollAmount)
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxDown(navResult.SelectionAmount)
			}

		case ActionTop:
			if m.navigation.SelectionEnabled {
				m.SetSelectedItemIdx(0)
			} else {
				m.display.topItemIdx = 0
				m.display.topItemLineOffset = 0
			}

		case ActionBottom:
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxDown(m.content.numItems())
			} else {
				maxItemIdx, maxTopLineOffset := m.maxItemIdxAndMaxTopLineOffset()
				m.safelySetTopItemIdxAndOffset(maxItemIdx, maxTopLineOffset)
			}

		default:
			// no-op on keypress that doesn't produce a selection action
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the viewport
func (m *Model[T]) View() string {
	var builder strings.Builder
	wrap := m.config.wrapText

	visibleHeaderLines := m.getVisibleHeaderLines()
	content := m.getVisibleContent()

	// pre-allocate capacity based on estimated size
	estimatedSize := (len(visibleHeaderLines) + len(content.lineBuffers) + 10) * (m.display.bounds.Width + 1)
	builder.Grow(estimatedSize)

	// header lines
	for i := range visibleHeaderLines {
		lineBuffer := NewLineBuffer(visibleHeaderLines[i])
		line, _ := lineBuffer.Take(0, m.display.bounds.Width, m.config.continuationIndicator, []Highlight{})
		builder.WriteString(line)
		builder.WriteByte('\n')
	}

	// content lines
	truncatedVisibleContentLines := make([]string, len(content.lineBuffers))
	currentItemIdxWidthToLeft := m.display.bounds.Width * m.display.topItemLineOffset
	for lbIdx := range content.lineBuffers {
		var truncated string
		if wrap {
			currentItemIdx := content.itemIndexes[lbIdx]
			var widthTaken int
			truncated, widthTaken = content.lineBuffers[lbIdx].Take(
				currentItemIdxWidthToLeft,
				m.display.bounds.Width,
				"",
				m.getHighlightsForItem(currentItemIdx),
			)
			if lbIdx+1 < len(content.lineBuffers) {
				nextItemIdx := content.itemIndexes[lbIdx+1]
				if nextItemIdx != currentItemIdx {
					currentItemIdxWidthToLeft = 0
				} else {
					currentItemIdxWidthToLeft += widthTaken
				}
			}
		} else {
			// if not wrapped, lineBuffers are not yet truncated or highlighted
			truncated, _ = content.lineBuffers[lbIdx].Take(
				m.display.xOffset,
				m.display.bounds.Width,
				m.config.continuationIndicator,
				m.getHighlightsForItem(content.itemIndexes[lbIdx]),
			)
		}

		truncatedIsSelection := m.navigation.SelectionEnabled && content.itemIndexes[lbIdx] == m.content.getSelectedIdx()
		if truncatedIsSelection {
			truncated = m.styleSelection(truncated)
		}

		pannedRight := m.display.xOffset > 0
		itemHasWidth := content.lineBuffers[lbIdx].Width() > 0
		pannedPastAllWidth := lipgloss.Width(truncated) == 0
		if !wrap && pannedRight && itemHasWidth && pannedPastAllWidth {
			// if panned right past where line ends, show continuation indicator
			lineBuffer := NewLineBuffer(m.getLineContinuationIndicator())
			truncated, _ = lineBuffer.Take(0, m.display.bounds.Width, "", []Highlight{})
			if truncatedIsSelection {
				truncated = m.styleSelection(truncated)
			}
		}

		if truncatedIsSelection && lipgloss.Width(truncated) == 0 {
			// ensure selection is visible even if line empty
			truncated = m.styleSelection(" ")
		}

		truncatedVisibleContentLines[lbIdx] = truncated
	}

	for i := range truncatedVisibleContentLines {
		builder.WriteString(truncatedVisibleContentLines[i])
		builder.WriteByte('\n')
	}

	nVisibleLines := len(content.lineBuffers)
	if content.showFooter {
		// pad so footer shows up at bottom
		padCount := max(0, m.getNumContentLinesWithFooterVisible()-nVisibleLines)
		for i := 0; i < padCount; i++ {
			builder.WriteByte('\n')
		}
		builder.WriteString(m.getTruncatedFooterLine(content))
	}

	return m.display.renderFinalView(strings.TrimSuffix(builder.String(), "\n"))
}

// SetKeyMap sets the key mapping for navigation controls.
func (m *Model[T]) SetKeyMap(keyMap KeyMap) {
	m.navigation.KeyMap = keyMap
}

// SetStyles sets the styling configuration for the viewport
func (m *Model[T]) SetStyles(styles Styles) {
	m.display.styles = styles
}

// SetContent sets the SingleItem, the selectable set of lines in the viewport
func (m *Model[T]) SetContent(content []T) {
	var initialNumLinesAboveSelection int
	var stayAtTop, stayAtBottom bool
	var prevSelection T
	if m.navigation.SelectionEnabled {
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			initialNumLinesAboveSelection = inView.numLinesAboveSelection
		}
		currentItems := m.content.items
		selectedIdx := m.content.getSelectedIdx()
		if m.navigation.TopSticky && len(currentItems) > 0 && selectedIdx == 0 {
			stayAtTop = true
		} else if m.navigation.BottomSticky && (len(currentItems) == 0 || (selectedIdx == len(currentItems)-1)) {
			stayAtBottom = true
		} else if m.content.compareFn != nil && 0 <= selectedIdx && selectedIdx < len(currentItems) {
			prevSelection = currentItems[selectedIdx]
		}
	}

	m.content.items = content
	// ensure scroll position is valid given new LineBuffer
	m.safelySetTopItemIdxAndOffset(m.display.topItemIdx, m.display.topItemLineOffset)

	// ensure xOffset is valid given new LineBuffer
	m.SetXOffsetWidth(m.display.xOffset)

	if m.navigation.SelectionEnabled {
		if stayAtTop {
			m.content.setSelectedIdx(0)
		} else if stayAtBottom {
			m.content.setSelectedIdx(max(0, m.content.numItems()-1))
			m.scrollSoSelectionInView()
		} else if m.content.compareFn != nil {
			// TODO: could flag when items are sorted & comparable and use binary search instead
			found := false
			items := m.content.items
			for i := range items {
				if m.content.compareFn(items[i], prevSelection) {
					m.content.setSelectedIdx(i)
					found = true
					break
				}
			}
			if !found {
				m.content.setSelectedIdx(0)
			}
		}

		// when staying at bottom, just want to scroll so selection in view, which is done above
		if !stayAtBottom {
			m.content.validateSelectedIdx()
			m.scrollSoSelectionInView()
			if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
				m.scrollUp(initialNumLinesAboveSelection - inView.numLinesAboveSelection)
			}
		}
	}
}

// SetTopSticky sets whether selection should stay at top when new SingleItem added and selection is at the top
func (m *Model[T]) SetTopSticky(topSticky bool) {
	m.navigation.TopSticky = topSticky
}

// SetBottomSticky sets whether selection should stay at bottom when new SingleItem added and selection is at the bottom
func (m *Model[T]) SetBottomSticky(bottomSticky bool) {
	m.navigation.BottomSticky = bottomSticky
}

// SetSelectionEnabled sets whether the viewport allows line selection
func (m *Model[T]) SetSelectionEnabled(selectionEnabled bool) {
	wasEnabled := m.navigation.SelectionEnabled
	m.navigation.SelectionEnabled = selectionEnabled

	// when enabling selection, set the selected item to the top visible item and ensure the top line is in view
	if selectionEnabled && !wasEnabled && !m.content.isEmpty() {
		topVisibleItemIdx := clampValZeroToMax(m.display.topItemIdx, m.content.numItems()-1)
		m.content.setSelectedIdx(topVisibleItemIdx)
		m.scrollSoSelectionInView()
	}
}

// SetFooterEnabled sets whether the viewport shows the footer when it overflows
func (m *Model[T]) SetFooterEnabled(footerEnabled bool) {
	m.config.footerEnabled = footerEnabled
}

// SetSelectionComparator sets the comparator function for maintaining the current selection when SingleItem changes.
// If compareFn is non-nil, the viewport will try to maintain the current selection when SingleItem changes.
func (m *Model[T]) SetSelectionComparator(compareFn CompareFn[T]) {
	m.content.compareFn = compareFn
}

// GetSelectionEnabled returns whether the viewport allows line selection
func (m *Model[T]) GetSelectionEnabled() bool {
	return m.navigation.SelectionEnabled
}

// SetWrapText sets whether the viewport wraps text
func (m *Model[T]) SetWrapText(wrapText bool) {
	var initialNumLinesAboveSelection int
	if m.navigation.SelectionEnabled {
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			initialNumLinesAboveSelection = inView.numLinesAboveSelection
		}
	}
	m.config.wrapText = wrapText
	m.display.topItemLineOffset = 0
	m.display.xOffset = 0
	if m.navigation.SelectionEnabled {
		m.scrollSoSelectionInView()
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			m.scrollUp(initialNumLinesAboveSelection - inView.numLinesAboveSelection)
			m.scrollSoSelectionInView()
		}
	}
	m.safelySetTopItemIdxAndOffset(m.display.topItemIdx, m.display.topItemLineOffset)
}

// GetWrapText returns whether the viewport wraps text
func (m *Model[T]) GetWrapText() bool {
	return m.config.wrapText
}

// SetWidth sets the viewport's width
func (m *Model[T]) SetWidth(width int) {
	m.setWidthHeight(width, m.display.bounds.Height)
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.setWidthHeight(m.display.bounds.Width, height)
}

// SetSelectedItemIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model[T]) SetSelectedItemIdx(selectedItemIdx int) {
	if !m.navigation.SelectionEnabled {
		return
	}
	m.content.setSelectedIdx(selectedItemIdx)
	m.scrollSoSelectionInView()
}

// GetSelectedItemIdx returns the currently selected item index
func (m *Model[T]) GetSelectedItemIdx() int {
	if !m.navigation.SelectionEnabled {
		return 0
	}
	return m.content.getSelectedIdx()
}

// GetSelectedItem returns a pointer to the currently selected item
func (m *Model[T]) GetSelectedItem() *T {
	if !m.navigation.SelectionEnabled {
		return nil
	}
	return m.content.getSelectedItem()
}

// SetHeader sets the header, an unselectable set of lines at the top of the viewport
func (m *Model[T]) SetHeader(header []string) {
	m.content.header = header
}

// GetWidth returns the viewport width
func (m *Model[T]) GetWidth() int {
	return m.display.bounds.Width
}

// GetHeight returns the viewport height
func (m *Model[T]) GetHeight() int {
	return m.display.bounds.Height
}

// ScrollSoItemIdxInView scrolls the viewport to ensure the specified item index is visible.
// If the desired item is above the current content, this scrolls so that
func (m *Model[T]) ScrollSoItemIdxInView(itemIdx int) {
	if m.content.isEmpty() {
		m.safelySetTopItemIdxAndOffset(0, 0)
		return
	}
	originalTopItemIdx, originalTopItemLineOffset := m.display.topItemIdx, m.display.topItemLineOffset

	content := m.getVisibleContent()
	numLinesInViewForItem := 0
	for i := range content.itemIndexes {
		if content.itemIndexes[i] == itemIdx {
			numLinesInViewForItem++
		}
	}

	numLinesInItem := m.numLinesForItem(itemIdx)
	if numLinesInItem != numLinesInViewForItem {
		priorTopItemIdx := m.display.topItemIdx

		// scroll so item is at the top of the content
		m.display.topItemIdx = itemIdx
		m.display.topItemLineOffset = 0

		if priorTopItemIdx < itemIdx {
			// if the desired visible item is below the content previously on screen,
			// scroll up so that item is at the bottom
			m.scrollUp(max(0, m.getNumContentLinesWithFooterVisible()-numLinesInItem))
		}
	}

	if m.navigation.SelectionEnabled {
		// if scrolled such that selection is now fully out of view, undo it
		if m.selectionInViewInfo().numLinesSelectionInView == 0 {
			m.display.topItemIdx = originalTopItemIdx
			m.display.topItemLineOffset = originalTopItemLineOffset
		}
	}
}

// SetXOffsetWidth sets the horizontal offset, in terminal cell width, for panning when text wrapping is disabled
func (m *Model[T]) SetXOffsetWidth(width int) {
	if m.config.wrapText {
		return
	}
	maxXOffset := m.maxItemWidth() - m.display.bounds.Width
	m.display.xOffset = max(0, min(maxXOffset, width))
}

// SetHighlights sets specific positions to highlight with custom styles in the viewport.
func (m *Model[T]) SetHighlights(highlights []Highlight) {
	m.content.setHighlights(highlights)
}

// GetHighlights returns all highlights.
func (m *Model[T]) GetHighlights() []Highlight {
	return m.content.getHighlights()
}

func (m *Model[T]) maxItemWidth() int {
	if m.config.wrapText {
		panic("maxItemWidth should not be called when wrapping is enabled")
	}

	maxLineWidth := 0

	headerLines := m.getVisibleHeaderLines()
	for i := range headerLines {
		if w := lipgloss.Width(headerLines[i]); w > maxLineWidth {
			maxLineWidth = w
		}
	}

	// check content line widths without fully rendering all of them
	if !m.content.isEmpty() {
		items := m.content.items
		startIdx := clampValZeroToMax(m.display.topItemIdx, m.content.numItems()-1)
		numItemsToCheck := min(m.content.numItems()-startIdx, m.display.bounds.Height)

		for i := 0; i < numItemsToCheck; i++ {
			itemIdx := startIdx + i
			if itemIdx >= m.content.numItems() {
				break
			}
			lb := items[itemIdx].Render()
			if w := lb.Width(); w > maxLineWidth {
				maxLineWidth = w
			}
		}
	}

	return maxLineWidth
}

func (m *Model[T]) numLinesForItem(itemIdx int) int {
	if !m.config.wrapText {
		return 1
	}
	if m.display.bounds.Width == 0 {
		return 0
	}
	if m.content.isEmpty() || itemIdx < 0 || itemIdx >= m.content.numItems() {
		return 0
	}
	items := m.content.items
	lb := items[itemIdx].Render()
	return lb.NumWrappedLines(m.display.bounds.Width)
}

func (m *Model[T]) setWidthHeight(width, height int) {
	m.display.setBounds(width, height)
	if m.navigation.SelectionEnabled {
		m.scrollSoSelectionInView()
	}
	m.safelySetTopItemIdxAndOffset(m.display.topItemIdx, m.display.topItemLineOffset)
}

func (m *Model[T]) safelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset int) {
	maxTopItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	m.display.safelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset, maxTopItemIdx, maxTopItemLineOffset)
}

// getNumContentLinesWithFooterVisible returns the number of lines of between the header and footer
func (m *Model[T]) getNumContentLinesWithFooterVisible() int {
	return m.display.getNumContentLines(len(m.getVisibleHeaderLines()), true)
}

func (m *Model[T]) scrollSoSelectionInView() {
	if !m.navigation.SelectionEnabled {
		panic("scrollSoSelectionInView called when selection is not enabled")
	}
	m.ScrollSoItemIdxInView(m.content.getSelectedIdx())
}

func (m *Model[T]) selectedItemIdxDown(n int) {
	m.SetSelectedItemIdx(m.content.getSelectedIdx() + n)
}

func (m *Model[T]) selectedItemIdxUp(n int) {
	m.SetSelectedItemIdx(m.content.getSelectedIdx() - n)
}

func (m *Model[T]) scrollDown(n int) {
	m.scrollByNLines(n)
}

func (m *Model[T]) scrollUp(n int) {
	m.scrollByNLines(-n)
}

func (m *Model[T]) viewLeft(n int) {
	m.SetXOffsetWidth(m.display.xOffset - n)
}

func (m *Model[T]) viewRight(n int) {
	m.SetXOffsetWidth(m.display.xOffset + n)
}

// getItemIdxAbove consumes n lines by moving up through items, returning the final item index and line offset
func (m *Model[T]) getItemIdxAbove(startItemIdx, startLineOffset, linesToConsume int) (finalItemIdx, finalLineOffset int) {
	itemIdx := startItemIdx
	lineOffset := startLineOffset
	remaining := linesToConsume

	for remaining > 0 {
		itemIdx--
		if itemIdx < 0 {
			return 0, 0
		}
		numLinesInItem := m.numLinesForItem(itemIdx)
		if remaining <= numLinesInItem {
			return itemIdx, numLinesInItem - remaining
		}
		remaining -= numLinesInItem
	}
	return itemIdx, lineOffset
}

// getItemIdxBelow consumes n lines by moving down through items, returning the final item index and line offset
func (m *Model[T]) getItemIdxBelow(startItemIdx, linesToConsume int) (finalItemIdx, finalLineOffset int) {
	itemIdx := startItemIdx
	remaining := linesToConsume

	for remaining > 0 {
		itemIdx++
		if itemIdx >= m.content.numItems() {
			return m.content.numItems() - 1, 0
		}
		numLinesInItem := m.numLinesForItem(itemIdx)
		if remaining <= numLinesInItem {
			return itemIdx, remaining - 1
		}
		remaining -= numLinesInItem
	}
	return itemIdx, 0
}

// scrollByNLines edits topItemIdx and topItemLineOffset to scroll the viewport by n lines (negative for up, positive for down)
func (m *Model[T]) scrollByNLines(n int) {
	if n == 0 {
		return
	}

	// scrolling down past bottom
	if n > 0 && m.isScrolledToBottom() {
		return
	}

	// scrolling up past top
	if n < 0 && m.display.topItemIdx == 0 && m.display.topItemLineOffset == 0 {
		return
	}

	newTopItemIdx, newTopItemLineOffset := m.display.topItemIdx, m.display.topItemLineOffset
	if !m.config.wrapText {
		newTopItemIdx = m.display.topItemIdx + n
	} else {
		// wrapped
		if n < 0 { // negative n, scrolling up
			if newTopItemLineOffset >= -n {
				// same item, just change offset
				newTopItemLineOffset += n
			} else {
				// need to scroll up through multiple items
				linesToConsume := -n - newTopItemLineOffset
				newTopItemIdx, newTopItemLineOffset = m.getItemIdxAbove(newTopItemIdx, newTopItemLineOffset, linesToConsume)
			}
		} else { // positive n, scrolling down
			numLinesInTopItem := m.numLinesForItem(newTopItemIdx)
			if newTopItemLineOffset+n < numLinesInTopItem {
				// same item, just change offset
				newTopItemLineOffset += n
			} else {
				// need to scroll down through multiple items
				linesToConsume := n - (numLinesInTopItem - (newTopItemLineOffset + 1))
				newTopItemIdx, newTopItemLineOffset = m.getItemIdxBelow(newTopItemIdx, linesToConsume)
			}
		}
	}
	m.safelySetTopItemIdxAndOffset(newTopItemIdx, newTopItemLineOffset)
	m.SetXOffsetWidth(m.display.xOffset)
}

// getVisibleHeaderLines returns the lines of header that are visible in the viewport as strings.
// Header lines will take precedence over content and footer if there is not enough vertical height
func (m *Model[T]) getVisibleHeaderLines() []string {
	if m.display.bounds.Height == 0 {
		return nil
	}

	headerLbs := make([]Item, len(m.content.header))
	for i := range m.content.header {
		headerLbs[i] = NewLineBuffer(m.content.header[i])
	}

	lbs := m.getLineBuffersForLines(
		0,
		0,
		m.display.bounds.Height,
		headerLbs,
	)

	headerLines := make([]string, len(lbs.lineBuffers))
	currentItemIdxWidthToLeft := 0
	for lbIdx := range lbs.lineBuffers {
		var truncated string
		if m.config.wrapText {
			currentItemIdx := lbs.itemIndexes[lbIdx]
			var widthTaken int
			truncated, widthTaken = lbs.lineBuffers[lbIdx].Take(
				currentItemIdxWidthToLeft,
				m.display.bounds.Width,
				"",
				[]Highlight{}, // no highlights for header
			)
			if lbIdx+1 < len(lbs.lineBuffers) {
				nextItemIdx := lbs.itemIndexes[lbIdx+1]
				if nextItemIdx != currentItemIdx {
					currentItemIdxWidthToLeft = 0
				} else {
					currentItemIdxWidthToLeft += widthTaken
				}
			}
		} else {
			// if not wrapped, lineBuffers are not yet truncated or highlighted
			truncated, _ = lbs.lineBuffers[lbIdx].Take(
				0, // header doesn't pan horizontally
				m.display.bounds.Width,
				m.config.continuationIndicator,
				[]Highlight{}, // no highlights for header
			)
		}
		headerLines[lbIdx] = truncated
	}

	return headerLines
}

type visibleContent struct {
	// lineBuffers contains the lineBuffers that have at least one line currently visible
	lineBuffers []Item
	// itemIndexes is the index of the item in allItems that corresponds to each  len(itemIndexes) == len(lineBuffers)
	itemIndexes []int
	// showFooter indicates if the footer is visible
	showFooter bool
}

// getVisibleContent returns the lineBuffers, associated item indexes, and whether to show the footer for the current
// scroll position
func (m *Model[T]) getVisibleContent() visibleContent {
	if m.display.bounds.Width == 0 {
		return visibleContent{lineBuffers: nil, itemIndexes: nil, showFooter: false}
	}
	if m.content.isEmpty() {
		return visibleContent{lineBuffers: nil, itemIndexes: nil, showFooter: false}
	}

	numLinesAfterHeader := max(0, m.display.bounds.Height-len(m.getVisibleHeaderLines()))

	lbs := m.getLineBuffersForLines(
		m.display.topItemIdx,
		m.display.topItemLineOffset,
		numLinesAfterHeader,
		renderAll(m.content.items),
	)
	if len(lbs.lineBuffers) == 0 {
		return visibleContent{lineBuffers: nil, itemIndexes: nil, showFooter: false}
	}

	scrolledToTop := m.display.topItemIdx == 0 && m.display.topItemLineOffset == 0
	contentFillsScreen := len(lbs.lineBuffers)+1 >= numLinesAfterHeader
	showFooter := m.config.footerEnabled && (!scrolledToTop || contentFillsScreen)
	if showFooter {
		// leave one line for the footer
		lbs.lineBuffers = safeSliceUpToIdx(lbs.lineBuffers, numLinesAfterHeader-1)
		lbs.itemIndexes = safeSliceUpToIdx(lbs.itemIndexes, numLinesAfterHeader-1)
	}
	return visibleContent{lineBuffers: lbs.lineBuffers, itemIndexes: lbs.itemIndexes, showFooter: showFooter}
}

func renderAll[T Renderable](items []T) []Item {
	lineBuffers := make([]Item, len(items))
	for i := range items {
		lineBuffers[i] = items[i].Render()
	}
	return lineBuffers
}

type lineBuffersAndItemIndexes struct {
	// lineBuffers contains the lineBuffers that have at least one line currently visible in the content
	lineBuffers []Item
	// itemIndexes is the index of the item that corresponds to each  len(itemIndexes) == len(lineBuffers)
	itemIndexes []int
}

// getLineBuffersForLines returns the lineBuffers and associated item indexes for the offset and num lines specified
func (m *Model[T]) getLineBuffersForLines(
	topItemIdx int,
	topItemLineOffset int,
	totalNumLines int,
	allLineBuffers []Item,
) lineBuffersAndItemIndexes {
	if len(allLineBuffers) == 0 || totalNumLines == 0 {
		return lineBuffersAndItemIndexes{lineBuffers: nil, itemIndexes: nil}
	}

	var lineBuffers []Item
	var itemIndexes []int

	addLine := func(l Item, itemIndex int) bool {
		lineBuffers = append(lineBuffers, l)
		itemIndexes = append(itemIndexes, itemIndex)
		return len(lineBuffers) == totalNumLines
	}

	currLineBufferIdx := clampValZeroToMax(topItemIdx, len(allLineBuffers)-1)

	currLineBuffer := allLineBuffers[currLineBufferIdx]
	done := totalNumLines == 0
	if done {
		return lineBuffersAndItemIndexes{lineBuffers: lineBuffers, itemIndexes: itemIndexes}
	}

	if m.config.wrapText {
		// first item has potentially fewer lines depending on the line offset
		numLines := max(0, currLineBuffer.NumWrappedLines(m.display.bounds.Width)-topItemLineOffset)
		for range numLines {
			// adding untruncated, unstyled lineBuffers
			done = addLine(currLineBuffer, currLineBufferIdx)
			if done {
				break
			}
		}

		for !done {
			currLineBufferIdx++
			if currLineBufferIdx >= len(allLineBuffers) {
				done = true
			} else {
				currLineBuffer = allLineBuffers[currLineBufferIdx]
				numLines = currLineBuffer.NumWrappedLines(m.display.bounds.Width)
				for range numLines {
					// adding untruncated, unstyled lineBuffers
					done = addLine(currLineBuffer, currLineBufferIdx)
					if done {
						break
					}
				}
			}
		}
	} else {
		done = addLine(currLineBuffer, currLineBufferIdx)
		for !done {
			currLineBufferIdx++
			if currLineBufferIdx >= len(allLineBuffers) {
				done = true
			} else {
				currLineBuffer = allLineBuffers[currLineBufferIdx]
				done = addLine(currLineBuffer, currLineBufferIdx)
			}
		}
	}
	return lineBuffersAndItemIndexes{lineBuffers: lineBuffers, itemIndexes: itemIndexes}
}

// TODO LEO: reuse this for selection styling
//func (m *Model[T]) highlightStyle(itemIdx int) lipgloss.Style {
//	return m.display.getHighlightStyle(m.navigation.SelectionEnabled && itemIdx == m.content.getSelectedIdx())
//}

func (m *Model[T]) getTruncatedFooterLine(visibleContentLines visibleContent) string {
	numerator := m.content.getSelectedIdx() + 1 // 0 indexed
	denominator := m.content.numItems()
	if !visibleContentLines.showFooter {
		panic("getTruncatedFooterLine called when footer should not be shown")
	}
	if len(visibleContentLines.lineBuffers) == 0 {
		return ""
	}

	// if selection is disabled, numerator should be item index of bottom visible line
	if !m.navigation.SelectionEnabled {
		numerator = visibleContentLines.itemIndexes[len(visibleContentLines.itemIndexes)-1] + 1
		if m.config.wrapText && numerator == denominator && !m.isScrolledToBottom() {
			// if wrapped && bottom visible line is max item index, but actually not fully scrolled to bottom, show 99%
			return m.display.styles.FooterStyle.Render(fmt.Sprintf("99%% (%d/%d)", numerator, denominator))
		}
	}

	percentScrolled := percent(numerator, denominator)
	footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)

	footerBuffer := NewLineBuffer(footerString)
	f, _ := footerBuffer.Take(0, m.display.bounds.Width, m.config.continuationIndicator, []Highlight{})
	return m.display.styles.FooterStyle.Render(f)
}

func (m *Model[T]) getLineContinuationIndicator() string {
	if m.config.wrapText {
		return ""
	}
	return m.config.continuationIndicator
}

func (m *Model[T]) isScrolledToBottom() bool {
	maxItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	if m.display.topItemIdx > maxItemIdx {
		return true
	}
	if m.display.topItemIdx == maxItemIdx {
		return m.display.topItemLineOffset >= maxTopItemLineOffset
	}
	return false
}

type selectionInViewInfoResult struct {
	numLinesSelectionInView int
	numLinesAboveSelection  int
}

func (m *Model[T]) selectionInViewInfo() selectionInViewInfoResult {
	if !m.navigation.SelectionEnabled {
		panic("selectionInViewInfo called when selection is disabled")
	}
	content := m.getVisibleContent()
	numLinesSelectionInView := 0
	numLinesAboveSelection := 0
	assignedNumLinesAboveSelection := false
	for i := range content.itemIndexes {
		if content.itemIndexes[i] == m.content.getSelectedIdx() {
			if !assignedNumLinesAboveSelection {
				numLinesAboveSelection = i
				assignedNumLinesAboveSelection = true
			}
			numLinesSelectionInView++
		}
	}
	return selectionInViewInfoResult{
		numLinesSelectionInView: numLinesSelectionInView,
		numLinesAboveSelection:  numLinesAboveSelection,
	}
}

func (m *Model[T]) maxItemIdxAndMaxTopLineOffset() (int, int) {
	numItems := m.content.numItems()
	if numItems == 0 {
		return 0, 0
	}

	headerLines := len(m.getVisibleHeaderLines())
	// assume footer will be shown - if it isn't, max item idx and offset will both be 0
	numContentLines := max(0, m.display.bounds.Height-headerLines-1)

	if !m.config.wrapText {
		return max(0, numItems-numContentLines), 0
	}

	// wrapped
	maxTopItemIdx, maxTopItemLineOffset := numItems-1, 0
	numLinesLastItem := m.numLinesForItem(numItems - 1)
	if numContentLines <= numLinesLastItem {
		// last item takes up whole screen or more, adjust offset accordingly
		maxTopItemLineOffset = numLinesLastItem - numContentLines
	} else {
		// need to scroll up through multiple items to fill the screen
		linesToConsume := numContentLines - numLinesLastItem
		maxTopItemIdx, maxTopItemLineOffset = m.getItemIdxAbove(maxTopItemIdx, maxTopItemLineOffset, linesToConsume)
	}
	return max(0, maxTopItemIdx), max(0, maxTopItemLineOffset)
}

// getHighlightsForItem returns highlights for the specific item index
func (m *Model[T]) getHighlightsForItem(itemIndex int) []Highlight {
	return m.content.getHighlightsForItem(itemIndex)
}

func (m *Model[T]) getNumVisibleItems() int {
	if !m.config.wrapText {
		return m.getNumContentLinesWithFooterVisible()
	}
	content := m.getVisibleContent()
	// return distinct number of items
	itemIndexSet := make(map[int]struct{})
	for _, i := range content.itemIndexes {
		itemIndexSet[i] = struct{}{}
	}
	return len(itemIndexSet)
}

func (m *Model[T]) styleSelection(selection string) string {
	split := surroundingAnsiRegex.Split(selection, -1)
	matches := surroundingAnsiRegex.FindAllString(selection, -1)
	var builder strings.Builder

	// pre-allocate the builder's capacity based on the selection string length
	// optional but can improve performance for longer strings
	builder.Grow(len(selection))

	for i, section := range split {
		if section != "" {
			builder.WriteString(m.display.styles.SelectedItemStyle.Render(section))
		}
		if i < len(split)-1 && i < len(matches) {
			builder.WriteString(matches[i])
		}
	}
	return builder.String()
}

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
}

func safeSliceUpToIdx[T any](s []T, i int) []T {
	if i > len(s) {
		return s
	}
	if i < 0 {
		return []T{}
	}
	return s[:i]
}
