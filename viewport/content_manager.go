package viewport

import "github.com/robinovitch61/bubbleo/viewport/item"

// contentManager manages the actual Item and selection state
type contentManager[T Object] struct {
	// objects is the viewport objects
	objects []T

	// header is the unselectable lines at the top of the viewport
	// these lines wrap, but don't pan horizontally like other non-wrapped lines
	header []string

	// selectedIdx is the index of objects of the current selection (only relevant when selection is enabled)
	selectedIdx int

	// highlights is what to highlight wherever it shows up within an item, even wrapped between lines
	highlights []item.Highlight

	// highlightsByItem is a cache of highlights indexed by item index
	highlightsByItem map[int][]item.Highlight

	// compareFn is an optional function to compare items for maintaining the selection when Item changes
	// if set, the viewport will try to maintain the previous selected item when Item changes
	compareFn CompareFn[T]
}

// newContentManager creates a new contentManager with empty initial state
func newContentManager[T Object]() *contentManager[T] {
	return &contentManager[T]{
		objects:          make([]T, 0),
		header:           []string{},
		selectedIdx:      0,
		highlightsByItem: make(map[int][]item.Highlight),
	}
}

// setSelectedIdx sets the selected item index
func (cm *contentManager[T]) setSelectedIdx(idx int) {
	cm.selectedIdx = clampValZeroToMax(idx, len(cm.objects)-1)
}

// getSelectedIdx returns the current selected item index
func (cm *contentManager[T]) getSelectedIdx() int {
	return cm.selectedIdx
}

// getSelectedItem returns a pointer to the currently selected item, or nil if none selected
func (cm *contentManager[T]) getSelectedItem() *T {
	if cm.selectedIdx >= len(cm.objects) || cm.selectedIdx < 0 {
		return nil
	}
	return &cm.objects[cm.selectedIdx]
}

// numItems returns the total number of items
func (cm *contentManager[T]) numItems() int {
	return len(cm.objects)
}

// isEmpty returns true if there are no items
func (cm *contentManager[T]) isEmpty() bool {
	return len(cm.objects) == 0
}

// rebuildHighlightsCache rebuilds the internal highlight cache
func (cm *contentManager[T]) rebuildHighlightsCache() {
	cm.highlightsByItem = make(map[int][]item.Highlight)
	for _, highlight := range cm.highlights {
		itemIdx := highlight.ItemIndex
		cm.highlightsByItem[itemIdx] = append(cm.highlightsByItem[itemIdx], highlight)
	}
}

// setHighlights sets the highlights
func (cm *contentManager[T]) setHighlights(highlights []item.Highlight) {
	cm.highlights = highlights
	cm.rebuildHighlightsCache()
}

// getHighlights returns all highlights
func (cm *contentManager[T]) getHighlights() []item.Highlight {
	return cm.highlights
}

// getHighlightsForItem returns highlights for a specific item index
func (cm *contentManager[T]) getHighlightsForItem(itemIndex int) []item.Highlight {
	return cm.highlightsByItem[itemIndex]
}
