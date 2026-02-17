package viewport

import "github.com/robinovitch61/viewport/viewport/item"

// Object is implemented by types that can return an Item
// It exists to allow the viewport to return the selected object without (de)serializing it
type Object interface {
	GetItem() item.Item
}
