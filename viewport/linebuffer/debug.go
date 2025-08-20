package linebuffer

import (
	"fmt"
)

func main() {
	lbs := []LineBufferer{
		New("hello world"),
		NewMulti(New("hello world")),
		NewMulti(
			New("hello"),
			New(" world"),
		),
		NewMulti(
			New("hel"),
			New("lo "),
			New("wo"),
			New("rld"),
		),
		NewMulti(
			New("h"),
			New("e"),
			New("l"),
			New("l"),
			New("o"),
			New(" "),
			New("w"),
			New("o"),
			New("r"),
			New("l"),
			New("d"),
		),
	}

	widthToLeft := 1
	takeWidth := 7
	continuation := "..."
	toHighlight := "lo "
	highlightStyle := redBg
	expected := "...o..."
	for _, lb := range lbs {
		var highlights []Highlight
		highlights = ExtractHighlights([]string{lb.Content()}, toHighlight, highlightStyle)
		actual, _ := lb.Take(widthToLeft, takeWidth, continuation, highlights)
		println(fmt.Sprintf("for %s, expected %q, got %q"), lb.Repr(), expected, actual)
	}
}
