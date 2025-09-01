package viewport

// Renderable represents objects that can be rendered as line items.
type Renderable interface {
	Render() Item
}

// ExampleItem is a simple implementation of Renderable that holds a line buffer
type ExampleItem struct {
	LineBuffer SingleItem
}

// Render returns the line buffer for the ExampleItem
func (i ExampleItem) Render() Item {
	return i.LineBuffer
}

// ItemCompareFn is a comparator function
func ItemCompareFn(a, b ExampleItem) bool {
	return a.LineBuffer.Content() == b.LineBuffer.Content()
}

// assert ExampleItem implements viewport.Renderable
var _ Renderable = ExampleItem{}
