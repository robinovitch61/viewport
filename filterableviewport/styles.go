package filterableviewport

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles contains styling configuration for the filterable viewport
type Styles struct {
	CursorStyle lipgloss.Style
	Match       MatchStyles
}

// MatchStyles contains styles for matches in the filterable viewport
type MatchStyles struct {
	Focused   lipgloss.Style
	Unfocused lipgloss.Style
}

// DefaultMatchStyles returns a set of default styles for matches
func DefaultMatchStyles() MatchStyles {
	return MatchStyles{
		Focused:   lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("11")),
		Unfocused: lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("7")),
	}
}

// DefaultStyles returns a set of default styles for the filterable viewport
func DefaultStyles() Styles {
	return Styles{
		CursorStyle: lipgloss.Style{}, // default to bubbles.textinput cursor style
		Match:       DefaultMatchStyles(),
	}
}
