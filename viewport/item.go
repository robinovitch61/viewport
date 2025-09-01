package viewport

// Item defines the interface for viewport item implementations
type Item interface {
	// Width returns the total width in terminal cells
	Width() int
	// Content returns the underlying complete string
	Content() string
	// Take takes a substring of the content with a specified widthToLeft and taking takeWidth.
	// continuation replaces the start and end if the content exceeds the bounds of widthToLeft to widthToLeft + takeWidth.
	// toHighlight is a substring to highlight, and highlightStyle is the style to apply to it.
	// Take returns the taken content and the actual width taken.
	Take(
		widthToLeft,
		takeWidth int,
		continuation string,
		highlights []Highlight,
	) (string, int)
	// NumWrappedLines returns the number of wrapped lines given a wrap width
	NumWrappedLines(wrapWidth int) int
	// Repr returns a representation of the Linebufferer as a string for debugging
	Repr() string
}
