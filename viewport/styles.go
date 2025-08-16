package viewport

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles contains styling configuration for the viewport
type Styles struct {
	FooterStyle              lipgloss.Style
	HighlightStyle           lipgloss.Style
	HighlightStyleIfSelected lipgloss.Style
	SelectedItemStyle        lipgloss.Style
}

// DefaultStyles returns a set of default styles for the viewport
func DefaultStyles(hasDarkBackground bool) Styles {
	lightGrey := lipgloss.Color("7")
	darkGrey := lipgloss.Color("245")
	textForeground := lightGrey
	if hasDarkBackground {
		textForeground = darkGrey
	}

	return Styles{
		FooterStyle:              lipgloss.NewStyle().Foreground(textForeground).Padding(0, 1),
		HighlightStyle:           lipgloss.NewStyle().Foreground(textForeground).Background(lipgloss.Color("2")),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(textForeground).Background(lipgloss.Color("3")),
		SelectedItemStyle:        lipgloss.NewStyle().Foreground(textForeground).Background(lipgloss.Color("2")),
	}
}
