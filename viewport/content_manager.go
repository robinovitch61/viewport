package viewport

// contentManager manages the actual SingleItem and selection state
type contentManager[T Renderable] struct {
	// items is the complete list of items to be rendered in the viewport
	items []T

	// header is the fixed header lines at the top of the viewport
	// these lines wrap and are horizontally scrollable similar to other rendered items
	header []string

	// selectedIdx is the index of items of the current selection (only relevant when selection is enabled)
	selectedIdx int

	// highlights is what to highlight wherever it shows up within an item, even wrapped between lines
	highlights []Highlight

	// highlightsByItem is a cache of highlights indexed by item index for O(1) lookup
	highlightsByItem map[int][]Highlight

	// compareFn is an optional function to compare items for maintaining the selection when SingleItem changes
	// if set, the viewport will try to maintain the previous selected item when SingleItem changes
	compareFn CompareFn[T]
}

// newContentManager creates a new contentManager with empty initial state.
func newContentManager[T Renderable]() *contentManager[T] {
	return &contentManager[T]{
		items:            []T{},
		header:           []string{},
		selectedIdx:      0,
		highlightsByItem: make(map[int][]Highlight),
	}
}

// setSelectedIdx sets the selected item index.
func (cm *contentManager[T]) setSelectedIdx(idx int) {
	cm.selectedIdx = clampValZeroToMax(idx, len(cm.items)-1)
}

// getSelectedIdx returns the current selected item index.
func (cm *contentManager[T]) getSelectedIdx() int {
	return cm.selectedIdx
}

// getSelectedItem returns a pointer to the currently selected item, or nil if none selected.
func (cm *contentManager[T]) getSelectedItem() *T {
	if cm.selectedIdx >= len(cm.items) || cm.selectedIdx < 0 {
		return nil
	}
	return &cm.items[cm.selectedIdx]
}

// numItems returns the total number of items.
func (cm *contentManager[T]) numItems() int {
	return len(cm.items)
}

// isEmpty returns true if there are no items.
func (cm *contentManager[T]) isEmpty() bool {
	return len(cm.items) == 0
}

// validateSelectedIdx ensures the selected index is within valid bounds.
func (cm *contentManager[T]) validateSelectedIdx() {
	if len(cm.items) == 0 {
		cm.selectedIdx = 0
		return
	}
	cm.selectedIdx = clampValZeroToMax(cm.selectedIdx, len(cm.items)-1)
}

// rebuildHighlightsCache rebuilds the highlights-by-item-index cache for O(1) lookup.
func (cm *contentManager[T]) rebuildHighlightsCache() {
	cm.highlightsByItem = make(map[int][]Highlight)
	for _, highlight := range cm.highlights {
		itemIdx := highlight.ItemIndex
		cm.highlightsByItem[itemIdx] = append(cm.highlightsByItem[itemIdx], highlight)
	}
}

// setHighlights sets the highlights and rebuilds the cache.
func (cm *contentManager[T]) setHighlights(highlights []Highlight) {
	cm.highlights = highlights
	cm.rebuildHighlightsCache()
}

// getHighlights returns all highlights.
func (cm *contentManager[T]) getHighlights() []Highlight {
	return cm.highlights
}

// getHighlightsForItem returns highlights for a specific item index in O(1) time.
func (cm *contentManager[T]) getHighlightsForItem(itemIndex int) []Highlight {
	return cm.highlightsByItem[itemIndex]
}
