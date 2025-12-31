package viewport

import (
	"github.com/charmbracelet/bubbles/key"
)

// configuration consolidates all configuration options for the viewport
type configuration struct {
	// wrapText is true if the viewport wraps text rather than showing that a line is truncated/horizontally scrollable
	wrapText bool

	// footerEnabled is true if the viewport currently shows the footer based on its dimensions and content
	footerEnabled bool

	// continuationIndicator is the string to use to indicate that an unwrapped line continues to the left or right
	continuationIndicator string

	// saveDir is the directory where files are saved when the save key is pressed
	saveDir string

	// saveKey is the key binding for saving viewport content to a file
	saveKey key.Binding
}

// newConfiguration creates a new configuration with default settings.
func newConfiguration() *configuration {
	return &configuration{
		wrapText:              false,
		footerEnabled:         true,
		continuationIndicator: "...",
		saveDir:               "",
		saveKey:               key.NewBinding(),
	}
}
