package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport/item"
)

// Highlight represents a specific position and style to highlight
type Highlight struct {
	ItemIndex                int            // index of the item
	Style                    lipgloss.Style // style to apply to this highlight
	ByteRangeUnstyledContent item.ByteRange // byte range in the unstyled content to highlight
}
