package viewport

// configuration consolidates all configuration options for the viewport
type configuration struct {
	// wrapText is true if the viewport wraps text rather than showing that a line is truncated/horizontally scrollable
	wrapText bool

	// footerEnabled is true if the viewport currently shows the footer based on its dimensions and content
	footerEnabled bool

	// continuationIndicator is the string to use to indicate that an unwrapped line continues to the left or right
	continuationIndicator string
}

// newConfiguration creates a new configuration with default settings.
func newConfiguration() *configuration {
	return &configuration{
		wrapText:              false,
		footerEnabled:         true,
		continuationIndicator: "...",
	}
}
