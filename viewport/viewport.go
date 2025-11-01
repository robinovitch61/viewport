package viewport

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport/item"
)

// Terminology:
// - object: an object of type T that implements the Object interface, i.e. has an Item() method
// - item: the item.Item returned by an object's Item() method. A single item may span multiple viewport lines.
//         if selection is enabled, the item is the selectable unit
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
type Option[T Object] func(*Model[T])

// WithKeyMap sets the key mapping for the viewport
func WithKeyMap[T Object](keyMap KeyMap) Option[T] {
	return func(m *Model[T]) {
		m.navigation.keyMap = keyMap
	}
}

// WithStyles sets the styling for the viewport
func WithStyles[T Object](styles Styles) Option[T] {
	return func(m *Model[T]) {
		m.display.styles = styles
	}
}

// WithWrapText sets whether the viewport wraps text
func WithWrapText[T Object](wrap bool) Option[T] {
	return func(m *Model[T]) {
		m.SetWrapText(wrap)
	}
}

// WithSelectionEnabled sets whether the viewport allows selection
func WithSelectionEnabled[T Object](enabled bool) Option[T] {
	return func(m *Model[T]) {
		m.SetSelectionEnabled(enabled)
	}
}

// WithFooterEnabled sets whether the viewport shows the footer
func WithFooterEnabled[T Object](enabled bool) Option[T] {
	return func(m *Model[T]) {
		m.SetFooterEnabled(enabled)
	}
}

// Model represents a viewport component
type Model[T Object] struct {
	// content manages the content and selection state
	content *contentManager[T]

	// display handles rendering
	display *displayManager

	// navigation manages keyboard input and navigation logic
	navigation *navigationManager

	// config manages configuration options
	config *configuration
}

// New creates a new viewport model with reasonable defaults
func New[T Object](width, height int, opts ...Option[T]) (m *Model[T]) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	m = &Model[T]{}
	m.content = newContentManager[T]()
	m.display = newDisplayManager(width, height, DefaultStyles())
	m.navigation = newNavigationManager(DefaultKeyMap())
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
		navCtx := navigationContext{
			wrapText:        m.config.wrapText,
			dimensions:      m.display.bounds,
			numContentLines: m.getNumContentLines(),
			numVisibleItems: m.getNumVisibleItems(),
		}
		navResult := m.navigation.processKeyMsg(msg, navCtx)

		switch navResult.action {
		case actionUp:
			if m.navigation.selectionEnabled {
				m.SetSelectedItemIdx(m.content.getSelectedIdx() - navResult.selectionAmount)
			} else {
				m.scrollDownLines(-navResult.scrollAmount)
			}

		case actionDown:
			if m.navigation.selectionEnabled {
				m.SetSelectedItemIdx(m.content.getSelectedIdx() + navResult.selectionAmount)
			} else {
				m.scrollDownLines(navResult.scrollAmount)
			}

		case actionLeft:
			if !m.config.wrapText {
				m.SetXOffset(m.display.xOffset - navResult.scrollAmount)
			}

		case actionRight:
			if !m.config.wrapText {
				m.SetXOffset(m.display.xOffset + navResult.scrollAmount)
			}

		case actionHalfPageUp, actionPageUp:
			m.scrollDownLines(-navResult.scrollAmount)
			if m.navigation.selectionEnabled {
				m.SetSelectedItemIdx(m.content.getSelectedIdx() - navResult.selectionAmount)
			}

		case actionHalfPageDown, actionPageDown:
			m.scrollDownLines(navResult.scrollAmount)
			if m.navigation.selectionEnabled {
				m.SetSelectedItemIdx(m.content.getSelectedIdx() + navResult.selectionAmount)
			}

		case actionTop:
			if m.navigation.selectionEnabled {
				m.SetSelectedItemIdx(0)
			} else {
				m.display.topItemIdx = 0
				m.display.topItemLineOffset = 0
			}

		case actionBottom:
			if m.navigation.selectionEnabled {
				m.SetSelectedItemIdx(m.content.getSelectedIdx() + m.content.numItems())
			} else {
				maxItemIdx, maxTopLineOffset := m.maxItemIdxAndMaxTopLineOffset()
				m.display.setTopItemIdxAndOffset(maxItemIdx, maxTopLineOffset)
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
	itemIndexes := m.getVisibleContentItemIndexes()

	// pre-allocate capacity based on estimated size
	estimatedSize := (len(visibleHeaderLines) + len(itemIndexes) + 10) * (m.display.bounds.width + 1)
	builder.Grow(estimatedSize)

	// header lines
	for i := range visibleHeaderLines {
		headerItem := item.NewItem(visibleHeaderLines[i])
		line, _ := headerItem.Take(0, m.display.bounds.width, m.config.continuationIndicator, []item.Highlight{})
		builder.WriteString(line)
		builder.WriteByte('\n')
	}

	// content lines
	truncatedVisibleContentLines := make([]string, len(itemIndexes))
	currentItemIdxWidthToLeft := m.display.bounds.width * m.display.topItemLineOffset
	for idx, itemIdx := range itemIndexes {
		var truncated string
		if wrap {
			var widthTaken int
			truncated, widthTaken = m.content.objects[itemIdx].GetItem().Take(
				currentItemIdxWidthToLeft,
				m.display.bounds.width,
				"",
				m.getHighlightsForItem(itemIdx),
			)
			if idx+1 < len(itemIndexes) {
				nextItemIdx := itemIndexes[idx+1]
				if nextItemIdx != itemIdx {
					currentItemIdxWidthToLeft = 0
				} else {
					currentItemIdxWidthToLeft += widthTaken
				}
			}
		} else {
			// if not wrapped, items are not yet truncated or highlighted
			truncated, _ = m.content.objects[itemIdx].GetItem().Take(
				m.display.xOffset,
				m.display.bounds.width,
				m.config.continuationIndicator,
				m.getHighlightsForItem(itemIndexes[idx]),
			)
		}

		truncatedIsSelection := m.navigation.selectionEnabled && itemIndexes[idx] == m.content.getSelectedIdx()
		if truncatedIsSelection {
			truncated = m.styleSelection(truncated)
		}

		pannedRight := m.display.xOffset > 0
		itemHasWidth := m.content.objects[itemIdx].GetItem().Width() > 0
		pannedPastAllWidth := lipgloss.Width(truncated) == 0
		if !wrap && pannedRight && itemHasWidth && pannedPastAllWidth {
			// if panned right past where line ends, show continuation indicator
			continuation := item.NewItem(m.config.continuationIndicator)
			truncated, _ = continuation.Take(0, m.display.bounds.width, "", []item.Highlight{})
			if truncatedIsSelection {
				truncated = m.styleSelection(truncated)
			}
		}

		if truncatedIsSelection && lipgloss.Width(truncated) == 0 {
			// ensure selection is visible even if line empty
			truncated = m.styleSelection(" ")
		}

		truncatedVisibleContentLines[idx] = truncated
	}

	for i := range truncatedVisibleContentLines {
		builder.WriteString(truncatedVisibleContentLines[i])
		builder.WriteByte('\n')
	}

	nVisibleLines := len(itemIndexes)
	if m.config.footerEnabled {
		// pad so footer shows up at bottom
		padCount := max(0, m.getNumContentLines()-nVisibleLines)
		for i := 0; i < padCount; i++ {
			builder.WriteByte('\n')
		}
		builder.WriteString(m.getTruncatedFooterLine(itemIndexes))
	}

	return m.display.render(strings.TrimSuffix(builder.String(), "\n"))
}

// SetObjects sets the objects
func (m *Model[T]) SetObjects(objects []T) {
	var initialNumLinesAboveSelection int
	var stayAtTop, stayAtBottom bool
	var prevSelection T
	if m.navigation.selectionEnabled {
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			initialNumLinesAboveSelection = inView.numLinesAboveSelection
		}
		currentItems := m.content.objects
		selectedIdx := m.content.getSelectedIdx()
		if m.navigation.topSticky && len(currentItems) > 0 && selectedIdx == 0 {
			stayAtTop = true
		} else if m.navigation.bottomSticky && (len(currentItems) == 0 || (selectedIdx == len(currentItems)-1)) {
			stayAtBottom = true
		} else if m.content.compareFn != nil && 0 <= selectedIdx && selectedIdx < len(currentItems) {
			prevSelection = currentItems[selectedIdx]
		}
	}

	m.content.objects = objects
	// ensure scroll position is valid given new Item
	m.safelySetTopItemIdxAndOffset(m.display.topItemIdx, m.display.topItemLineOffset)

	// ensure xOffset is valid given new Item
	m.SetXOffset(m.display.xOffset)

	if m.navigation.selectionEnabled {
		if stayAtTop {
			m.content.setSelectedIdx(0)
		} else if stayAtBottom {
			m.content.setSelectedIdx(max(0, m.content.numItems()-1))
			m.scrollSoSelectionInView()
		} else if m.content.compareFn != nil {
			// TODO: could flag when items are sorted & comparable and use binary search instead
			found := false
			items := m.content.objects
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
			m.content.selectedIdx = clampValZeroToMax(m.content.selectedIdx, len(m.content.objects)-1)
			m.scrollSoSelectionInView()
			if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
				deltaLinesAbove := initialNumLinesAboveSelection - inView.numLinesAboveSelection
				m.scrollDownLines(-deltaLinesAbove)
			}
		}
	}
}

// SetTopSticky sets whether selection should stay at top when new Item added and selection is at the top
func (m *Model[T]) SetTopSticky(topSticky bool) {
	m.navigation.topSticky = topSticky
}

// SetBottomSticky sets whether selection should stay at bottom when new Item added and selection is at the bottom
func (m *Model[T]) SetBottomSticky(bottomSticky bool) {
	m.navigation.bottomSticky = bottomSticky
}

// SetSelectionEnabled sets whether the viewport allows line selection
func (m *Model[T]) SetSelectionEnabled(selectionEnabled bool) {
	wasEnabled := m.navigation.selectionEnabled
	m.navigation.selectionEnabled = selectionEnabled

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

// SetSelectionComparator sets the comparator function for maintaining the current selection when Item changes.
// If compareFn is non-nil, the viewport will try to maintain the current selection when Item changes.
func (m *Model[T]) SetSelectionComparator(compareFn CompareFn[T]) {
	m.content.compareFn = compareFn
}

// GetSelectionEnabled returns whether the viewport allows line selection
func (m *Model[T]) GetSelectionEnabled() bool {
	return m.navigation.selectionEnabled
}

// SetWrapText sets whether the viewport wraps text
func (m *Model[T]) SetWrapText(wrapText bool) {
	var initialNumLinesAboveSelection int
	if m.navigation.selectionEnabled {
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			initialNumLinesAboveSelection = inView.numLinesAboveSelection
		}
	}
	m.config.wrapText = wrapText
	m.display.topItemLineOffset = 0
	m.display.xOffset = 0
	if m.navigation.selectionEnabled {
		m.scrollSoSelectionInView()
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			deltaLinesAbove := initialNumLinesAboveSelection - inView.numLinesAboveSelection
			m.scrollDownLines(-deltaLinesAbove)
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
	m.setWidthHeight(width, m.display.bounds.height)
}

// GetWidth returns the viewport width
func (m *Model[T]) GetWidth() int {
	return m.display.bounds.width
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.setWidthHeight(m.display.bounds.width, height)
}

// GetHeight returns the viewport height
func (m *Model[T]) GetHeight() int {
	return m.display.bounds.height
}

// GetTopItemIdxAndLineOffset returns the current top item index and line offset within that item
func (m *Model[T]) GetTopItemIdxAndLineOffset() (int, int) {
	return m.display.topItemIdx, m.display.topItemLineOffset
}

// SetSelectedItemIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model[T]) SetSelectedItemIdx(selectedItemIdx int) {
	if !m.navigation.selectionEnabled {
		return
	}
	m.content.setSelectedIdx(selectedItemIdx)
	m.scrollSoSelectionInView()
}

// GetSelectedItemIdx returns the currently selected item index
func (m *Model[T]) GetSelectedItemIdx() int {
	if !m.navigation.selectionEnabled {
		return 0
	}
	return m.content.getSelectedIdx()
}

// GetSelectedItem returns a pointer to the currently selected item
func (m *Model[T]) GetSelectedItem() *T {
	if !m.navigation.selectionEnabled {
		return nil
	}
	return m.content.getSelectedItem()
}

// SetHeader sets the header, an unselectable set of lines at the top of the viewport
func (m *Model[T]) SetHeader(header []string) {
	m.content.header = header
}

// EnsureItemInView scrolls or pans the viewport so that the specified portion of an item is visible.
// If the desired item portion is above or below the current view, it scrolls vertically to bring it into view, leaving
// verticalPad number of lines of context if possible.
// If the desired item portion is to the left or right of the current view, it pans horizontally to bring it into view,
// leaving horizontalPad number of columns of context if possible.
// Afterwards, it's possible that the selection is out of view of the viewport.
func (m *Model[T]) EnsureItemInView(itemIdx, startWidth, endWidth, verticalPad, horizontalPad int) {
	if m.display.bounds.width == 0 {
		return
	}
	if m.content.isEmpty() {
		m.safelySetTopItemIdxAndOffset(0, 0)
		return
	}

	itemIdx, startWidth, endWidth = m.clampItemAndWidthParams(itemIdx, startWidth, endWidth)

	if m.config.wrapText {
		m.ensureWrappedPortionInView(itemIdx, startWidth, endWidth, verticalPad)
	} else {
		m.ensureUnwrappedItemVerticallyInView(itemIdx, verticalPad)
		m.ensureUnwrappedPortionHorizontallyInView(startWidth, endWidth, horizontalPad)
	}
}

// clampItemAndWidthParams clamps itemIdx, startWidth, and endWidth to valid ranges
func (m *Model[T]) clampItemAndWidthParams(itemIdx, startWidth, endWidth int) (int, int, int) {
	itemIdx = max(0, min(itemIdx, m.content.numItems()-1))
	itemWidth := m.content.objects[itemIdx].GetItem().Width()
	startWidth = max(0, min(startWidth, itemWidth))
	endWidth = max(startWidth, min(endWidth, itemWidth))
	return itemIdx, startWidth, endWidth
}

// ensureWrappedPortionInView ensures the specified portion is visible in wrapped mode
// TODO LEO: use verticalPad
func (m *Model[T]) ensureWrappedPortionInView(itemIdx, startWidth, endWidth, verticalPad int) {
	if !m.config.wrapText {
		panic("ensureWrappedPortionInView called when wrapText is false")
	}
	viewportWidth := m.display.bounds.width
	startLineOffset := startWidth / viewportWidth
	endLineOffset := (endWidth - 1) / viewportWidth
	if endWidth == 0 {
		endLineOffset = 0
	}

	// check if already in view
	if m.isWrappedPortionInView(itemIdx, startLineOffset, endLineOffset) {
		return
	}

	numLinesInPortion := endLineOffset - startLineOffset + 1
	numContentLines := m.getNumContentLines()

	// if portion larger than viewport, align top
	if numLinesInPortion >= numContentLines {
		m.safelySetTopItemIdxAndOffset(itemIdx, startLineOffset)
		return
	}

	scrollingDown := m.isScrollingDown(itemIdx, startLineOffset)
	if scrollingDown {
		// align bottom of portion with bottom of viewport
		m.safelySetTopItemIdxAndOffset(itemIdx, endLineOffset)
		linesFromTarget := m.linesBetweenCurrentTopAndTarget(itemIdx, endLineOffset)
		linesToScrollUp := max(0, numContentLines-1-linesFromTarget)
		m.scrollDownLines(-linesToScrollUp)
	} else {
		m.safelySetTopItemIdxAndOffset(itemIdx, startLineOffset)
	}
}

// isWrappedPortionInView checks if both start and end of a wrapped portion are currently visible
func (m *Model[T]) isWrappedPortionInView(itemIdx, startLineOffset, endLineOffset int) bool {
	if !m.config.wrapText {
		panic("isWrappedPortionInView called when wrapText is false")
	}
	itemIndexes := m.getVisibleContentItemIndexes()
	itemFirstSeenAt := -1
	portionStartInView := false
	portionEndInView := false

	for i, visibleItemIdx := range itemIndexes {
		if visibleItemIdx == itemIdx {
			if itemFirstSeenAt == -1 {
				itemFirstSeenAt = i
			}
			lineOffsetInItem := i - itemFirstSeenAt
			if m.display.topItemIdx == itemIdx && itemFirstSeenAt == 0 {
				lineOffsetInItem += m.display.topItemLineOffset
			}
			if lineOffsetInItem == startLineOffset {
				portionStartInView = true
			}
			if lineOffsetInItem == endLineOffset {
				portionEndInView = true
			}
		}
	}

	return portionStartInView && portionEndInView
}

// isScrollingDown determines if we're scrolling down based on current position
func (m *Model[T]) isScrollingDown(itemIdx, startLineOffset int) bool {
	if m.display.topItemIdx < itemIdx {
		return true
	}
	if m.display.topItemIdx == itemIdx && m.display.topItemLineOffset < startLineOffset {
		return true
	}
	return false
}

// linesBetweenCurrentTopAndTarget calculates how many lines separate current top line from target position
func (m *Model[T]) linesBetweenCurrentTopAndTarget(targetItemIdx, targetLineOffset int) int {
	if m.display.topItemIdx > targetItemIdx {
		panic("current top item index is after target item index")
	}

	if m.display.topItemIdx == targetItemIdx {
		return targetLineOffset - m.display.topItemLineOffset
	}

	// count lines from top item to target
	linesFromTarget := m.numLinesForItem(m.display.topItemIdx) - m.display.topItemLineOffset
	for idx := m.display.topItemIdx + 1; idx < targetItemIdx; idx++ {
		linesFromTarget += m.numLinesForItem(idx)
	}
	linesFromTarget += targetLineOffset

	return linesFromTarget
}

// ensureUnwrappedItemVerticallyInView scrolls vertically to bring item into view
func (m *Model[T]) ensureUnwrappedItemVerticallyInView(itemIdx, verticalPad int) {
	if m.config.wrapText {
		panic("ensureUnwrappedItemVerticallyInView called when wrapText is true")
	}
	itemIndexes := m.getVisibleContentItemIndexes()
	numContentLines := m.getNumContentLines()

	// check if already visible
	visiblePosition := -1
	for i, visibleItemIdx := range itemIndexes {
		if visibleItemIdx == itemIdx {
			visiblePosition = i
			break
		}
	}

	scrollingDown := m.display.topItemIdx < itemIdx

	// when padding can't be satisfied on both sides, center the item
	if verticalPad*2+1 > numContentLines {
		desiredLinesAbove := (numContentLines - 1) / 2
		targetTopItemIdx := max(0, itemIdx-desiredLinesAbove)
		m.safelySetTopItemIdxAndOffset(targetTopItemIdx, 0)
		return
	}

	desiredPad := min(verticalPad, numContentLines-1)

	if visiblePosition >= 0 {
		// item is visible, check if padding is respected
		linesAbove := visiblePosition
		linesBelow := len(itemIndexes) - visiblePosition - 1

		if linesAbove >= desiredPad && linesBelow >= desiredPad {
			return
		}

		// adjust position based on scrolling direction
		if scrollingDown {
			targetTopItemIdx := max(0, itemIdx-numContentLines+1+desiredPad)
			m.safelySetTopItemIdxAndOffset(targetTopItemIdx, 0)
		} else {
			targetTopItemIdx := max(0, itemIdx-desiredPad)
			m.safelySetTopItemIdxAndOffset(targetTopItemIdx, 0)
		}
		return
	}

	// not visible, position based on scrolling direction
	if scrollingDown {
		// scrolling down: leave desiredPad lines below
		targetTopItemIdx := max(0, itemIdx-numContentLines+1+desiredPad)
		m.safelySetTopItemIdxAndOffset(targetTopItemIdx, 0)
	} else {
		// scrolling up: leave desiredPad lines above
		targetTopItemIdx := max(0, itemIdx-desiredPad)
		m.safelySetTopItemIdxAndOffset(targetTopItemIdx, 0)
	}
}

// ensureUnwrappedPortionHorizontallyInView pans horizontally to bring portion into view
func (m *Model[T]) ensureUnwrappedPortionHorizontallyInView(startWidth, endWidth, horizontalPad int) {
	if m.config.wrapText {
		panic("ensureUnwrappedPortionHorizontallyInView called when wrapText is true")
	}
	viewportWidth := m.display.bounds.width
	currentXOffset := m.display.xOffset

	visibleStartWidth := currentXOffset + 1
	visibleEndWidth := currentXOffset + viewportWidth

	portionStartInView := startWidth >= visibleStartWidth && startWidth <= visibleEndWidth
	portionEndInView := endWidth >= visibleStartWidth && endWidth <= visibleEndWidth

	portionWidth := endWidth - startWidth
	panningRight := startWidth > visibleStartWidth

	// portion wider than viewport: align left edge with padding
	if portionWidth > viewportWidth {
		desiredColumnsLeft := min(horizontalPad, viewportWidth-1)
		targetXOffset := max(0, startWidth-desiredColumnsLeft)
		m.SetXOffset(targetXOffset)
		return
	}

	// when padding can't be satisfied on both sides, center the portion
	if horizontalPad*2+portionWidth > viewportWidth {
		desiredColumnsLeft := (viewportWidth - portionWidth) / 2
		targetXOffset := max(0, startWidth-desiredColumnsLeft)
		m.SetXOffset(targetXOffset)
		return
	}

	desiredPad := min(horizontalPad, viewportWidth-portionWidth)

	if portionStartInView && portionEndInView {
		// already fully visible, check if padding is respected
		columnsLeft := startWidth - currentXOffset
		columnsRight := currentXOffset + viewportWidth - endWidth

		if columnsLeft >= desiredPad && columnsRight >= desiredPad {
			return
		}

		// adjust position based on panning direction
		if panningRight {
			targetXOffset := max(0, endWidth+desiredPad-viewportWidth)
			m.SetXOffset(targetXOffset)
		} else {
			targetXOffset := max(0, startWidth-desiredPad)
			m.SetXOffset(targetXOffset)
		}
		return
	}

	// not visible, position based on panning direction
	if panningRight {
		// panning right: leave desiredPad columns to the right
		targetXOffset := max(0, endWidth+desiredPad-viewportWidth)
		m.SetXOffset(targetXOffset)
	} else {
		// panning left: leave desiredPad columns to the left
		targetXOffset := max(0, startWidth-desiredPad)
		m.SetXOffset(targetXOffset)
	}
}

// SetXOffset sets the horizontal offset, in terminal cell width, for panning when text wrapping is disabled
func (m *Model[T]) SetXOffset(widthOffset int) {
	if m.config.wrapText {
		return
	}
	maxXOffset := m.maxItemWidth() - m.display.bounds.width
	m.display.xOffset = max(0, min(maxXOffset, widthOffset))
}

// GetXOffsetWidth returns the horizontal offset, in terminal cell width, for panning when text wrapping is disabled
func (m *Model[T]) GetXOffsetWidth() int {
	if m.config.wrapText {
		return 0
	}
	return m.display.xOffset
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
		items := m.content.objects
		startIdx := clampValZeroToMax(m.display.topItemIdx, m.content.numItems()-1)
		numItemsToCheck := min(m.content.numItems()-startIdx, m.display.bounds.height)

		for i := 0; i < numItemsToCheck; i++ {
			itemIdx := startIdx + i
			if itemIdx >= m.content.numItems() {
				break
			}
			currItem := items[itemIdx].GetItem()
			if w := currItem.Width(); w > maxLineWidth {
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
	if m.display.bounds.width == 0 {
		return 0
	}
	if m.content.isEmpty() || itemIdx < 0 || itemIdx >= m.content.numItems() {
		return 0
	}
	items := m.content.objects
	return items[itemIdx].GetItem().NumWrappedLines(m.display.bounds.width)
}

func (m *Model[T]) setWidthHeight(width, height int) {
	m.display.setBounds(rectangle{width: width, height: height})
	if m.navigation.selectionEnabled {
		m.safelySetTopItemIdxAndOffset(m.content.getSelectedIdx(), 0)
	}
	m.safelySetTopItemIdxAndOffset(m.display.topItemIdx, m.display.topItemLineOffset)
}

func (m *Model[T]) safelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset int) {
	maxTopItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	if topItemIdx < 0 {
		topItemIdx = 0
		topItemLineOffset = 0
	}
	if topItemIdx > maxTopItemIdx {
		topItemIdx = maxTopItemIdx
		topItemLineOffset = maxTopItemLineOffset
	}
	if topItemIdx == maxTopItemIdx {
		topItemLineOffset = clampValZeroToMax(topItemLineOffset, maxTopItemLineOffset)
	}
	m.display.setTopItemIdxAndOffset(topItemIdx, topItemLineOffset)
}

// getNumContentLines returns the number of lines of between the header and footer
func (m *Model[T]) getNumContentLines() int {
	return m.display.getNumContentLines(len(m.getVisibleHeaderLines()), true)
}

func (m *Model[T]) scrollSoSelectionInView() {
	if !m.navigation.selectionEnabled {
		panic("scrollSoSelectionInView called when selection is not enabled")
	}
	selectedItem := m.content.getSelectedItem()
	if selectedItem == nil {
		return
	}
	selectedItemWidth := (*selectedItem).GetItem().Width()
	startWidth := 0
	endWidth := selectedItemWidth
	if !m.config.wrapText && m.display.xOffset > 0 {
		if selectedItemWidth < m.display.xOffset {
			// ensure the selection is visible by scrolling, but maintain xOffset if possible
			prevXOffset := m.display.xOffset
			m.EnsureItemInView(m.content.selectedIdx, 0, 0, 0, 0)
			m.SetXOffset(prevXOffset)
			return
		}
		startWidth = m.display.xOffset
		endWidth = m.display.xOffset + m.display.bounds.width - 1
	}
	m.EnsureItemInView(m.content.selectedIdx, startWidth, endWidth, 0, 0)
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

// scrollDownLines edits topItemIdx and topItemLineOffset to scroll the viewport by n lines (negative for up, positive for down)
func (m *Model[T]) scrollDownLines(numLinesDown int) {
	if numLinesDown == 0 {
		return
	}

	// scrolling down past bottom
	if numLinesDown > 0 && m.isScrolledToBottom() {
		return
	}

	// scrolling up past top
	if numLinesDown < 0 && m.display.topItemIdx == 0 && m.display.topItemLineOffset == 0 {
		return
	}

	newTopItemIdx, newTopItemLineOffset := m.display.topItemIdx, m.display.topItemLineOffset
	if !m.config.wrapText {
		newTopItemIdx = m.display.topItemIdx + numLinesDown
	} else {
		// wrapped
		if numLinesDown < 0 { // scrolling up
			if newTopItemLineOffset >= -numLinesDown {
				// same item, just change offset
				newTopItemLineOffset += numLinesDown
			} else {
				// need to scroll up through multiple items
				linesToConsume := -numLinesDown - newTopItemLineOffset
				newTopItemIdx, newTopItemLineOffset = m.getItemIdxAbove(newTopItemIdx, newTopItemLineOffset, linesToConsume)
			}
		} else { // scrolling down
			numLinesInTopItem := m.numLinesForItem(newTopItemIdx)
			if newTopItemLineOffset+numLinesDown < numLinesInTopItem {
				// same item, just change offset
				newTopItemLineOffset += numLinesDown
			} else {
				// need to scroll down through multiple items
				linesToConsume := numLinesDown - (numLinesInTopItem - (newTopItemLineOffset + 1))
				newTopItemIdx, newTopItemLineOffset = m.getItemIdxBelow(newTopItemIdx, linesToConsume)
			}
		}
	}
	m.safelySetTopItemIdxAndOffset(newTopItemIdx, newTopItemLineOffset)
	m.SetXOffset(m.display.xOffset)
}

// getVisibleHeaderLines returns the lines of header that are visible in the viewport as strings.
// header lines will take precedence over content and footer if there is not enough vertical height
func (m *Model[T]) getVisibleHeaderLines() []string {
	if m.display.bounds.height == 0 {
		return nil
	}

	headerItems := make([]item.Item, len(m.content.header))
	for i := range m.content.header {
		headerItems[i] = item.NewItem(m.content.header[i])
	}

	itemIndexes := m.getItemIndexesSpanningLines(
		0,
		0,
		m.display.bounds.height,
		headerItems,
	)

	headerLines := make([]string, len(itemIndexes))
	currentItemIdxWidthToLeft := 0
	for idx, itemIdx := range itemIndexes {
		var truncated string
		if m.config.wrapText {
			currentItemIdx := itemIndexes[idx]
			var widthTaken int
			truncated, widthTaken = headerItems[itemIdx].Take(
				currentItemIdxWidthToLeft,
				m.display.bounds.width,
				"",
				[]item.Highlight{}, // no highlights for header
			)
			if idx+1 < len(itemIndexes) {
				nextItemIdx := itemIndexes[idx+1]
				if nextItemIdx != currentItemIdx {
					currentItemIdxWidthToLeft = 0
				} else {
					currentItemIdxWidthToLeft += widthTaken
				}
			}
		} else {
			// if not wrapped, items are not yet truncated or highlighted
			truncated, _ = headerItems[itemIdx].Take(
				0, // header doesn't pan horizontally
				m.display.bounds.width,
				m.config.continuationIndicator,
				[]item.Highlight{}, // no highlights for header
			)
		}
		headerLines[idx] = truncated
	}

	return headerLines
}

// getVisibleContentItemIndexes returns the item indexes of content that are visible in the viewport
func (m *Model[T]) getVisibleContentItemIndexes() []int {
	if m.display.bounds.width == 0 || m.content.isEmpty() {
		return nil
	}

	numLinesAfterHeader := max(0, m.display.bounds.height-len(m.getVisibleHeaderLines()))

	itemIndexes := m.getItemIndexesSpanningLines(
		m.display.topItemIdx,
		m.display.topItemLineOffset,
		numLinesAfterHeader,
		renderAll(m.content.objects),
	)
	if len(itemIndexes) == 0 {
		return nil
	}

	if m.config.footerEnabled {
		// leave one line for the footer
		itemIndexes = safeSliceUpToIdx(itemIndexes, numLinesAfterHeader-1)
	}
	return itemIndexes
}

func renderAll[T Object](itemGetters []T) []item.Item {
	items := make([]item.Item, len(itemGetters))
	for i := range itemGetters {
		items[i] = itemGetters[i].GetItem()
	}
	return items
}

// getItemIndexesSpanningLines returns the item indexes for each line given a top item index, offset and num lines
func (m *Model[T]) getItemIndexesSpanningLines(
	topItemIdx int,
	topItemLineOffset int,
	totalNumLines int,
	allItems []item.Item,
) []int {
	if len(allItems) == 0 || totalNumLines == 0 {
		return nil
	}

	var itemIndexes []int

	addLine := func(itemIndex int) bool {
		itemIndexes = append(itemIndexes, itemIndex)
		return len(itemIndexes) == totalNumLines
	}

	currItemIdx := clampValZeroToMax(topItemIdx, len(allItems)-1)

	currItem := allItems[currItemIdx]
	done := totalNumLines == 0
	if done {
		return itemIndexes
	}

	if m.config.wrapText {
		// first item has potentially fewer lines depending on the line offset
		numLines := max(0, currItem.NumWrappedLines(m.display.bounds.width)-topItemLineOffset)
		for range numLines {
			// adding untruncated, unstyled items
			done = addLine(currItemIdx)
			if done {
				break
			}
		}

		for !done {
			currItemIdx++
			if currItemIdx >= len(allItems) {
				done = true
			} else {
				currItem = allItems[currItemIdx]
				numLines = currItem.NumWrappedLines(m.display.bounds.width)
				for range numLines {
					// adding untruncated, unstyled items
					done = addLine(currItemIdx)
					if done {
						break
					}
				}
			}
		}
	} else {
		done = addLine(currItemIdx)
		for !done {
			currItemIdx++
			if currItemIdx >= len(allItems) {
				done = true
			} else {
				done = addLine(currItemIdx)
			}
		}
	}
	return itemIndexes
}

// TODO LEO: reuse this for selection styling
//func (m *Model[T]) highlightStyle(itemIdx int) lipgloss.Style {
//	return m.display.getHighlightStyle(m.navigation.selectionEnabled && itemIdx == m.content.getSelectedIdx())
//}

func (m *Model[T]) getTruncatedFooterLine(visibleContentItemIndexes []int) string {
	numerator := m.content.getSelectedIdx() + 1 // 0 indexed
	denominator := m.content.numItems()
	if denominator == 0 {
		return ""
	}
	if !m.config.footerEnabled {
		panic("getTruncatedFooterLine called when footer should not be shown")
	}
	if len(visibleContentItemIndexes) == 0 {
		return ""
	}

	var footerString string

	// if selection is disabled, numerator should be item index of bottom visible line
	if !m.navigation.selectionEnabled {
		numerator = visibleContentItemIndexes[len(visibleContentItemIndexes)-1] + 1
		if m.config.wrapText && numerator == denominator && !m.isScrolledToBottom() {
			// if wrapped && bottom visible line is max item index, but actually not fully scrolled to bottom, show 99%
			footerString = fmt.Sprintf("99%% (%d/%d)", numerator, denominator)
		}
	}

	if footerString == "" {
		percentScrolled := percent(numerator, denominator)
		footerString = fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)
	}

	footerItem := item.NewItem(footerString)
	f, _ := footerItem.Take(0, m.display.bounds.width, m.config.continuationIndicator, []item.Highlight{})
	return m.display.styles.FooterStyle.Render(f)
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
	if !m.navigation.selectionEnabled {
		panic("selectionInViewInfo called when selection is disabled")
	}
	itemIndexes := m.getVisibleContentItemIndexes()
	numLinesSelectionInView := 0
	numLinesAboveSelection := 0
	assignedNumLinesAboveSelection := false
	for i := range itemIndexes {
		if itemIndexes[i] == m.content.getSelectedIdx() {
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
	numContentLines := max(0, m.display.bounds.height-headerLines-1)

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
func (m *Model[T]) getHighlightsForItem(itemIndex int) []item.Highlight {
	return m.content.getItemHighlightsForItem(itemIndex)
}

func (m *Model[T]) getNumVisibleItems() int {
	if !m.config.wrapText {
		return m.getNumContentLines()
	}
	itemIndexes := m.getVisibleContentItemIndexes()
	// return distinct number of items
	itemIndexSet := make(map[int]struct{})
	for _, i := range itemIndexes {
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
	if b == 0 {
		return 100
	}
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

func clampValZeroToMax(v, maximum int) int {
	return max(0, min(maximum, v))
}
