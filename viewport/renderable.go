package viewport

import "github.com/robinovitch61/bubbleo/viewport/linebuffer"

// Renderable represents objects that can be rendered as line buffers.
type Renderable interface {
	Render() linebuffer.LineBufferer
}

// Item is a simple implementation of Renderable that holds a line buffer
type Item struct {
	LineBuffer linebuffer.LineBuffer
}

// Render returns the line buffer for the Item
func (i Item) Render() linebuffer.LineBufferer {
	return i.LineBuffer
}

// ItemCompareFn is a comparator function
func ItemCompareFn(a, b Item) bool {
	return a.LineBuffer.Content() == b.LineBuffer.Content()
}

// assert Item implements viewport.Renderable
var _ Renderable = Item{}
