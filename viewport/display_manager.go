package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport/internal"
)

// DisplayManager handles all display/rendering concerns
type DisplayManager struct {
	// Bounds contains the viewport dimensions
	Bounds internal.Rectangle

	// TopItemIdx is the items index of the topmost visible item
	TopItemIdx int

	// TopItemLineOffset is the number of lines in the top item that are out of view of the topmost visible line.
	// Only non-zero when wrapped
	TopItemLineOffset int

	// XOffset is the number of terminal cells scrolled right when rendered lines overflow the viewport and wrapping is off
	XOffset int

	// Styles contains the styling configuration
	Styles Styles
}

// NewDisplayManager creates a new DisplayManager with the specified dimensions and styles.
func NewDisplayManager(width, height int, styles Styles) *DisplayManager {
	return &DisplayManager{
		Bounds: internal.Rectangle{
			Width:  max(0, width),
			Height: max(0, height),
		},
		TopItemIdx:        0,
		TopItemLineOffset: 0,
		XOffset:           0,
		Styles:            styles,
	}
}

// SetBounds sets the viewport dimensions with validation
func (dm *DisplayManager) SetBounds(width, height int) {
	dm.Bounds.Width = max(0, width)
	dm.Bounds.Height = max(0, height)
}

// GetHighlightStyle returns the appropriate highlight style based on selection state.
func (dm *DisplayManager) GetHighlightStyle(isSelected bool) lipgloss.Style {
	if isSelected {
		return dm.Styles.HighlightStyleIfSelected
	}
	return dm.Styles.HighlightStyle
}

// SafelySetTopItemIdxAndOffset safely sets the top item index and offset within bounds.
func (dm *DisplayManager) SafelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset, maxTopItemIdx, maxTopItemLineOffset int) {
	dm.TopItemIdx = clampValZeroToMax(topItemIdx, maxTopItemIdx)
	dm.TopItemLineOffset = topItemLineOffset
	if dm.TopItemIdx == maxTopItemIdx {
		dm.TopItemLineOffset = clampValZeroToMax(topItemLineOffset, maxTopItemLineOffset)
	}
}

// GetNumContentLines returns the number of lines available for LineBuffer display.
func (dm *DisplayManager) GetNumContentLines(headerLines int, showFooter bool) int {
	contentHeight := dm.Bounds.Height - headerLines
	if showFooter {
		contentHeight-- // one for footer
	}
	return max(0, contentHeight)
}

// RenderFinalView applies final styling to the rendered LineBuffer.
func (dm *DisplayManager) RenderFinalView(content string) string {
	return lipgloss.NewStyle().Width(dm.Bounds.Width).Height(dm.Bounds.Height).Render(content)
}
