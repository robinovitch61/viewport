package item

// Item defines the interface for item implementations
type Item interface {
	// Width returns the total width in terminal cells
	Width() int

	// Content returns the underlying complete string
	Content() string

	// Take takes a substring (line) of the content with a specified widthToLeft and taking takeWidth.
	// continuation replaces the start and end if the content exceeds the bounds.
	// highlights is a list of highlights to apply to the taken content.
	// Returns the line and the actual width taken
	// TODO LEO: figure out how to make this private while still using it in viewport and filterableviewport packages
	Take(
		widthToLeft,
		takeWidth int,
		continuation string,
		highlights []Highlight,
	) (string, int)

	// NumWrappedLines returns the number of wrapped lines given a wrap width
	// TODO LEO: make this private
	NumWrappedLines(wrapWidth int) int

	// repr returns a representation of the object as a string for debugging
	repr() string
}
