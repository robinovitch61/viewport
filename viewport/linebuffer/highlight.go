package linebuffer

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TODO LEO: index-agnostic highlights

// Highlight represents a specific position and style to highlight
type Highlight struct {
	ItemIndex       int            // index of the item containing the highlight
	StartByteOffset int            // start byte offset within the item's content
	EndByteOffset   int            // end byte offset within the item's content
	Style           lipgloss.Style // style to apply to this highlight
}

// ExtractHighlights extracts highlights from a slice of strings and a match string
func ExtractHighlights(
	items []string,
	exactMatch string,
	highlightStyle lipgloss.Style,
) []Highlight {
	var highlights []Highlight

	if exactMatch == "" {
		return highlights
	}

	for i, item := range items {
		plainLine := stripAnsi(item)
		startIndex := 0
		for {
			foundIndex := strings.Index(plainLine[startIndex:], exactMatch)
			if foundIndex == -1 {
				break
			}
			actualStartIndex := startIndex + foundIndex
			endIndex := actualStartIndex + len(exactMatch)

			highlights = append(highlights, Highlight{
				ItemIndex:       i,
				StartByteOffset: actualStartIndex,
				EndByteOffset:   endIndex,
				Style:           highlightStyle,
			})
			startIndex = actualStartIndex + 1 // Move past this match to find overlapping matches
		}
	}
	return highlights
}

// ExtractHighlightsRegexMatch extracts highlights from a slice of strings based on a regex match
func ExtractHighlightsRegexMatch(
	items []string,
	regexPattern string,
	highlightStyle lipgloss.Style,
) ([]Highlight, error) {
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	var highlights []Highlight
	for i, item := range items {
		plainLine := stripAnsi(item)
		matches := regex.FindAllStringIndex(plainLine, -1)
		for _, match := range matches {
			highlights = append(highlights, Highlight{
				ItemIndex:       i,
				StartByteOffset: match[0],
				EndByteOffset:   match[1],
				Style:           highlightStyle,
			})
		}
	}
	return highlights, nil
}
