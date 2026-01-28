package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
)

// Styles contains styling configuration for the viewport
type Styles struct {
	FooterStyle              lipgloss.Style
	HighlightStyle           lipgloss.Style
	HighlightStyleIfSelected lipgloss.Style
	SelectedItemStyle        lipgloss.Style
}

// DefaultStyles returns a set of default styles for the viewport
func DefaultStyles() Styles {
	// in the future could pass in `hasDarkBackground` and determine foreground/background colors based on that
	white := lipgloss.Color("255")
	darkGrey := lipgloss.Color("245")

	return Styles{
		FooterStyle:              lipgloss.NewStyle().Foreground(darkGrey).Padding(0, 1, 0, 0),
		HighlightStyle:           lipgloss.NewStyle().Foreground(white).Background(lipgloss.Color("5")),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(darkGrey).Background(lipgloss.Color("14")),
		SelectedItemStyle:        lipgloss.NewStyle().Foreground(white).Background(lipgloss.Color("12")),
	}
}
