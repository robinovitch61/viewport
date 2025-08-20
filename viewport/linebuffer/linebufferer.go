package linebuffer

// LineBufferer defines the interface for line buffer implementations.
type LineBufferer interface {
	// Width returns the total width in terminal cells
	Width() int
	// Content returns the underlying complete string
	Content() string
	// Take returns a substring of the content with a specified widthToLeft and taking takeWidth
	// continuation replaces the start and end if the content exceeds the bounds of widthToLeft to widthToLeft + takeWidth
	// toHighlight is a substring to highlight, and highlightStyle is the style to apply to it
	Take(
		widthToLeft,
		takeWidth int,
		continuation string,
		highlights []Highlight,
	) (string, int)
	// WrappedLines returns the content as a slice of strings, wrapping at width
	// maxLinesEachEnd is the maximum number of lines to return from the beginning and end of the content
	// toHighlight is a substring to highlight, and highlightStyle is the style to apply to it
	WrappedLines(width int, maxLinesEachEnd int, highlights []Highlight) []string
	// Matches returns true if the content contains the given string, ignoring ansi styling
	Matches(s string) bool
	// MatchesRegex returns true if the content matches the given regex pattern, ignoring ansi styling
	//MatchesRegex(r regexp.Regexp) bool
	// Repr returns a representation of the Linebufferer as a string for debugging
	Repr() string
}
