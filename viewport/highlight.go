package viewport

import (
	"github.com/robinovitch61/viewport/viewport/item"
)

// Highlight represents a specific position and style to highlight
type Highlight struct {
	ItemIndex     int // index of the item
	ItemHighlight item.Highlight
}
