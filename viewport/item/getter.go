package item

// Getter is implemented by types that can return an Item
// TODO LEO: rename to ViewportObject or move to viewport and do viewport.Object?
type Getter interface {
	Get() Item
}

// SimpleGetter is a simple implementation of Getter that holds an Item
type SimpleGetter struct {
	Item Item
}

// Get returns the Item
func (i SimpleGetter) Get() Item {
	return i.Item
}

// SimpleGettersEqual is a comparator function for SimpleGetter items
func SimpleGettersEqual(a, b SimpleGetter) bool {
	if a.Item == nil || b.Item == nil {
		return a.Item == b.Item
	}
	return a.Item.Content() == b.Item.Content()
}

// assert SimpleGetter implements viewport.Renderable
var _ Getter = SimpleGetter{}
