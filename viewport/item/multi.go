package item

import (
	"strings"
)

// MultiItem implements Item by wrapping multiple SingleItem's without extra memory allocation
type MultiItem struct {
	items         []SingleItem
	totalWidth    int    // cached total width across all items
	contentNoAnsi string // cached concatenated content without ANSI escape codes
}

// type assertion that MultiItem implements Item
var _ Item = MultiItem{}

// type assertion that *MultiItem implements Item
var _ Item = (*MultiItem)(nil)

// NewMulti creates a new MultiItem from the given items
func NewMulti(items ...SingleItem) MultiItem {
	if len(items) == 0 {
		return MultiItem{}
	}

	totalWidth := 0
	for _, item := range items {
		totalWidth += item.Width()
	}

	return MultiItem{
		items:      items,
		totalWidth: totalWidth,
	}
}

// Width returns the total width across all items.
func (m MultiItem) Width() int {
	return m.totalWidth
}

// Content returns the concatenated content of all items.
func (m MultiItem) Content() string {
	if len(m.items) == 0 {
		return ""
	}

	if len(m.items) == 1 {
		return m.items[0].Content()
	}

	totalLen := 0
	for _, items := range m.items {
		totalLen += len(items.Content())
	}

	var builder strings.Builder
	builder.Grow(totalLen)

	for _, item := range m.items {
		builder.WriteString(item.Content())
	}

	return builder.String()
}

// ContentNoAnsi returns the concatenated content of all items without ANSI escape codes that style the string
func (m MultiItem) ContentNoAnsi() string {
	if m.contentNoAnsi != "" {
		return m.contentNoAnsi
	}

	if len(m.items) == 0 {
		return ""
	}

	if len(m.items) == 1 {
		m.contentNoAnsi = m.items[0].ContentNoAnsi()
		return m.contentNoAnsi
	}

	// make a single allocation for the concatenated string
	totalLen := 0
	for _, items := range m.items {
		totalLen += len(items.ContentNoAnsi())

	}
	var builder strings.Builder
	builder.Grow(totalLen)
	for _, item := range m.items {
		builder.WriteString(item.ContentNoAnsi())
	}
	m.contentNoAnsi = builder.String()
	return m.contentNoAnsi
}

// Take returns a substring of the item that fits within the specified width
func (m MultiItem) Take(
	widthToLeft,
	takeWidth int,
	continuation string,
	highlights []Highlight,
) (string, int) {
	if len(m.items) == 0 {
		return "", 0
	}
	if len(m.items) == 1 {
		return m.items[0].Take(widthToLeft, takeWidth, continuation, highlights)
	}
	if widthToLeft >= m.totalWidth {
		return "", 0
	}

	// find which item contains our start position
	skippedWidth := 0
	skippedBytes := 0
	firstItemIdx := 0
	firstByteIdx := 0
	startWidthFirstItem := widthToLeft

	for i := range m.items {
		itemWidth := m.items[i].Width()
		if skippedWidth+itemWidth > widthToLeft {
			firstItemIdx = i
			startWidthFirstItem = widthToLeft - skippedWidth

			runeIdx := m.items[i].findRuneIndexWithWidthToLeft(startWidthFirstItem)
			var firstItemByteIdx int
			if runeIdx < m.items[i].numNoAnsiRunes {
				firstItemByteIdx = int(m.items[i].getByteOffsetAtRuneIdx(runeIdx))
			} else {
				firstItemByteIdx = len(m.items[i].line)
			}
			firstByteIdx = skippedBytes + firstItemByteIdx
			break
		}
		skippedWidth += itemWidth
		skippedBytes += len(m.items[i].lineNoAnsi)
		startWidthFirstItem -= itemWidth
	}

	// take from first item
	res, takenWidth := m.items[firstItemIdx].Take(startWidthFirstItem, takeWidth, "", []Highlight{})
	remainingTotalWidth := takeWidth - takenWidth

	// if we have more width to take and more items available, continue
	currentItemIdx := firstItemIdx + 1
	for remainingTotalWidth > 0 && currentItemIdx < len(m.items) {
		nextPart, partWidth := m.items[currentItemIdx].Take(0, remainingTotalWidth, "", []Highlight{})
		if partWidth == 0 {
			break
		}
		res += nextPart
		remainingTotalWidth -= partWidth
		currentItemIdx++
	}

	res = highlightString(
		res,
		highlights,
		firstByteIdx,
		firstByteIdx+len(stripAnsi(res)),
	)

	// apply continuation indicators if needed
	if len(continuation) > 0 {
		contentToLeft := widthToLeft > 0
		contentToRight := m.totalWidth-widthToLeft > takeWidth-remainingTotalWidth
		if contentToLeft || contentToRight {
			continuationRunes := []rune(continuation)
			if contentToLeft {
				res = replaceStartWithContinuation(res, continuationRunes)
			}
			if contentToRight {
				res = replaceEndWithContinuation(res, continuationRunes)
			}
		}
	}

	res = removeEmptyAnsiSequences(res)
	return res, takeWidth - remainingTotalWidth
}

// NumWrappedLines returns the number of wrapped lines given a wrap width
func (m MultiItem) NumWrappedLines(wrapWidth int) int {
	if wrapWidth <= 0 {
		return 0
	} else if m.totalWidth == 0 {
		return 1
	}
	return (m.totalWidth + wrapWidth - 1) / wrapWidth
}

// Repr returns a string representation of the MultiItem for debugging.
func (m MultiItem) repr() string {
	v := "Multi("
	for i := range m.items {
		if i > 0 {
			v += ", "
		}
		v += m.items[i].repr()
	}
	v += ")"
	return v
}
