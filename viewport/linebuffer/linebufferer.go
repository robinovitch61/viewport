package linebuffer

// TakenMetaData contains metadata about what was taken from the buffer
type TakenMetaData struct {
	Width     int // actual width taken in terminal cells
	StartByte int // start byte offset in original styled content
	EndByte   int // end byte offset in original styled content
}

// LineBufferer defines the interface for line buffer implementations.
type LineBufferer interface {
	// Width returns the total width in terminal cells
	Width() int
	// Content returns the underlying complete string
	Content() string
	// PlainContent returns the content without ANSI codes
	PlainContent() string
	// Take takes a substring of the content with a specified widthToLeft and taking takeWidth.
	// continuation replaces the start and end if the content exceeds the bounds of widthToLeft to widthToLeft + takeWidth.
	// Take returns the taken content and metadata about what was taken.
	Take(
		widthToLeft,
		takeWidth int,
		continuation string,
	) (string, TakenMetaData)
	// WrappedLinesWithoutHighlights returns the content as a slice of strings, wrapping at width, without applying highlights.
	// This is optimized for layout calculations that don't need styling.
	WrappedLinesWithoutHighlights(width int, maxLinesEachEnd int) []string
	// WrappedLines returns the content as a slice of strings, wrapping at width.
	// maxLinesEachEnd is the maximum number of lines to return from the beginning and end of the content.
	// highlights are applied to the content
	WrappedLines(width int, maxLinesEachEnd int, highlights []Highlight) []string
	// Repr returns a representation of the Linebufferer as a string for debugging
	Repr() string
}
