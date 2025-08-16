package viewport

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/viewport/internal"
)

// NavigationManager manages keyboard input and navigation logic
type NavigationManager struct {
	// KeyMap is the keymap for the viewport
	KeyMap KeyMap

	// SelectionEnabled is true if the viewport allows individual line selection
	SelectionEnabled bool

	// TopSticky is true when selection should remain at the top until user manually scrolls down
	TopSticky bool

	// BottomSticky is true when selection should remain at the bottom until user manually scrolls up
	BottomSticky bool
}

// NewNavigationManager creates a new NavigationManager with the specified key mappings.
func NewNavigationManager(keyMap KeyMap) *NavigationManager {
	return &NavigationManager{
		KeyMap:           keyMap,
		SelectionEnabled: false,
		TopSticky:        false,
		BottomSticky:     false,
	}
}

// NavigationAction represents a navigation command
type NavigationAction int

const (
	// ActionNone represents no navigation action.
	ActionNone NavigationAction = iota
	// ActionUp represents moving up one item.
	ActionUp
	// ActionDown represents moving down one item.
	ActionDown
	// ActionLeft represents moving left horizontally.
	ActionLeft
	// ActionRight represents moving right horizontally.
	ActionRight
	// ActionHalfPageUp represents moving up half a page.
	ActionHalfPageUp
	// ActionHalfPageDown represents moving down half a page.
	ActionHalfPageDown
	// ActionPageUp represents moving up one page.
	ActionPageUp
	// ActionPageDown represents moving down one page.
	ActionPageDown
	// ActionTop represents moving to the top.
	ActionTop
	// ActionBottom represents moving to the bottom.
	ActionBottom
)

// NavigationContext contains the context needed for navigation calculations
type NavigationContext struct {
	WrapText        bool
	Dimensions      internal.Rectangle
	NumContentLines int
	NumVisibleItems int
}

// NavigationResult contains the result of processing a navigation action
type NavigationResult struct {
	Action          NavigationAction
	ScrollAmount    int // lines to scroll
	SelectionAmount int // items to move selection
}

// ProcessKeyMsg processes a keyboard message and returns the corresponding navigation action
func (nm NavigationManager) ProcessKeyMsg(msg tea.KeyMsg, ctx NavigationContext) NavigationResult {
	switch {
	case key.Matches(msg, nm.KeyMap.Up):
		return NavigationResult{Action: ActionUp, ScrollAmount: 1, SelectionAmount: 1}

	case key.Matches(msg, nm.KeyMap.Down):
		return NavigationResult{Action: ActionDown, ScrollAmount: 1, SelectionAmount: 1}

	case key.Matches(msg, nm.KeyMap.Left):
		if !ctx.WrapText {
			return NavigationResult{Action: ActionLeft, ScrollAmount: ctx.Dimensions.Width / 4}
		}

	case key.Matches(msg, nm.KeyMap.Right):
		if !ctx.WrapText {
			return NavigationResult{Action: ActionRight, ScrollAmount: ctx.Dimensions.Width / 4}
		}

	case key.Matches(msg, nm.KeyMap.HalfPageUp):
		scrollAmount := ctx.NumContentLines / 2
		selectionAmount := max(1, ctx.NumVisibleItems/2)
		return NavigationResult{Action: ActionHalfPageUp, ScrollAmount: scrollAmount, SelectionAmount: selectionAmount}

	case key.Matches(msg, nm.KeyMap.HalfPageDown):
		scrollAmount := ctx.NumContentLines / 2
		selectionAmount := max(1, ctx.NumVisibleItems/2)
		return NavigationResult{Action: ActionHalfPageDown, ScrollAmount: scrollAmount, SelectionAmount: selectionAmount}

	case key.Matches(msg, nm.KeyMap.PageUp):
		scrollAmount := ctx.NumContentLines
		selectionAmount := ctx.NumVisibleItems
		return NavigationResult{Action: ActionPageUp, ScrollAmount: scrollAmount, SelectionAmount: selectionAmount}

	case key.Matches(msg, nm.KeyMap.PageDown):
		scrollAmount := ctx.NumContentLines
		selectionAmount := ctx.NumVisibleItems
		return NavigationResult{Action: ActionPageDown, ScrollAmount: scrollAmount, SelectionAmount: selectionAmount}

	case key.Matches(msg, nm.KeyMap.Top):
		return NavigationResult{Action: ActionTop}

	case key.Matches(msg, nm.KeyMap.Bottom):
		return NavigationResult{Action: ActionBottom}
	}

	return NavigationResult{Action: ActionNone}
}
