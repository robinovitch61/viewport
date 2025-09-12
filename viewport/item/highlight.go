package item

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Highlight represents a specific position and style to highlight
type Highlight struct {
	Match Match          // the match details
	Style lipgloss.Style // style to apply to this highlight
}

// Match represents a match of a substring within an item
type Match struct {
	ItemIndex       int // index of the item containing the match
	StartByteOffset int // start position of the match within the item's content
	EndByteOffset   int // end position of the match within the item's content
}

// ExtractMatches extracts highlights from a slice of strings and a match string
// Strings should not contain ansi styling codes
func ExtractMatches(vals []string, exactMatch string) []Match {
	var matches []Match

	if exactMatch == "" {
		return matches
	}

	for i, item := range vals {
		startIndex := 0
		for {
			foundIndex := strings.Index(item[startIndex:], exactMatch)
			if foundIndex == -1 {
				break
			}
			actualStartIndex := startIndex + foundIndex
			endIndex := actualStartIndex + len(exactMatch)

			matches = append(matches, Match{
				ItemIndex:       i,
				StartByteOffset: actualStartIndex,
				EndByteOffset:   endIndex,
			})
			startIndex = actualStartIndex + 1 // move past this match to find overlapping matches
		}
	}
	return matches
}

// ExtractMatchesRegex extracts matches from a slice of strings based on a regex pattern
// Strings should not contain ansi styling codes
func ExtractMatchesRegex(vals []string, regexPattern string) ([]Match, error) {
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	var matches []Match
	for i, item := range vals {
		regexMatches := regex.FindAllStringIndex(item, -1)
		for _, regexMatch := range regexMatches {
			matches = append(matches, Match{
				ItemIndex:       i,
				StartByteOffset: regexMatch[0],
				EndByteOffset:   regexMatch[1],
			})
		}
	}
	return matches, nil
}
