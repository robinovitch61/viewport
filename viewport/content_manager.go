package viewport

import "github.com/robinovitch61/bubbleo/viewport/linebuffer"

// ContentManager manages the actual LineBuffer and selection state
type ContentManager[T Renderable] struct {
	// Items is the complete list of items to be rendered in the viewport
	Items []T

	// Header is the fixed header lines at the top of the viewport
	// these lines wrap and are horizontally scrollable similar to other rendered items
	Header []string

	// selectedIdx is the index of Items of the current selection (only relevant when selection is enabled)
	selectedIdx int

	// Highlights is what to highlight wherever it shows up within an item, even wrapped between lines
	Highlights []linebuffer.Highlight

	// highlightsByItem is a cache of highlights indexed by item index for O(1) lookup
	highlightsByItem map[int][]linebuffer.Highlight

	// CompareFn is an optional function to compare items for maintaining the selection when LineBuffer changes
	// if set, the viewport will try to maintain the previous selected item when LineBuffer changes
	CompareFn CompareFn[T]
}

// NewContentManager creates a new ContentManager with empty initial state.
func NewContentManager[T Renderable]() *ContentManager[T] {
	return &ContentManager[T]{
		Items:            []T{},
		Header:           []string{},
		selectedIdx:      0,
		highlightsByItem: make(map[int][]linebuffer.Highlight),
	}
}

// SetSelectedIdx sets the selected item index.
func (cm *ContentManager[T]) SetSelectedIdx(idx int) {
	cm.selectedIdx = clampValZeroToMax(idx, len(cm.Items)-1)
}

// GetSelectedIdx returns the current selected item index.
func (cm *ContentManager[T]) GetSelectedIdx() int {
	return cm.selectedIdx
}

// GetSelectedItem returns a pointer to the currently selected item, or nil if none selected.
func (cm *ContentManager[T]) GetSelectedItem() *T {
	if cm.selectedIdx >= len(cm.Items) || cm.selectedIdx < 0 {
		return nil
	}
	return &cm.Items[cm.selectedIdx]
}

// NumItems returns the total number of items.
func (cm *ContentManager[T]) NumItems() int {
	return len(cm.Items)
}

// IsEmpty returns true if there are no items.
func (cm *ContentManager[T]) IsEmpty() bool {
	return len(cm.Items) == 0
}

// ValidateSelectedIdx ensures the selected index is within valid bounds.
func (cm *ContentManager[T]) ValidateSelectedIdx() {
	if len(cm.Items) == 0 {
		cm.selectedIdx = 0
		return
	}
	cm.selectedIdx = clampValZeroToMax(cm.selectedIdx, len(cm.Items)-1)
}

// rebuildHighlightsCache rebuilds the highlights-by-item-index cache for O(1) lookup.
func (cm *ContentManager[T]) rebuildHighlightsCache() {
	cm.highlightsByItem = make(map[int][]linebuffer.Highlight)
	for _, highlight := range cm.Highlights {
		itemIdx := highlight.ItemIndex
		cm.highlightsByItem[itemIdx] = append(cm.highlightsByItem[itemIdx], highlight)
	}
}

// SetHighlights sets the highlights and rebuilds the cache.
func (cm *ContentManager[T]) SetHighlights(highlights []linebuffer.Highlight) {
	cm.Highlights = highlights
	cm.rebuildHighlightsCache()
}

// GetHighlightsForItem returns highlights for a specific item index in O(1) time.
func (cm *ContentManager[T]) GetHighlightsForItem(itemIndex int) []linebuffer.Highlight {
	return cm.highlightsByItem[itemIndex]
}
