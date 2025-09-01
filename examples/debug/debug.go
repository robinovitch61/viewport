package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/robinovitch61/bubbleo/viewport"
)

func main() {
	lbs := []viewport.Item{
		viewport.NewLineBuffer("hello world"),
		viewport.NewMulti(viewport.NewLineBuffer("hello world")),
		viewport.NewMulti(
			viewport.NewLineBuffer("hello"),
			viewport.NewLineBuffer(" world"),
		),
		viewport.NewMulti(
			viewport.NewLineBuffer("hel"),
			viewport.NewLineBuffer("lo "),
			viewport.NewLineBuffer("wo"),
			viewport.NewLineBuffer("rld"),
		),
		viewport.NewMulti(
			viewport.NewLineBuffer("h"),
			viewport.NewLineBuffer("e"),
			viewport.NewLineBuffer("l"),
			viewport.NewLineBuffer("l"),
			viewport.NewLineBuffer("o"),
			viewport.NewLineBuffer(" "),
			viewport.NewLineBuffer("w"),
			viewport.NewLineBuffer("o"),
			viewport.NewLineBuffer("r"),
			viewport.NewLineBuffer("l"),
			viewport.NewLineBuffer("d"),
		),
	}

	widthToLeft := 1
	takeWidth := 7
	continuation := "..."
	toHighlight := "lo "
	lipgloss.SetColorProfile(termenv.TrueColor)
	red := lipgloss.Color("#FF0000")
	redBg := lipgloss.NewStyle().Background(red)
	highlightStyle := redBg

	expected := "..\x1b[48;2;255;0;0m.o.\x1b[0m.."
	for _, lb := range lbs {
		highlights := viewport.ExtractHighlights([]string{lb.Content()}, toHighlight, highlightStyle)
		//println(fmt.Sprintf("for %s, highlights: %v", lb.Repr(), highlights))
		actual, _ := lb.Take(widthToLeft, takeWidth, continuation, highlights)
		if actual != expected {
			println(fmt.Sprintf("for %s, expected %q, got %q", lb.Repr(), expected, actual))
		} else {
			println(fmt.Sprintf("for %s, got expected %q", lb.Repr(), actual))
		}
	}
}
