package filterableviewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport"
)

// Styles contains styling configuration for the filterable viewport
type Styles struct {
	Viewport          viewport.Styles
	FocusedMatchStyle lipgloss.Style
}

// DefaultStyles returns a set of default styles for the filterable viewport
func DefaultStyles() Styles {
	// more prominent style for the focused match
	focusedMatchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("11"))

	return Styles{
		Viewport:          viewport.DefaultStyles(),
		FocusedMatchStyle: focusedMatchStyle,
	}
}
