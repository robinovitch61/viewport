package viewport

import (
	"github.com/charmbracelet/lipgloss"
)

// displayManager handles all display/rendering concerns
type displayManager struct {
	// bounds contains the viewport dimensions
	bounds Rectangle

	// topItemIdx is the items index of the topmost visible item
	topItemIdx int

	// topItemLineOffset is the number of lines in the top item that are out of view of the topmost visible line.
	// Only non-zero when wrapped
	topItemLineOffset int

	// xOffset is the number of terminal cells scrolled right when rendered lines overflow the viewport and wrapping is off
	xOffset int

	// styles contains the styling configuration
	styles Styles
}

// newDisplayManager creates a new displayManager with the specified dimensions and styles.
func newDisplayManager(width, height int, styles Styles) *displayManager {
	return &displayManager{
		bounds: Rectangle{
			Width:  max(0, width),
			Height: max(0, height),
		},
		topItemIdx:        0,
		topItemLineOffset: 0,
		xOffset:           0,
		styles:            styles,
	}
}

// setBounds sets the viewport dimensions with validation
func (dm *displayManager) setBounds(width, height int) {
	dm.bounds.Width = max(0, width)
	dm.bounds.Height = max(0, height)
}

// safelySetTopItemIdxAndOffset safely sets the top item index and offset within bounds.
func (dm *displayManager) safelySetTopItemIdxAndOffset(topItemIdx, topItemLineOffset, maxTopItemIdx, maxTopItemLineOffset int) {
	dm.topItemIdx = clampValZeroToMax(topItemIdx, maxTopItemIdx)
	dm.topItemLineOffset = topItemLineOffset
	if dm.topItemIdx == maxTopItemIdx {
		dm.topItemLineOffset = clampValZeroToMax(topItemLineOffset, maxTopItemLineOffset)
	}
}

// getNumContentLines returns the number of lines available for SingleItem display.
func (dm *displayManager) getNumContentLines(headerLines int, showFooter bool) int {
	contentHeight := dm.bounds.Height - headerLines
	if showFooter {
		contentHeight-- // one for footer
	}
	return max(0, contentHeight)
}

// renderFinalView applies final styling to the rendered
func (dm *displayManager) renderFinalView(content string) string {
	return lipgloss.NewStyle().Width(dm.bounds.Width).Height(dm.bounds.Height).Render(content)
}
