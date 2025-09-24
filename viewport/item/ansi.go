package item

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// highlightRange represents a highlight with start/end positions and style
type highlightRange struct {
	startByte int
	endByte   int
	style     lipgloss.Style
}

// RST is the ansi escape sequence for resetting styles
// note that in future charm library versions this may change to "\x1b[m"
const RST = "\x1b[0m"

// reapplyAnsi reconstructs ANSI escape sequences in a truncated string based on their positions in the original.
// It ensures that any active text formatting (colors, styles) from the original string is correctly maintained
// in the truncated output, and adds proper reset codes where needed.
//
// Parameters:
//   - original: the source string containing ANSI escape sequences
//   - truncated: the truncated version of the string, without ANSI sequences
//   - truncByteOffset: byte offset in the original string where truncation started
//   - ansiCodeIndexes: pairs of start/end byte positions of ANSI codes in the original string
//
// Returns a string with ANSI escape sequences reapplied at appropriate positions,
// maintaining the original text formatting while preserving proper UTF-8 encoding.
func reapplyAnsi(original, truncated string, truncByteOffset int, ansiCodeIndexes [][]uint32) string {
	var result strings.Builder
	result.Grow(len(truncated))
	var lenAnsiAdded int
	isReset := true

	for i := 0; i < len(truncated); {
		// collect all ansi codes that should be applied immediately before the current runes
		var ansisToAdd []string
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := int(candidateAnsi[0]), int(candidateAnsi[1])
			originalByteIdx := truncByteOffset + i + lenAnsiAdded
			if codeStart <= originalByteIdx {
				code := original[codeStart:codeEnd]
				isReset = code == RST
				ansisToAdd = append(ansisToAdd, code)
				lenAnsiAdded += codeEnd - codeStart
				ansiCodeIndexes = ansiCodeIndexes[1:]
			} else {
				break
			}
		}

		for _, ansi := range simplifyAnsiCodes(ansisToAdd) {
			result.WriteString(ansi)
		}

		// add the bytes of the current rune
		_, size := utf8.DecodeRuneInString(truncated[i:])
		result.WriteString(truncated[i : i+size])
		i += size
	}

	if !isReset {
		result.WriteString(RST)
	}
	return result.String()
}

// getNonAnsiBytes extracts a substring of specified length from the input string, excluding ANSI escape sequences.
// It reads from the given start position until it has collected the requested number of non-ANSI bytes.
//
// Parameters:
//   - s: The input string that may contain ANSI escape sequences
//   - startIdx: The byte position in the input to start reading from
//   - numBytes: The number of non-ANSI bytes to collect
//
// Returns a string containing bytesToExtract bytes of the input with ANSI sequences removed. If the input text ends
// before collecting bytesToExtract bytes, returns all available non-ANSI bytes.
func getNonAnsiBytes(s string, startIdx, numBytes int) string {
	var result strings.Builder
	currentPos := startIdx
	bytesCollected := 0
	for currentPos < len(s) && bytesCollected < numBytes {
		if strings.HasPrefix(s[currentPos:], "\x1b[") {
			escEnd := currentPos + strings.Index(s[currentPos:], "m") + 1
			currentPos = escEnd
			continue
		}
		result.WriteByte(s[currentPos])
		bytesCollected++
		currentPos++
	}
	return result.String()
}

// highlightLine highlights a string in a line that potentially has ansi codes in it without disrupting them
// start and end are the byte offsets for which highlighting is considered in the line, not counting ansi codes
func highlightLine(styledLine, highlight string, highlightStyle lipgloss.Style, startByte, endByte int) string {
	if styledLine == "" || highlight == "" {
		return styledLine
	}

	renderedHighlight := highlightStyle.Render(highlight)
	var result strings.Builder
	var activeStyles []string
	inAnsi := false
	nonAnsiBytes := 0

	i := 0
	for i < len(styledLine) {
		if strings.HasPrefix(styledLine[i:], "\x1b[") {
			// found start of ansi
			inAnsi = true
			ansiLen := strings.Index(styledLine[i:], "m")
			if ansiLen != -1 {
				escEnd := i + ansiLen + 1
				ansi := styledLine[i:escEnd]
				if ansi == RST {
					activeStyles = []string{} // reset
				} else {
					activeStyles = append(activeStyles, ansi) // add new active style
				}
				result.WriteString(ansi)
				i = escEnd
				inAnsi = false
				continue
			}
		}

		// check if current position starts a highlight match
		if !inAnsi && nonAnsiBytes >= startByte && nonAnsiBytes < endByte {
			textToCheck := getNonAnsiBytes(styledLine, i, len(highlight))
			if textToCheck == highlight {
				// reset current styles, if any
				if len(activeStyles) > 0 {
					result.WriteString(RST)
				}
				// apply highlight
				result.WriteString(renderedHighlight)
				// restore previous styles, if any
				if len(activeStyles) > 0 {
					for j := range activeStyles {
						result.WriteString(activeStyles[j])
					}
				}

				// skip to end of matched text
				count := 0
				for count < len(highlight) {
					if strings.HasPrefix(styledLine[i:], "\x1b[") {
						escEnd := i + strings.Index(styledLine[i:], "m") + 1
						result.WriteString(styledLine[i:escEnd])
						i = escEnd
						continue
					}
					i++
					count++
					nonAnsiBytes++
				}
				continue
			}
		}
		result.WriteByte(styledLine[i])
		nonAnsiBytes++
		i++
	}
	return removeEmptyAnsiSequences(result.String())
}

// highlightString applies highlighting to a segment of text while handling cases where the highlight
// might overflow the segment boundaries. It preserves any existing ANSI styling in the segment.
//
// Parameters:
//   - styledSegment: the text segment to highlight, which may contain ANSI codes
//   - highlights: a list of Highlight structs defining the styledLine byte offsets and styles to apply
//   - plainLineSegmentStartByte: byte offset where styledSegment starts in full line without ansi codes
//   - plainLineSegmentEndByte: byte offset where styledSegment ends in full line without ansi codes
//
// Returns the segment with highlighting applied, preserving original ANSI codes.
func highlightString(
	styledSegment string,
	highlights []Highlight,
	plainLineSegmentStartByte int,
	plainLineSegmentEndByte int,
) string {
	if len(highlights) == 0 {
		return styledSegment
	}

	var applicableHighlights []highlightRange
	for _, highlight := range highlights {
		if highlight.Match.StartByteUnstyledContent < plainLineSegmentEndByte && highlight.Match.EndByteUnstyledContent > plainLineSegmentStartByte {
			startByte := max(highlight.Match.StartByteUnstyledContent, plainLineSegmentStartByte) - plainLineSegmentStartByte
			endByte := min(highlight.Match.EndByteUnstyledContent, plainLineSegmentEndByte) - plainLineSegmentStartByte
			applicableHighlights = append(applicableHighlights, highlightRange{
				startByte: startByte,
				endByte:   endByte,
				style:     highlight.Style,
			})
		}
	}

	if len(applicableHighlights) == 0 {
		return styledSegment
	}

	// sort highlights by start position
	for i := 0; i < len(applicableHighlights); i++ {
		for j := i + 1; j < len(applicableHighlights); j++ {
			if applicableHighlights[j].startByte < applicableHighlights[i].startByte {
				applicableHighlights[i], applicableHighlights[j] = applicableHighlights[j], applicableHighlights[i]
			}
		}
	}

	var result strings.Builder
	// pre-allocation based on highlight density (~50 bytes per highlight for styling)
	estimatedSize := len(styledSegment) + len(applicableHighlights)*50
	result.Grow(estimatedSize)

	var activeStyles []string
	nonAnsiBytes := 0
	highlightIdx := 0
	inAnsi := false

	i := 0
	for i < len(styledSegment) {
		// handle ansi sequences
		if strings.HasPrefix(styledSegment[i:], "\x1b[") {
			inAnsi = true
			ansiLen := strings.Index(styledSegment[i:], "m")
			if ansiLen != -1 {
				escEnd := i + ansiLen + 1
				ansi := styledSegment[i:escEnd]
				if ansi == RST {
					activeStyles = []string{} // reset
				} else {
					activeStyles = append(activeStyles, ansi) // add new active style
				}
				result.WriteString(ansi)
				i = escEnd
				inAnsi = false
				continue
			}
		}

		if !inAnsi {
			// check if need to start a highlight at this position
			for highlightIdx < len(applicableHighlights) &&
				applicableHighlights[highlightIdx].startByte == nonAnsiBytes {
				highlight := applicableHighlights[highlightIdx]

				// reset current styles if any
				if len(activeStyles) > 0 {
					result.WriteString(RST)
				}

				// extract and apply highlight text
				plainText := getNonAnsiBytes(styledSegment, i, highlight.endByte-highlight.startByte)
				result.WriteString(highlight.style.Render(plainText))

				// restore previous styles if any
				if len(activeStyles) > 0 {
					for _, style := range activeStyles {
						result.WriteString(style)
					}
				}

				// skip highlighted text
				count := 0
				for count < len(plainText) && i < len(styledSegment) {
					if strings.HasPrefix(styledSegment[i:], "\x1b[") {
						escEnd := i + strings.Index(styledSegment[i:], "m") + 1
						result.WriteString(styledSegment[i:escEnd])
						i = escEnd
						continue
					}
					i++
					count++
				}
				nonAnsiBytes += len(plainText)
				highlightIdx++

				// skip to next highlight that doesn't overlap
				for highlightIdx < len(applicableHighlights) &&
					applicableHighlights[highlightIdx].startByte < nonAnsiBytes {
					highlightIdx++
				}

				continue
			}
		}

		// regular character
		if i < len(styledSegment) {
			result.WriteByte(styledSegment[i])
			if !inAnsi {
				nonAnsiBytes++
			}
		}
		i++
	}

	return removeEmptyAnsiSequences(result.String())
}

func stripAnsi(input string) string {
	ranges := findAnsiByteRanges(input)
	if len(ranges) == 0 {
		return input
	}

	totalAnsiLen := 0
	for _, r := range ranges {
		totalAnsiLen += int(r[1] - r[0])
	}

	finalLen := len(input) - totalAnsiLen
	var builder strings.Builder
	builder.Grow(finalLen)

	lastPos := 0
	for _, r := range ranges {
		builder.WriteString(input[lastPos:int(r[0])])
		lastPos = int(r[1])
	}

	builder.WriteString(input[lastPos:])
	return builder.String()
}

func simplifyAnsiCodes(ansis []string) []string {
	if len(ansis) == 0 {
		return []string{}
	}

	// if there's just a bunch of reset sequences, compress it to one
	allReset := true
	for _, ansi := range ansis {
		if ansi != RST {
			allReset = false
			break
		}
	}
	if allReset {
		return []string{RST}
	}

	// return all ansis to the right of the rightmost reset seq
	for i := len(ansis) - 1; i >= 0; i-- {
		if ansis[i] == RST {
			result := ansis[i+1:]
			// keep reset at the start if present
			if ansis[0] == RST {
				return append([]string{RST}, result...)
			}
			return result
		}
	}
	return ansis
}

func runesHaveAnsiPrefix(runes []rune) bool {
	return len(runes) >= 2 && runes[0] == '\x1b' && runes[1] == '['
}

func findAnsiByteRanges(s string) [][]uint32 {
	// pre-count to allocate exact size
	count := strings.Count(s, "\x1b[")
	if count == 0 {
		return nil
	}

	allRanges := make([]uint32, count*2)
	ranges := make([][]uint32, count)

	for i := 0; i < count; i++ {
		ranges[i] = allRanges[i*2 : i*2+2]
	}

	rangeIdx := 0
	for i := 0; i < len(s); {
		if i+1 < len(s) && s[i] == '\x1b' && s[i+1] == '[' {
			start := i
			i += 2 // skip \x1b[

			// find the 'm' that ends this sequence
			for i < len(s) && s[i] != 'm' {
				i++
			}

			if i < len(s) && s[i] == 'm' {
				allRanges[rangeIdx*2] = clampIntToUint32(start)
				allRanges[rangeIdx*2+1] = clampIntToUint32(i + 1)
				rangeIdx++
				i++
				continue
			}
		}
		i++
	}
	return ranges[:rangeIdx]
}

func findAnsiRuneRanges(s string) [][]uint32 {
	// pre-count to allocate exact size
	count := strings.Count(s, "\x1b[")
	if count == 0 {
		return nil
	}

	allRanges := make([]uint32, count*2)
	ranges := make([][]uint32, count)

	for i := 0; i < count; i++ {
		ranges[i] = allRanges[i*2 : i*2+2]
	}

	rangeIdx := 0
	runes := []rune(s)
	for i := 0; i < len(runes); {
		if i+1 < len(runes) && runes[i] == '\x1b' && runes[i+1] == '[' {
			start := i
			i += 2 // skip \x1b[

			// find the 'm' that ends this sequence
			for i < len(runes) && runes[i] != 'm' {
				i++
			}

			if i < len(runes) && runes[i] == 'm' {
				allRanges[rangeIdx*2] = clampIntToUint32(start)
				allRanges[rangeIdx*2+1] = clampIntToUint32(i + 1)
				rangeIdx++
				i++
				continue
			}
		}
		i++
	}
	return ranges[:rangeIdx]
}

func removeEmptyAnsiSequences(s string) string {
	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		if i < len(s)-4 && s[i:i+2] == "\x1b[" {
			// find the end of this ansi sequence
			end := i + 2
			for end < len(s) && s[end] != 'm' {
				end++
			}
			if end < len(s) {
				end++ // include the 'm'
				ansiSeq := s[i:end]

				// check if this is followed immediately by a reset sequence
				if end < len(s)-2 && s[end:end+2] == "\x1b[" {
					resetEnd := end + 2
					for resetEnd < len(s) && s[resetEnd] != 'm' {
						resetEnd++
					}
					if resetEnd < len(s) {
						resetEnd++ // include the 'm'
						resetSeq := s[end:resetEnd]

						// if this is a reset sequence (\x1b[0m or \x1b[m), skip both sequences
						if resetSeq == "\x1b[0m" || resetSeq == "\x1b[m" {
							i = resetEnd
							continue
						}
					}
				}

				// not followed by reset, keep the sequence
				result.WriteString(ansiSeq)
				i = end
				continue
			}
		}

		result.WriteByte(s[i])
		i++
	}

	return result.String()
}
