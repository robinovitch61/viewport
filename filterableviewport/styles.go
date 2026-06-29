package filterableviewport

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

// Styles contains styling configuration for the filterable viewport
type Styles struct {
	Filter       FilterStyles
	MatchesCount MatchesCountStyles
	Match        MatchStyles
}

// MatchStyles contains styles for matches in the filterable viewport
type MatchStyles struct {
	Focused           lipgloss.Style
	FocusedIfSelected lipgloss.Style // used when the focused match is on the selected item
	Unfocused         lipgloss.Style
}

// DefaultMatchStyles returns a set of default styles for matches.
// Uses only reverse video and safe ANSI colors — no 256-color or true-color values.
func DefaultMatchStyles() MatchStyles {
	return MatchStyles{
		Focused:           lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Cyan),
		FocusedIfSelected: lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.Cyan),
		Unfocused:         lipgloss.NewStyle().Reverse(true).Foreground(lipgloss.BrightRed),
	}
}

// FilterStyles contains styles for the filter line in the filterable viewport
type FilterStyles struct {
	Empty       lipgloss.Style
	Placeholder lipgloss.Style
	Cursor      textinput.CursorStyle
	Focused     StateFilterStyles
	Unfocused   StateFilterStyles
}

// StateFilterStyles contains styles for the filter line for a given focus state
type StateFilterStyles struct {
	Prefix    lipgloss.Style
	TextInput textinput.StyleState
}

// DefaultFilterStyles returns a set of default styles for the filter line.
func DefaultFilterStyles() FilterStyles {
	return FilterStyles{
		Placeholder: lipgloss.NewStyle().Faint(true),
		Focused: StateFilterStyles{
			Prefix: lipgloss.NewStyle().Bold(true),
		},
	}
}

// MatchesCountStyles contains styles for the matches count
type MatchesCountStyles struct {
	NoMatches lipgloss.Style
	Matches   lipgloss.Style
}

// DefaultMatchesCountStyles returns a set of default styles for the matches suffix.
func DefaultMatchesCountStyles() MatchesCountStyles {
	return MatchesCountStyles{
		NoMatches: lipgloss.NewStyle().Faint(true),
	}
}

// DefaultStyles returns a set of default styles for the filterable viewport
func DefaultStyles() Styles {
	return Styles{
		Match:        DefaultMatchStyles(),
		MatchesCount: DefaultMatchesCountStyles(),
		Filter:       DefaultFilterStyles(),
	}
}
