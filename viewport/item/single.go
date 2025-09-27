package item

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// SingleItem provides functionality to get sequential strings of a specified terminal cell width, accounting
// for the ansi escape codes styling the line.
type SingleItem struct {
	line                 string     // underlying string with ansi codes. utf-8 encoded bytes
	lineNoAnsi           string     // line without ansi codes. utf-8 encoded bytes
	lineNoAnsiRuneWidths []uint8    // packed terminal cell widths, 4 widths per byte (2 bits each)
	ansiCodeIndexes      [][]uint32 // slice of startByte, endByte indexes of ansi codes
	numNoAnsiRunes       int        // number of runes in lineNoAnsi
	totalWidth           int        // total width in terminal cells

	sparsity                        int      // interval for which to store cumulative cell width
	sparseRuneIdxToNoAnsiByteOffset []uint32 // rune idx to byte offset of lineNoAnsi, stored every sparsity runes
	sparseLineNoAnsiCumRuneWidths   []uint32 // cumulative terminal cell width, stored every sparsity runes
}

// type assertion that SingleItem implements Item
var _ Item = SingleItem{}

// type assertion that *SingleItem implements Item
var _ Item = (*SingleItem)(nil)

// NewItem creates a new SingleItem from the given string.
func NewItem(line string) SingleItem {
	if len(line) <= 0 {
		return SingleItem{line: line}
	}

	// keep sparsity 1 for short lines
	sparsity := 1
	if len(line) > 1000 {
		sparsity = 10 // tradeoff between memory usage and CPU. 10 seems to be a good balance
	}

	item := SingleItem{
		line:     line,
		sparsity: sparsity,
	}

	item.ansiCodeIndexes = findAnsiByteRanges(line)

	if len(item.ansiCodeIndexes) > 0 {
		totalLen := len(line)
		for _, r := range item.ansiCodeIndexes {
			totalLen -= int(r[1] - r[0])
		}

		noAnsiBytes := make([]byte, 0, totalLen)
		lastPos := 0
		for _, r := range item.ansiCodeIndexes {
			noAnsiBytes = append(noAnsiBytes, line[lastPos:int(r[0])]...)
			lastPos = int(r[1])
		}
		noAnsiBytes = append(noAnsiBytes, line[lastPos:]...)
		item.lineNoAnsi = string(noAnsiBytes)
	} else {
		item.lineNoAnsi = line
	}

	numRunes := utf8.RuneCountInString(item.lineNoAnsi)

	// calculate size needed for sparse cumulative widths
	sparseLen := (numRunes + item.sparsity - 1) / item.sparsity
	item.sparseRuneIdxToNoAnsiByteOffset = make([]uint32, sparseLen)
	item.sparseLineNoAnsiCumRuneWidths = make([]uint32, sparseLen)

	// calculate size needed for packed rune widths (4 widths per byte)
	packedLen := (numRunes + 3) / 4
	item.lineNoAnsiRuneWidths = make([]uint8, packedLen)

	var currentOffset uint32
	var cumWidth uint32
	runeIdx := 0
	for byteOffset := 0; byteOffset < len(item.lineNoAnsi); {
		r, runeNumBytes := utf8.DecodeRuneInString(item.lineNoAnsi[byteOffset:])
		rw := runewidth.RuneWidth(r)
		width := clampIntToUint8(rw)

		// pack 4 widths per byte (2 bits each)
		packedIdx := runeIdx / 4
		bitPos := (runeIdx % 4) * 2
		// clear the 2 bits at the position and set the new width
		item.lineNoAnsiRuneWidths[packedIdx] &= ^(uint8(3) << bitPos)
		item.lineNoAnsiRuneWidths[packedIdx] |= width << bitPos

		cumWidth += uint32(width)
		if runeIdx%item.sparsity == 0 {
			item.sparseRuneIdxToNoAnsiByteOffset[runeIdx/item.sparsity] = currentOffset
			item.sparseLineNoAnsiCumRuneWidths[runeIdx/item.sparsity] = cumWidth
		}
		if runeIdx == numRunes-1 {
			item.totalWidth = int(cumWidth)
		}
		currentOffset += clampIntToUint32(runeNumBytes)
		runeIdx++
		byteOffset += runeNumBytes
	}
	item.numNoAnsiRunes = runeIdx

	return item
}

// Width returns the total width in terminal cells.
func (l SingleItem) Width() int {
	if len(l.line) == 0 {
		return 0
	}
	return l.totalWidth
}

// Content returns the underlying string content
func (l SingleItem) Content() string {
	return l.line
}

// ContentNoAnsi returns the underlying string content without ANSI escape codes
func (l SingleItem) ContentNoAnsi() string {
	return l.lineNoAnsi
}

// Take returns a substring of the item that fits within the specified width
func (l SingleItem) Take(
	widthToLeft,
	takeWidth int,
	continuation string,
	highlights []Highlight,
) (string, int) {
	if widthToLeft < 0 {
		widthToLeft = 0
	}

	widthToLeft = min(widthToLeft, l.Width())
	startRuneIdx := l.findRuneIndexWithWidthToLeft(widthToLeft)

	if startRuneIdx >= l.numNoAnsiRunes || takeWidth == 0 {
		return "", 0
	}

	var result strings.Builder
	remainingWidth := takeWidth
	leftRuneIdx := startRuneIdx
	startByteOffset := l.getByteOffsetAtRuneIdx(startRuneIdx)

	runesWritten := 0
	for ; remainingWidth > 0 && leftRuneIdx < l.numNoAnsiRunes; leftRuneIdx++ {
		r := l.runeAt(leftRuneIdx)
		runeWidth := l.getRuneWidth(leftRuneIdx)
		if int(runeWidth) > remainingWidth {
			break
		}

		result.WriteRune(r)
		runesWritten++
		remainingWidth -= int(runeWidth)
	}

	// if only zero-width runes were written, return ""
	for i := 0; i < runesWritten; i++ {
		if runewidth.RuneWidth(l.runeAt(startRuneIdx+i)) > 0 {
			break
		}
		if i == runesWritten-1 {
			return "", 0
		}
	}

	// write the subsequent zero-width runes, e.g. the accent on an 'e'
	if result.Len() > 0 {
		for ; leftRuneIdx < l.numNoAnsiRunes; leftRuneIdx++ {
			r := l.runeAt(leftRuneIdx)
			if runewidth.RuneWidth(r) == 0 {
				result.WriteRune(r)
			} else {
				break
			}
		}
	}

	res := result.String()

	// reapply original styling
	if len(l.ansiCodeIndexes) > 0 {
		res = reapplyAnsi(l.line, res, int(startByteOffset), l.ansiCodeIndexes)
	}

	// highlight the desired string
	var endByteOffset int
	if leftRuneIdx < l.numNoAnsiRunes {
		endByteOffset = int(l.getByteOffsetAtRuneIdx(leftRuneIdx))
	} else {
		endByteOffset = len(l.lineNoAnsi)
	}
	res = highlightString(
		res,
		highlights,
		int(startByteOffset),
		endByteOffset,
	)

	// apply left/right line continuation indicators
	if len(continuation) > 0 && (startRuneIdx > 0 || leftRuneIdx < l.numNoAnsiRunes) {
		continuationRunes := []rune(continuation)

		// if more runes to the left of the result, replace start runes with continuation indicator
		if startRuneIdx > 0 {
			res = replaceStartWithContinuation(res, continuationRunes)
		}

		// if more runes to the right, replace final runes in result with continuation indicator
		if leftRuneIdx < l.numNoAnsiRunes {
			res = replaceEndWithContinuation(res, continuationRunes)
		}
	}

	res = removeEmptyAnsiSequences(res)
	return res, takeWidth - remainingWidth
}

// NumWrappedLines returns the number of wrapped lines given a wrap width
func (l SingleItem) NumWrappedLines(wrapWidth int) int {
	if wrapWidth <= 0 {
		return 0
	} else if l.totalWidth == 0 {
		return 1
	}
	return (l.totalWidth + wrapWidth - 1) / wrapWidth
}

// Repr returns a string representation for debugging.
func (l SingleItem) repr() string {
	return fmt.Sprintf("Item(%q)", l.line)
}

// runeAt decodes the desired rune from the lineNoAnsi string
// it serves as a memory-saving technique compared to storing all the runes in a slice
func (l SingleItem) runeAt(runeIdx int) rune {
	if runeIdx < 0 || runeIdx >= l.numNoAnsiRunes {
		return -1
	}
	start := l.getByteOffsetAtRuneIdx(runeIdx)
	var end uint32
	if runeIdx+1 >= l.numNoAnsiRunes {
		end = clampIntToUint32(len(l.lineNoAnsi))
	} else {
		end = l.getByteOffsetAtRuneIdx(runeIdx + 1)
	}
	r, _ := utf8.DecodeRuneInString(l.lineNoAnsi[start:end])
	return r
}

func (l SingleItem) getByteOffsetAtRuneIdx(runeIdx int) uint32 {
	if runeIdx < 0 {
		panic("runeIdx must be greater or equal to 0")
	}
	if runeIdx == 0 || len(l.line) == 0 || l.sparsity == 0 {
		return 0
	}
	if runeIdx >= l.numNoAnsiRunes {
		panic("rune index greater than num runes")
	}

	// get the last stored byte offset before this index
	sparseIdx := runeIdx / l.sparsity
	baseRuneIdx := sparseIdx * l.sparsity

	if baseRuneIdx == runeIdx {
		return l.sparseRuneIdxToNoAnsiByteOffset[sparseIdx]
	}

	currRuneIdx := baseRuneIdx
	byteOffset := l.sparseRuneIdxToNoAnsiByteOffset[sparseIdx]
	for ; currRuneIdx != runeIdx; currRuneIdx++ {
		_, nBytes := utf8.DecodeRuneInString(l.lineNoAnsi[byteOffset:])
		byteOffset += clampIntToUint32(nBytes)
	}
	return byteOffset
}

// getRuneWidth extracts the width of a rune from the packed array
func (l SingleItem) getRuneWidth(runeIdx int) uint8 {
	if runeIdx < 0 || runeIdx >= l.numNoAnsiRunes {
		return 0
	}

	packedIdx := runeIdx / 4
	bitPos := (runeIdx % 4) * 2
	return (l.lineNoAnsiRuneWidths[packedIdx] >> bitPos) & 3
}

func (l SingleItem) getCumulativeWidthAtRuneIdx(runeIdx int) uint32 {
	if runeIdx < 0 {
		return 0
	}
	if runeIdx >= l.numNoAnsiRunes {
		panic("runeIdx greater than num runes")
	}

	// get the last stored cumulative width before this index
	sparseIdx := runeIdx / l.sparsity
	baseRuneIdx := sparseIdx * l.sparsity

	if baseRuneIdx == runeIdx {
		return l.sparseLineNoAnsiCumRuneWidths[sparseIdx]
	}

	// sum the widths from the last stored point to our target index
	var additionalWidth uint32
	for i := baseRuneIdx + 1; i <= runeIdx; i++ {
		additionalWidth += uint32(l.getRuneWidth(i))
	}

	return l.sparseLineNoAnsiCumRuneWidths[sparseIdx] + additionalWidth
}

// findRuneIndexWithWidthToLeft returns the index of the rune that has the input width to the left of it
func (l SingleItem) findRuneIndexWithWidthToLeft(widthToLeft int) int {
	if widthToLeft < 0 {
		panic("widthToLeft less than 0")
	}
	if widthToLeft == 0 || l.numNoAnsiRunes == 0 {
		return 0
	}
	if widthToLeft > l.Width() {
		panic("widthToLeft greater than total width")
	}

	left, right := 0, l.numNoAnsiRunes-1
	widthToLeftUint32 := clampIntToUint32(widthToLeft)
	if l.getCumulativeWidthAtRuneIdx(right) < widthToLeftUint32 {
		return l.numNoAnsiRunes
	}

	for left < right {
		mid := left + (right-left)/2
		if l.getCumulativeWidthAtRuneIdx(mid) >= widthToLeftUint32 {
			right = mid
		} else {
			left = mid + 1
		}
	}

	// skip over zero-width runes
	w := l.getCumulativeWidthAtRuneIdx(left)
	nextLeft := left + 1
	for nextLeft < l.numNoAnsiRunes && l.getCumulativeWidthAtRuneIdx(nextLeft) == w {
		left = nextLeft
		nextLeft++
	}

	return left + 1
}

// ExtractExactMatches extracts exact matches from the item's content without ANSI codes
func (l SingleItem) ExtractExactMatches(exactMatch string) []ByteRange {
	return extractExactMatches(l.lineNoAnsi, exactMatch)
}

// ExtractRegexMatches extracts regex matches from the item's content without ANSI codes
func (l SingleItem) ExtractRegexMatches(regex *regexp.Regexp) []ByteRange {
	return extractRegexMatches(l.lineNoAnsi, regex)
}
