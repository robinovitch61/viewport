package linebuffer

import (
	"regexp"
	"strings"
)

// MultiLineBuffer implements LineBufferer by wrapping multiple LineBuffers without extra memory allocation
type MultiLineBuffer struct {
	buffers    []LineBuffer
	totalWidth int // cached total width across all buffers
}

// type assertion that MultiLineBuffer implements LineBufferer
var _ LineBufferer = MultiLineBuffer{}

// type assertion that *MultiLineBuffer implements LineBufferer
var _ LineBufferer = (*MultiLineBuffer)(nil)

// NewMulti creates a new MultiLineBuffer from the given LineBuffers.
func NewMulti(buffers ...LineBuffer) MultiLineBuffer {
	if len(buffers) == 0 {
		return MultiLineBuffer{}
	}

	totalWidth := 0
	for _, buf := range buffers {
		totalWidth += buf.Width()
	}

	return MultiLineBuffer{
		buffers:    buffers,
		totalWidth: totalWidth,
	}
}

// Width returns the total width across all buffers.
func (m MultiLineBuffer) Width() int {
	return m.totalWidth
}

// Content returns the concatenated content of all buffers.
func (m MultiLineBuffer) Content() string {
	if len(m.buffers) == 0 {
		return ""
	}

	if len(m.buffers) == 1 {
		return m.buffers[0].Content()
	}

	totalLen := 0
	for _, buf := range m.buffers {
		totalLen += len(buf.Content())
	}

	var builder strings.Builder
	builder.Grow(totalLen)

	for _, buf := range m.buffers {
		builder.WriteString(buf.Content())
	}

	return builder.String()
}

// Take returns a string from the buffer
func (m MultiLineBuffer) Take(
	widthToLeft,
	takeWidth int,
	continuation string,
	highlights []Highlight,
) (string, int) {
	if len(m.buffers) == 0 {
		return "", 0
	}
	if len(m.buffers) == 1 {
		return m.buffers[0].Take(widthToLeft, takeWidth, continuation, highlights)
	}
	if widthToLeft >= m.totalWidth {
		return "", 0
	}

	// find which buffer contains our start position
	skippedWidth := 0
	firstBufferIdx := 0
	startWidthFirstBuffer := widthToLeft

	for i := range m.buffers {
		bufWidth := m.buffers[i].Width()
		if skippedWidth+bufWidth > widthToLeft {
			firstBufferIdx = i
			startWidthFirstBuffer = widthToLeft - skippedWidth
			break
		}
		skippedWidth += bufWidth
		startWidthFirstBuffer -= bufWidth
	}

	// get content before our start position for highlight context
	contextSize := calculateContextSize(highlights)
	leftContext := getBytesLeftOfWidth(contextSize, m.buffers, firstBufferIdx, startWidthFirstBuffer)

	// take from first buffer
	res, takenWidth := m.buffers[firstBufferIdx].Take(startWidthFirstBuffer, takeWidth, "", []Highlight{})
	remainingTotalWidth := takeWidth - takenWidth
	remainingBufferWidth := m.buffers[firstBufferIdx].Width() - takenWidth

	// if we have more width to take and more buffers available, continue
	currentBufferIdx := firstBufferIdx + 1
	for remainingTotalWidth > 0 && currentBufferIdx < len(m.buffers) {
		nextPart, partWidth := m.buffers[currentBufferIdx].Take(0, remainingTotalWidth, "", []Highlight{})
		remainingBufferWidth = m.buffers[currentBufferIdx].Width() - partWidth
		if partWidth == 0 {
			break
		}
		res += nextPart
		remainingTotalWidth -= partWidth
		currentBufferIdx++
	}

	// get content after our result for highlight context
	currentBufferIdx--
	rightContext := getBytesRightOfWidth(contextSize, m.buffers, currentBufferIdx, remainingBufferWidth)

	// highlight the desired string
	resNoAnsi := stripAnsi(res)
	lineNoAnsi := leftContext + resNoAnsi + rightContext
	res = highlightString(
		res,
		highlights,
		lineNoAnsi,
		len(leftContext), // TODO LEO: this could be too small! Could add len(leftContext) to highlights or adjust lineNoAnsi to be bigger or something
		len(leftContext)+len(resNoAnsi),
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

// WrappedLines returns the content broken into lines that fit within the specified width.
func (m MultiLineBuffer) WrappedLines(
	width int,
	maxLinesEachEnd int,
	highlights []Highlight,
) []string {
	if width <= 0 {
		return []string{}
	}
	if len(m.buffers) == 0 {
		return []string{}
	}
	if len(m.buffers) == 1 {
		return m.buffers[0].WrappedLines(width, maxLinesEachEnd, highlights)
	}

	totalLines := (m.totalWidth + width - 1) / width
	if totalLines == 0 {
		return []string{""}
	}

	return getWrappedLines(
		m,
		totalLines,
		width,
		maxLinesEachEnd,
		highlights,
	)
}

// Matches returns true if the content contains the specified string.
func (m MultiLineBuffer) Matches(s string) bool {
	return strings.Contains(m.concatenatedLineNoAnsi(), s)
}

// MatchesRegex returns true if the content matches the specified regular expression.
func (m MultiLineBuffer) MatchesRegex(r regexp.Regexp) bool {
	return r.MatchString(m.concatenatedLineNoAnsi())
}

// Repr returns a string representation of the MultiLineBuffer for debugging.
func (m MultiLineBuffer) Repr() string {
	v := "Multi("
	for i := range m.buffers {
		if i > 0 {
			v += ", "
		}
		v += m.buffers[i].Repr()
	}
	v += ")"
	return v
}

func (m MultiLineBuffer) concatenatedLineNoAnsi() string {
	// this isn't super efficient - could potentially consider trying to do string matches across
	// LineBuffer boundaries in the future without allocating new strings every time, but that's tricky
	var builder strings.Builder
	for i := range m.buffers {
		builder.WriteString(m.buffers[i].lineNoAnsi)
	}
	return builder.String()
}

// calculateContextSize determines how much context to gather around highlights
func calculateContextSize(highlights []Highlight) int {
	if len(highlights) == 0 {
		return 0
	}

	maxHighlightLength := 0
	for _, highlight := range highlights {
		highlightLength := highlight.EndByteOffset - highlight.StartByteOffset
		if highlightLength > maxHighlightLength {
			maxHighlightLength = highlightLength
		}
	}
	return maxHighlightLength * 2
}
