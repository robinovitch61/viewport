package item

import (
	"github.com/charmbracelet/lipgloss"
)

// ByteRange represents a range of bytes
type ByteRange struct {
	Start int
	End   int
}

// Highlight represents a range and style to highlight
type Highlight struct {
	Style                    lipgloss.Style
	ByteRangeUnstyledContent ByteRange
}
