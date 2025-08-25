package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
)

func main() {
	lbs := []linebuffer.LineBufferer{
		linebuffer.New("hello world"),
		linebuffer.NewMulti(linebuffer.New("hello world")),
		linebuffer.NewMulti(
			linebuffer.New("hello"),
			linebuffer.New(" world"),
		),
		linebuffer.NewMulti(
			linebuffer.New("hel"),
			linebuffer.New("lo "),
			linebuffer.New("wo"),
			linebuffer.New("rld"),
		),
		linebuffer.NewMulti(
			linebuffer.New("h"),
			linebuffer.New("e"),
			linebuffer.New("l"),
			linebuffer.New("l"),
			linebuffer.New("o"),
			linebuffer.New(" "),
			linebuffer.New("w"),
			linebuffer.New("o"),
			linebuffer.New("r"),
			linebuffer.New("l"),
			linebuffer.New("d"),
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
		highlights := linebuffer.ExtractHighlights([]string{lb.Content()}, toHighlight, highlightStyle)
		//println(fmt.Sprintf("for %s, highlights: %v", lb.Repr(), highlights))
		actual, metadata := lb.Take(widthToLeft, takeWidth, continuation)
		actual = linebuffer.HighlightString(actual, highlights, lb.PlainContent(), metadata.StartByte, metadata.EndByte)
		if actual != expected {
			println(fmt.Sprintf("for %s, expected %q, got %q", lb.Repr(), expected, actual))
		} else {
			println(fmt.Sprintf("for %s, got expected %q", lb.Repr(), actual))
		}
	}
}
