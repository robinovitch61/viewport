package filterable_viewport

import (
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/robinovitch61/bubbleo/viewport"
)

// KeyMap defines the key bindings for the filterable viewport
type KeyMap struct {
	ViewportKeyMap       viewport.KeyMap
	FilterKey            key.Binding
	RegexFilterKey       key.Binding
	ApplyFilterKey       key.Binding
	CancelFilterKey      key.Binding
	ToggleMatchesOnlyKey key.Binding
}

// DefaultKeyMap returns a default keymap for the filterable viewport
func DefaultKeyMap() KeyMap {
	return KeyMap{
		ViewportKeyMap: viewport.DefaultKeyMap(),
		FilterKey: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		RegexFilterKey: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "regex filter"),
		),
		ApplyFilterKey: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "apply filter"),
		),
		CancelFilterKey: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel filter"),
		),
		ToggleMatchesOnlyKey: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "toggle matches only"),
		),
	}
}
