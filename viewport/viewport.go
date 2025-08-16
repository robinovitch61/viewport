package viewport

import (
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
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
// wrap disabled, line overflow:
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

// Model represents a viewport component
type Model[T Renderable] struct {
	// content manages the content and selection state
	content *ContentManager[T]

	// display handles rendering
	display *DisplayManager

	// navigation manages keyboard input and navigation logic
	navigation *NavigationManager

	// config manages configuration options
	config *Configuration
}

// New creates a new viewport model with reasonable defaults
func New[T Renderable](width, height int, keyMap KeyMap, styles Styles) (m *Model[T]) {
	m = &Model[T]{}
	m.content = NewContentManager[T]()
	m.display = NewDisplayManager(width, height, styles)
	m.navigation = NewNavigationManager(keyMap)
	m.config = NewConfiguration()
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
			WrapText:        m.config.WrapText,
			Dimensions:      m.display.Bounds,
			NumContentLines: m.getNumContentLines(),
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
			if !m.config.WrapText {
				m.viewLeft(navResult.ScrollAmount)
			}

		case ActionRight:
			if !m.config.WrapText {
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
				m.display.TopItemIdx = 0
				m.display.TopItemLineOffset = 0
			}

		case ActionBottom:
			if m.navigation.SelectionEnabled {
				m.selectedItemIdxDown(m.content.NumItems())
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

	visibleHeaderLines := m.getVisibleHeaderLines()
	visibleContentLines := m.getVisibleContentLines()

	// pre-allocate capacity based on estimated size
	estimatedSize := (len(visibleHeaderLines) + len(visibleContentLines.lines) + 10) * (m.display.Bounds.Width + 1)
	builder.Grow(estimatedSize)

	for i := range visibleHeaderLines {
		lineBuffer := linebuffer.New(visibleHeaderLines[i])
		line, _ := lineBuffer.Take(0, m.display.Bounds.Width, m.config.ContinuationIndicator, linebuffer.HighlightData{}, lipgloss.NewStyle())
		builder.WriteString(line)
		builder.WriteByte('\n')
	}

	truncatedVisibleContentLines := make([]string, len(visibleContentLines.lines))
	for i := range visibleContentLines.lines {
		var truncated string
		if m.config.WrapText {
			truncated = visibleContentLines.lines[i].Content()
		} else {
			lineBuffer := visibleContentLines.lines[i]
			truncated, _ = lineBuffer.Take(
				m.display.XOffset,
				m.display.Bounds.Width,
				m.config.ContinuationIndicator,
				m.content.ToHighlight,
				m.highlightStyle(visibleContentLines.itemIndexes[i]),
			)
		}

		isSelection := m.navigation.SelectionEnabled && visibleContentLines.itemIndexes[i] == m.content.GetSelectedIdx()
		if isSelection {
			truncated = m.styleSelection(truncated)
		}

		if !m.config.WrapText && m.display.XOffset > 0 && lipgloss.Width(truncated) == 0 && visibleContentLines.lines[i].Width() > 0 {
			// if panned right past where line ends, show continuation indicator
			lineBuffer := linebuffer.New(m.getLineContinuationIndicator())
			truncated, _ = lineBuffer.Take(0, m.display.Bounds.Width, "", linebuffer.HighlightData{}, lipgloss.NewStyle())
			if isSelection {
				truncated = m.styleSelection(truncated)
			}
		}

		if isSelection && truncated == "" {
			// ensure selection is visible even if LineBuffer empty
			truncated = m.styleSelection(" ")
		}

		truncatedVisibleContentLines[i] = truncated
	}

	for i := range truncatedVisibleContentLines {
		builder.WriteString(truncatedVisibleContentLines[i])
		builder.WriteByte('\n')
	}

	nVisibleLines := len(visibleContentLines.lines)
	if visibleContentLines.showFooter {
		// pad so footer shows up at bottom
		padCount := max(0, m.getNumContentLines()-nVisibleLines-1) // 1 for footer itself
		for i := 0; i < padCount; i++ {
			builder.WriteByte('\n')
		}
		builder.WriteString(m.getTruncatedFooterLine(visibleContentLines))
	} else {
		if builder.Len() > 0 {
			content := builder.String()
			return m.display.RenderFinalView(strings.TrimSuffix(content, "\n"))
		}
	}

	return m.display.RenderFinalView(builder.String())
}

// SetKeyMap sets the key mapping for navigation controls.
func (m *Model[T]) SetKeyMap(keyMap KeyMap) {
	m.navigation.KeyMap = keyMap
}

// SetStyles sets the styling configuration for the viewport
func (m *Model[T]) SetStyles(styles Styles) {
	m.display.Styles = styles
}

// SetContent sets the LineBuffer, the selectable set of lines in the viewport
func (m *Model[T]) SetContent(content []T) {
	var initialNumLinesAboveSelection int
	var stayAtTop, stayAtBottom bool
	var prevSelection T
	if m.navigation.SelectionEnabled {
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			initialNumLinesAboveSelection = inView.numLinesAboveSelection
		}
		currentItems := m.content.Items
		selectedIdx := m.content.GetSelectedIdx()
		if m.navigation.TopSticky && len(currentItems) > 0 && selectedIdx == 0 {
			stayAtTop = true
		} else if m.navigation.BottomSticky && (len(currentItems) == 0 || (selectedIdx == len(currentItems)-1)) {
			stayAtBottom = true
		} else if m.content.CompareFn != nil && 0 <= selectedIdx && selectedIdx < len(currentItems) {
			prevSelection = currentItems[selectedIdx]
		}
	}

	m.content.Items = content
	// ensure scroll position is valid given new LineBuffer
	m.safelySetTopItemIdxAndOffset(m.display.TopItemIdx, m.display.TopItemLineOffset)

	// ensure xOffset is valid given new LineBuffer
	m.safelySetXOffset(m.display.XOffset)

	if m.navigation.SelectionEnabled {
		if stayAtTop {
			m.content.SetSelectedIdx(0)
		} else if stayAtBottom {
			m.content.SetSelectedIdx(max(0, m.content.NumItems()-1))
			m.scrollSoSelectionInView()
		} else if m.content.CompareFn != nil {
			// TODO: could flag when LineBuffer is sorted & comparable and use binary search instead
			found := false
			items := m.content.Items
			for i := range items {
				if m.content.CompareFn(items[i], prevSelection) {
					m.content.SetSelectedIdx(i)
					found = true
					break
				}
			}
			if !found {
				m.content.SetSelectedIdx(0)
			}
		}

		// when staying at bottom, just want to scroll so selection in view, which is done above
		if !stayAtBottom {
			m.content.ValidateSelectedIdx()
			m.scrollSoSelectionInView()
			if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
				m.scrollUp(initialNumLinesAboveSelection - inView.numLinesAboveSelection)
			}
		}
	}
}

// SetTopSticky sets whether selection should stay at top when new LineBuffer added and selection is at the top
func (m *Model[T]) SetTopSticky(topSticky bool) {
	m.navigation.TopSticky = topSticky
}

// SetBottomSticky sets whether selection should stay at bottom when new LineBuffer added and selection is at the bottom
func (m *Model[T]) SetBottomSticky(bottomSticky bool) {
	m.navigation.BottomSticky = bottomSticky
}

// SetSelectionEnabled sets whether the viewport allows line selection
func (m *Model[T]) SetSelectionEnabled(selectionEnabled bool) {
	wasEnabled := m.navigation.SelectionEnabled
	m.navigation.SelectionEnabled = selectionEnabled

	// when enabling selection, set the selected item to the top visible item and ensure the top line is in view
	if selectionEnabled && !wasEnabled && !m.content.IsEmpty() {
		topVisibleItemIdx := clampValZeroToMax(m.display.TopItemIdx, m.content.NumItems()-1)
		m.content.SetSelectedIdx(topVisibleItemIdx)
		m.scrollSoSelectionInView()
	}
}

// SetFooterEnabled sets whether the viewport shows the footer when it overflows
func (m *Model[T]) SetFooterEnabled(footerEnabled bool) {
	m.config.FooterEnabled = footerEnabled
}

// SetSelectionComparator sets the comparator function for maintaining the current selection when LineBuffer changes.
// If compareFn is non-nil, the viewport will try to maintain the current selection when LineBuffer changes.
func (m *Model[T]) SetSelectionComparator(compareFn CompareFn[T]) {
	m.content.CompareFn = compareFn
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
	m.config.WrapText = wrapText
	m.display.TopItemLineOffset = 0
	m.display.XOffset = 0
	if m.navigation.SelectionEnabled {
		m.scrollSoSelectionInView()
		if inView := m.selectionInViewInfo(); inView.numLinesSelectionInView > 0 {
			m.scrollUp(initialNumLinesAboveSelection - inView.numLinesAboveSelection)
			m.scrollSoSelectionInView()
		}
	}
	m.safelySetTopItemIdxAndOffset(m.display.TopItemIdx, m.display.TopItemLineOffset)
}

// GetWrapText returns whether the viewport wraps text
func (m *Model[T]) GetWrapText() bool {
	return m.config.WrapText
}

// SetWidth sets the viewport's width
func (m *Model[T]) SetWidth(width int) {
	m.setWidthHeight(width, m.display.Bounds.Height)
}

// SetHeight sets the viewport's height, including header and footer
func (m *Model[T]) SetHeight(height int) {
	m.setWidthHeight(m.display.Bounds.Width, height)
}

// SetSelectedItemIdx sets the selected context index. Automatically puts selection in view as necessary
func (m *Model[T]) SetSelectedItemIdx(selectedItemIdx int) {
	if !m.navigation.SelectionEnabled || m.getNumContentLines() == 0 {
		return
	}
	m.content.SetSelectedIdx(selectedItemIdx)
	m.scrollSoSelectionInView()
}

// GetSelectedItemIdx returns the currently selected item index
func (m *Model[T]) GetSelectedItemIdx() int {
	if !m.navigation.SelectionEnabled {
		return 0
	}
	return m.content.GetSelectedIdx()
}

// GetSelectedItem returns a pointer to the currently selected item
func (m *Model[T]) GetSelectedItem() *T {
	if !m.navigation.SelectionEnabled {
		return nil
	}
	return m.content.GetSelectedItem()
}

// SetStringToHighlight sets a string to highlight in the viewport. Can only set string or regex, not both.
func (m *Model[T]) SetStringToHighlight(h string) {
	m.content.ToHighlight = linebuffer.HighlightData{
		StringToHighlight: h,
		IsRegex:           false,
	}
}

// SetRegexToHighlight sets a regex to highlight in the viewport. Can only set string or regex, not both.
func (m *Model[T]) SetRegexToHighlight(r *regexp.Regexp) {
	m.content.ToHighlight = linebuffer.HighlightData{
		RegexPatternToHighlight: r,
		IsRegex:                 true,
	}
}

// SetHeader sets the header, an unselectable set of lines at the top of the viewport
func (m *Model[T]) SetHeader(header []string) {
	m.content.Header = header
}

// GetWidth returns the viewport width
func (m *Model[T]) GetWidth() int {
	return m.display.Bounds.Width
}

// GetHeight returns the viewport height
func (m *Model[T]) GetHeight() int {
	return m.display.Bounds.Height
}

// ScrollSoItemIdxInView scrolls the viewport to ensure the specified item index is visible.
func (m *Model[T]) ScrollSoItemIdxInView(itemIdx int) {
	if m.content.IsEmpty() {
		m.safelySetTopItemIdxAndOffset(0, 0)
		return
	}
	originalTopItemIdx, originalTopItemLineOffset := m.display.TopItemIdx, m.display.TopItemLineOffset

	numLinesInItem := 1
	if m.config.WrapText {
		numLinesInItem = m.numLinesForItem(itemIdx)
	}

	visibleLines := m.getVisibleContentLines()
	numItemLinesInView := 0
	for i := range visibleLines.itemIndexes {
		if visibleLines.itemIndexes[i] == itemIdx {
			numItemLinesInView++
		}
	}
	if numLinesInItem != numItemLinesInView {
		if m.display.TopItemIdx < itemIdx {
			// if item is below, scroll until it's fully in view at the bottom
			m.display.TopItemIdx = itemIdx
			m.display.TopItemLineOffset = 0
			// then scroll up so that item is at the bottom, unless it already takes up the whole screen
			m.scrollUp(max(0, m.getNumContentLines()-numLinesInItem))
		} else {
			// if item above, scroll until it's fully in view at the top
			m.display.TopItemIdx = itemIdx
			m.display.TopItemLineOffset = 0
		}
	}

	if m.navigation.SelectionEnabled {
		// if scrolled such that selection is fully out of view, undo it
		if m.selectionInViewInfo().numLinesSelectionInView == 0 {
			m.display.TopItemIdx = originalTopItemIdx
			m.display.TopItemLineOffset = originalTopItemLineOffset
		}
	}
}

func (m *Model[T]) maxLineWidth() int {
	maxLineWidth := 0

	headerLines := m.getVisibleHeaderLines()
	for i := range headerLines {
		if w := lipgloss.Width(headerLines[i]); w > maxLineWidth {
			maxLineWidth = w
		}
	}

	visibleContentLines := m.getVisibleContentLines()
	for i := range visibleContentLines.lines {
		if w := visibleContentLines.lines[i].Width(); w > maxLineWidth {
			maxLineWidth = w
		}
	}

	if visibleContentLines.showFooter {
		footerLine := m.getTruncatedFooterLine(visibleContentLines)
		if w := lipgloss.Width(footerLine); w > maxLineWidth {
			maxLineWidth = w
		}
	}

	return maxLineWidth
}

func (m *Model[T]) numLinesForItem(itemIdx int) int {
	if m.display.Bounds.Width == 0 {
		return 0
	}
	if m.content.IsEmpty() || itemIdx < 0 || itemIdx >= m.content.NumItems() {
		return 0
	}
	items := m.content.Items
	lb := items[itemIdx].Render()
	return len(lb.WrappedLines(m.display.Bounds.Width, m.display.Bounds.Height, linebuffer.HighlightData{}, lipgloss.NewStyle()))
}

func (m *Model[T]) safelySetXOffset(n int) {
	maxXOffset := m.maxLineWidth() - m.display.Bounds.Width
	m.display.XOffset = max(0, min(maxXOffset, n))
}

func (m *Model[T]) setWidthHeight(width, height int) {
	m.display.SetBounds(width, height)
	if m.navigation.SelectionEnabled {
		m.scrollSoSelectionInView()
	}
	m.safelySetTopItemIdxAndOffset(m.display.TopItemIdx, m.display.TopItemLineOffset)
}

func (m *Model[T]) safelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset int) {
	maxTopItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	m.display.SafelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset, maxTopItemIdx, maxTopItemLineOffset)
}

// getNumContentLines returns the number of lines of between the header and footer
func (m *Model[T]) getNumContentLines() int {
	visibleContentLines := m.getVisibleContentLines()
	return m.display.GetNumContentLines(len(m.getVisibleHeaderLines()), visibleContentLines.showFooter)
}

func (m *Model[T]) scrollSoSelectionInView() {
	if !m.navigation.SelectionEnabled {
		panic("scrollSoSelectionInView called when selection is not enabled")
	}
	m.ScrollSoItemIdxInView(m.content.GetSelectedIdx())
}

func (m *Model[T]) selectedItemIdxDown(n int) {
	m.SetSelectedItemIdx(m.content.GetSelectedIdx() + n)
}

func (m *Model[T]) selectedItemIdxUp(n int) {
	m.SetSelectedItemIdx(m.content.GetSelectedIdx() - n)
}

func (m *Model[T]) scrollDown(n int) {
	m.scrollByNLines(n)
}

func (m *Model[T]) scrollUp(n int) {
	m.scrollByNLines(-n)
}

func (m *Model[T]) viewLeft(n int) {
	m.safelySetXOffset(m.display.XOffset - n)
}

func (m *Model[T]) viewRight(n int) {
	m.safelySetXOffset(m.display.XOffset + n)
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
	if n < 0 && m.display.TopItemIdx == 0 && m.display.TopItemLineOffset == 0 {
		return
	}

	newTopItemIdx, newTopItemLineOffset := m.display.TopItemIdx, m.display.TopItemLineOffset
	if !m.config.WrapText {
		newTopItemIdx = m.display.TopItemIdx + n
	} else {
		// wrapped
		if n < 0 { // negative n, scrolling up
			// up
			if newTopItemLineOffset >= -n {
				// same item, just change offset
				newTopItemLineOffset += n
			} else {
				// take lines from items until scrolled up desired amount
				n += newTopItemLineOffset
				for n < 0 {
					newTopItemIdx--
					if newTopItemIdx < 0 {
						// scrolled up past top - stay at top
						newTopItemIdx = 0
						newTopItemLineOffset = 0
						break
					}
					numLinesInTopItem := m.numLinesForItem(newTopItemIdx)
					for i := range numLinesInTopItem {
						n++
						if n == 0 {
							newTopItemLineOffset = numLinesInTopItem - (i + 1)
							break
						}
					}
				}
			}
		} else { // positive n, scrolling down
			numLinesInTopItem := m.numLinesForItem(newTopItemIdx)
			if newTopItemLineOffset+n < numLinesInTopItem {
				// same item, just change offset
				newTopItemLineOffset += n
			} else {
				// take lines from items until scrolled down desired amount
				n -= numLinesInTopItem - (newTopItemLineOffset + 1)
				for n > 0 {
					newTopItemIdx++
					if newTopItemIdx >= m.content.NumItems() {
						newTopItemIdx = m.content.NumItems() - 1
						break
					}
					numLinesInTopItem = m.numLinesForItem(newTopItemIdx)
					for i := range numLinesInTopItem {
						n--
						if n == 0 {
							newTopItemLineOffset = i
							break
						}
					}
				}
			}
		}
	}
	m.safelySetTopItemIdxAndOffset(newTopItemIdx, newTopItemLineOffset)
	m.safelySetXOffset(m.display.XOffset)
}

// getVisibleHeaderLines returns the lines of header that are visible in the viewport
// header lines will take precedence over content and footer if there is not enough vertical height
func (m *Model[T]) getVisibleHeaderLines() []string {
	if m.display.Bounds.Height == 0 {
		return nil
	}

	header := m.content.Header
	if !m.config.WrapText {
		return safeSliceUpToIdx(header, m.display.Bounds.Height)
	}
	// wrapped
	var wrappedHeaderLines []string
	for _, s := range header {
		lb := linebuffer.New(s)
		wrappedHeaderLines = append(
			wrappedHeaderLines,
			lb.WrappedLines(m.display.Bounds.Width, m.display.Bounds.Height, linebuffer.HighlightData{}, lipgloss.NewStyle())...,
		)
	}
	return safeSliceUpToIdx(wrappedHeaderLines, m.display.Bounds.Height)
}

type visibleContentLinesResult struct {
	// lines is the untruncated visible lines, each corresponding to one terminal row
	lines []linebuffer.LineBufferer
	// itemIndexes is the index of the item in allItems that corresponds to each line. len(itemIndexes) == len(lines)
	itemIndexes []int
	// showFooter is true if the footer should be shown due to the num visible lines exceeding the vertical space
	showFooter bool
}

// getVisibleContentLines returns the lines of content that are visible in the viewport given vertical scroll position
// and the content. It also returns the item index for each associated visible line and whether or not to show the footer
func (m *Model[T]) getVisibleContentLines() visibleContentLinesResult {
	if m.display.Bounds.Width == 0 {
		return visibleContentLinesResult{lines: nil, itemIndexes: nil, showFooter: false}
	}
	if m.content.IsEmpty() {
		return visibleContentLinesResult{lines: nil, itemIndexes: nil, showFooter: false}
	}

	var contentLines []linebuffer.LineBufferer
	var itemIndexes []int

	numLinesAfterHeader := max(0, m.display.Bounds.Height-len(m.getVisibleHeaderLines()))

	addLine := func(l linebuffer.LineBufferer, itemIndex int) bool {
		contentLines = append(contentLines, l)
		itemIndexes = append(itemIndexes, itemIndex)
		return len(contentLines) == numLinesAfterHeader
	}
	addLines := func(ls []linebuffer.LineBufferer, itemIndex int) bool {
		for i := range ls {
			if addLine(ls[i], itemIndex) {
				return true
			}
		}
		return false
	}

	items := m.content.Items
	currItemIdx := clampValZeroToMax(m.display.TopItemIdx, m.content.NumItems()-1)

	currItem := items[currItemIdx]
	done := numLinesAfterHeader == 0
	if done {
		return visibleContentLinesResult{lines: contentLines, itemIndexes: itemIndexes, showFooter: false}
	}

	if m.config.WrapText {
		lb := currItem.Render()
		itemLines := lb.WrappedLines(m.display.Bounds.Width, m.display.Bounds.Height, m.content.ToHighlight, m.highlightStyle(currItemIdx))
		offsetLines := safeSliceFromIdx(itemLines, m.display.TopItemLineOffset)
		done = addLines(toLineBuffers(offsetLines), currItemIdx)

		for !done {
			currItemIdx++
			if currItemIdx >= m.content.NumItems() {
				done = true
			} else {
				currItem = items[currItemIdx]
				lb = currItem.Render()
				itemLines = lb.WrappedLines(m.display.Bounds.Width, m.display.Bounds.Height, m.content.ToHighlight, m.highlightStyle(currItemIdx))
				done = addLines(toLineBuffers(itemLines), currItemIdx)
			}
		}
	} else {
		done = addLine(currItem.Render(), currItemIdx)
		for !done {
			currItemIdx++
			if currItemIdx >= m.content.NumItems() {
				done = true
			} else {
				currItem = items[currItemIdx]
				done = addLine(currItem.Render(), currItemIdx)
			}
		}
	}

	scrolledToTop := m.display.TopItemIdx == 0 && m.display.TopItemLineOffset == 0
	var showFooter bool
	if scrolledToTop && len(contentLines)+1 >= numLinesAfterHeader {
		// if seeing all the LineBuffer on screen, show footer
		// if one blank line at bottom, still show footer
		// if two blank lines at bottom, do not show footer
		showFooter = true
	}
	if !scrolledToTop {
		// if scrolled at all, should be showing footer
		showFooter = true
	}

	if !m.config.FooterEnabled {
		showFooter = false
	}

	if showFooter {
		// num visible lines exceeds vertical space, leave one line for the footer
		contentLines = safeSliceUpToIdx(contentLines, numLinesAfterHeader-1)
		itemIndexes = safeSliceUpToIdx(itemIndexes, numLinesAfterHeader-1)
	}
	return visibleContentLinesResult{lines: contentLines, itemIndexes: itemIndexes, showFooter: showFooter}
}

func (m *Model[T]) highlightStyle(itemIdx int) lipgloss.Style {
	return m.display.GetHighlightStyle(m.navigation.SelectionEnabled && itemIdx == m.content.GetSelectedIdx())
}

func (m *Model[T]) getTruncatedFooterLine(visibleContentLines visibleContentLinesResult) string {
	numerator := m.content.GetSelectedIdx() + 1 // 0th line is 1st
	denominator := m.content.NumItems()
	if !visibleContentLines.showFooter {
		panic("getTruncatedFooterLine called when footer should not be shown")
	}
	if len(visibleContentLines.lines) == 0 {
		return ""
	}

	// if selection is disabled, numerator should be item index of bottom visible line
	if !m.navigation.SelectionEnabled {
		numerator = visibleContentLines.itemIndexes[len(visibleContentLines.itemIndexes)-1] + 1
		if m.config.WrapText && numerator == denominator && !m.isScrolledToBottom() {
			// if wrapped && bottom visible line is max item index, but actually not fully scrolled to bottom, show 99%
			return fmt.Sprintf("99%% (%d/%d)", numerator, denominator)
		}
	}

	percentScrolled := percent(numerator, denominator)
	footerString := fmt.Sprintf("%d%% (%d/%d)", percentScrolled, numerator, denominator)

	footerBuffer := linebuffer.New(footerString)
	f, _ := footerBuffer.Take(0, m.display.Bounds.Width, m.config.ContinuationIndicator, linebuffer.HighlightData{}, lipgloss.NewStyle())
	return m.display.Styles.FooterStyle.Render(f)
}

func (m *Model[T]) getLineContinuationIndicator() string {
	if m.config.WrapText {
		return ""
	}
	return m.config.ContinuationIndicator
}

func (m *Model[T]) isScrolledToBottom() bool {
	maxItemIdx, maxTopItemLineOffset := m.maxItemIdxAndMaxTopLineOffset()
	if m.display.TopItemIdx > maxItemIdx {
		return true
	}
	if m.display.TopItemIdx == maxItemIdx {
		return m.display.TopItemLineOffset >= maxTopItemLineOffset
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
	visibleContentLines := m.getVisibleContentLines()
	numLinesSelectionInView := 0
	numLinesAboveSelection := 0
	assignedNumLinesAboveSelection := false
	for i := range visibleContentLines.itemIndexes {
		if visibleContentLines.itemIndexes[i] == m.content.GetSelectedIdx() {
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
	lenAllItems := m.content.NumItems()
	if lenAllItems == 0 {
		return 0, 0
	}
	if !m.config.WrapText {
		return max(0, lenAllItems-m.getNumContentLines()), 0
	}
	// wrapped
	maxTopItemIdx, maxTopItemLineOffset := lenAllItems-1, 0
	nLinesLastItem := m.numLinesForItem(lenAllItems - 1)
	if m.getNumContentLines() <= nLinesLastItem {
		// same item, just change offset
		maxTopItemLineOffset = nLinesLastItem - m.getNumContentLines()
	} else {
		// take lines from items until scrolled up desired amount
		n := m.getNumContentLines() - nLinesLastItem
		for n > 0 {
			maxTopItemIdx--
			if maxTopItemIdx < 0 {
				// scrolled up past top - stay at top
				maxTopItemIdx = 0
				maxTopItemLineOffset = 0
				break
			}
			numLinesInTopItem := m.numLinesForItem(maxTopItemIdx)
			for i := range numLinesInTopItem {
				n--
				if n == 0 {
					maxTopItemLineOffset = numLinesInTopItem - (i + 1)
					break
				}
			}
		}
	}
	return max(0, maxTopItemIdx), max(0, maxTopItemLineOffset)
}

func (m *Model[T]) getNumVisibleItems() int {
	if !m.config.WrapText {
		return m.getNumContentLines()
	}
	visibleContentLines := m.getVisibleContentLines()
	// return distinct number of items
	itemIndexSet := make(map[int]struct{})
	for _, i := range visibleContentLines.itemIndexes {
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
			builder.WriteString(m.display.Styles.SelectedItemStyle.Render(section))
		}
		if i < len(split)-1 && i < len(matches) {
			builder.WriteString(matches[i])
		}
	}
	return builder.String()
}

func toLineBuffers(lines []string) []linebuffer.LineBufferer {
	res := make([]linebuffer.LineBufferer, len(lines))
	for i, line := range lines {
		res[i] = linebuffer.New(line)
	}
	return res
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

func safeSliceFromIdx(s []string, i int) []string {
	if i < 0 {
		return s
	}
	if i > len(s) {
		return []string{}
	}
	return s[i:]
}
