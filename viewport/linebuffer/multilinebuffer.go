package linebuffer

import (
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

// PlainContent returns the concatenated plain content without ANSI codes.
func (m MultiLineBuffer) PlainContent() string {
	return m.concatenatedLineNoAnsi()
}

// Take returns a string from the buffer
func (m MultiLineBuffer) Take(
	widthToLeft,
	takeWidth int,
	continuation string,
) (string, TakenMetaData) {
	if len(m.buffers) == 0 {
		return "", TakenMetaData{Width: 0, StartByte: 0, EndByte: 0}
	}
	if len(m.buffers) == 1 {
		return m.buffers[0].Take(widthToLeft, takeWidth, continuation)
	}
	if widthToLeft >= m.totalWidth {
		return "", TakenMetaData{Width: 0, StartByte: 0, EndByte: 0}
	}

	// find which buffer contains our start position
	skippedWidth := 0
	skippedBytes := 0
	firstBufferIdx := 0
	firstByteIdx := 0
	startWidthFirstBuffer := widthToLeft

	for i := range m.buffers {
		bufWidth := m.buffers[i].Width()
		if skippedWidth+bufWidth > widthToLeft {
			firstBufferIdx = i
			startWidthFirstBuffer = widthToLeft - skippedWidth

			runeIdx := m.buffers[i].findRuneIndexWithWidthToLeft(startWidthFirstBuffer)
			var firstBufferByteIdx int
			if runeIdx < m.buffers[i].numNoAnsiRunes {
				firstBufferByteIdx = int(m.buffers[i].getByteOffsetAtRuneIdx(runeIdx))
			} else {
				firstBufferByteIdx = len(m.buffers[i].line)
			}
			firstByteIdx = skippedBytes + firstBufferByteIdx
			break
		}
		skippedWidth += bufWidth
		skippedBytes += len(m.buffers[i].lineNoAnsi)
		startWidthFirstBuffer -= bufWidth
	}

	// take from first buffer
	res, firstMetaData := m.buffers[firstBufferIdx].Take(startWidthFirstBuffer, takeWidth, "")
	remainingTotalWidth := takeWidth - firstMetaData.Width

	// if we have more width to take and more buffers available, continue
	currentBufferIdx := firstBufferIdx + 1
	for remainingTotalWidth > 0 && currentBufferIdx < len(m.buffers) {
		nextPart, nextMetaData := m.buffers[currentBufferIdx].Take(0, remainingTotalWidth, "")
		if nextMetaData.Width == 0 {
			break
		}
		res += nextPart
		remainingTotalWidth -= nextMetaData.Width
		currentBufferIdx++
	}

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
	return res, TakenMetaData{
		Width:     takeWidth - remainingTotalWidth,
		StartByte: firstByteIdx,
		EndByte:   firstByteIdx + len(stripAnsi(res)),
	}
}

// WrappedLines returns the content broken into lines that fit within the specified width.
// WrappedLinesWithoutHighlights returns the content broken into lines without applying highlights.
// This is optimized for layout calculations that don't need styling.
func (m MultiLineBuffer) WrappedLinesWithoutHighlights(
	width int,
	maxLinesEachEnd int,
) []string {
	if width <= 0 {
		return []string{}
	}
	if len(m.buffers) == 0 {
		return []string{}
	}
	if len(m.buffers) == 1 {
		return m.buffers[0].WrappedLinesWithoutHighlights(width, maxLinesEachEnd)
	}

	totalLines := (m.totalWidth + width - 1) / width
	if totalLines == 0 {
		return []string{""}
	}

	return getWrappedLinesWithoutHighlights(
		m,
		totalLines,
		width,
		maxLinesEachEnd,
	)
}

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
	var builder strings.Builder
	for i := range m.buffers {
		builder.WriteString(m.buffers[i].lineNoAnsi)
	}
	return builder.String()
}
