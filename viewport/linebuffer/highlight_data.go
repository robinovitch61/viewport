package linebuffer

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)

// HighlightData contains information about what to highlight in each item in the viewport.
type HighlightData struct {
	StringToHighlight       string
	RegexPatternToHighlight *regexp.Regexp
	IsRegex                 bool
	SpecificHighlights      []Highlight
}

// Highlight represents a specific position and style to highlight
type Highlight struct {
	ItemIndex       int            // index of the item containing the highlight
	StartByteOffset int            // start byte offset within the item's content
	EndByteOffset   int            // end byte offset within the item's content
	Style           lipgloss.Style // style to apply to this highlight
}
