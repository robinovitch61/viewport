package viewport

import (
	"fmt"
	"testing"

	"github.com/robinovitch61/bubbleo/internal"

	"github.com/charmbracelet/lipgloss"
)

func getEquivalentLineBuffers() map[string][]Item {
	return map[string][]Item{
		"none": {},
		"hello world": {
			NewLineBuffer("hello world"),
			NewMulti(NewLineBuffer("hello world")),
			NewMulti(
				NewLineBuffer("hello"),
				NewLineBuffer(" world"),
			),
			NewMulti(
				NewLineBuffer("hel"),
				NewLineBuffer("lo "),
				NewLineBuffer("wo"),
				NewLineBuffer("rld"),
			),
			NewMulti(
				NewLineBuffer("h"),
				NewLineBuffer("e"),
				NewLineBuffer("l"),
				NewLineBuffer("l"),
				NewLineBuffer("o"),
				NewLineBuffer(" "),
				NewLineBuffer("w"),
				NewLineBuffer("o"),
				NewLineBuffer("r"),
				NewLineBuffer("l"),
				NewLineBuffer("d"),
			),
		},
		"ansi": {
			NewLineBuffer(redBg.Render("hello") + " " + blueBg.Render("world")),
			NewMulti(NewLineBuffer(redBg.Render("hello") + " " + blueBg.Render("world"))),
			NewMulti(
				NewLineBuffer(redBg.Render("hello")+" "),
				NewLineBuffer(blueBg.Render("world")),
			),
			NewMulti(
				NewLineBuffer(redBg.Render("hello")),
				NewLineBuffer(" "),
				NewLineBuffer(blueBg.Render("world")),
			),
		},
		"unicode_ansi": {
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
			NewLineBuffer(redBg.Render("A💖") + "中é"),
			NewMulti(NewLineBuffer(redBg.Render("A💖") + "中é")),
			NewMulti(
				NewLineBuffer(redBg.Render("A💖")),
				NewLineBuffer("中"),
				NewLineBuffer("é"),
			),
		}}
}

func TestMultiLineBuffer_Width(t *testing.T) {
	for _, eq := range getEquivalentLineBuffers() {
		for _, lb := range eq {
			if lb.Width() != eq[0].Width() {
				t.Errorf("expected %d, got %d for line buffer %s", eq[0].Width(), lb.Width(), lb.Repr())
			}
		}
	}
}

func TestMultiLineBuffer_Content(t *testing.T) {
	for _, eq := range getEquivalentLineBuffers() {
		for _, lb := range eq {
			if lb.Content() != eq[0].Content() {
				t.Errorf("expected %q, got %q for line buffer %s", eq[0].Content(), lb.Content(), lb.Repr())
			}
		}
	}
}

func TestMultiLineBuffer_Take(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		widthToLeft    int
		takeWidth      int
		continuation   string
		toHighlight    string
		highlightStyle lipgloss.Style
		expected       string
	}{
		{
			name:           "hello world start at 0",
			key:            "hello world",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "hello w",
		},
		{
			name:           "hello world start at 1",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "ello wo",
		},
		{
			name:           "hello world end",
			key:            "hello world",
			widthToLeft:    10,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "d",
		},
		{
			name:           "hello world past end",
			key:            "hello world",
			widthToLeft:    11,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "",
		},
		{
			name:           "hello world with continuation at end",
			key:            "hello world",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "hell...",
		},
		{
			name:           "hello world with continuation at start",
			key:            "hello world",
			widthToLeft:    4,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "...orld",
		},
		{
			name:           "hello world with continuation both ends",
			key:            "hello world",
			widthToLeft:    2,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "... ...",
		},
		{
			name:           "hello world with highlight whole word",
			key:            "hello world",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "hello",
			highlightStyle: redBg,
			expected:       redBg.Render("hello") + " world",
		},
		{
			name:           "hello world with highlight across buffer boundary",
			key:            "hello world",
			widthToLeft:    3,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: redBg,
			expected:       redBg.Render("lo wo") + "r",
		},
		{
			name:           "hello world with highlight and middle continuation",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "..",
			toHighlight:    "lo ",
			highlightStyle: redBg,
			expected:       ".." + redBg.Render("lo ") + "..",
		},
		{
			name:           "hello world with highlight and overlapping continuation",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "lo ",
			highlightStyle: redBg,
			expected:       "..\x1b[48;2;255;0;0m.o.\x1b[0m..",
		},
		{
			name:           "ansi start at 0",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("hello") + " " + blueBg.Render("w"),
		},
		{
			name:           "ansi start at 1",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("ello") + " " + blueBg.Render("wo"),
		},
		{
			name:           "ansi end",
			key:            "ansi",
			widthToLeft:    10,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       blueBg.Render("d"),
		},
		{
			name:           "ansi past end",
			key:            "ansi",
			widthToLeft:    11,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "",
		},
		{
			name:           "ansi with continuation at end",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("hell.") + "." + blueBg.Render("."),
		},
		{
			name:           "ansi with continuation at start",
			key:            "ansi",
			widthToLeft:    4,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render(".") + "." + blueBg.Render(".orld"),
		},
		{
			name:           "ansi with continuation both ends",
			key:            "ansi",
			widthToLeft:    2,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("...") + " " + blueBg.Render("..."),
		},
		{
			name:           "ansi with highlight whole word",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "hello",
			highlightStyle: greenBg,
			expected:       greenBg.Render("hello") + " " + blueBg.Render("world"),
		},
		{
			name:           "ansi with highlight partial word",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "ell",
			highlightStyle: greenBg,
			expected:       redBg.Render("h") + greenBg.Render("ell") + redBg.Render("o") + " " + blueBg.Render("world"),
		},
		{
			name:           "ansi with highlight across buffer boundary",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: greenBg,
			expected:       redBg.Render("hel") + greenBg.Render("lo wo") + blueBg.Render("rld"),
		},
		{
			name:           "ansi with highlight and middle continuation",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "..",
			toHighlight:    "lo ",
			highlightStyle: greenBg,
			expected:       redBg.Render("..") + greenBg.Render("lo ") + blueBg.Render(".."),
		},
		{
			name:           "ansi with highlight and overlapping continuation",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "lo ",
			highlightStyle: greenBg,
			expected:       redBg.Render("..") + greenBg.Render(".o.") + blueBg.Render(".."),
		},
		{
			name:           "unicode_ansi start at 0",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("A💖") + "中é",
		},
		{
			name:           "unicode_ansi start at 1",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("💖") + "中é",
		},
		{
			name:           "unicode_ansi end",
			key:            "unicode_ansi",
			widthToLeft:    5,
			takeWidth:      1,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "é",
		},
		{
			name:           "unicode_ansi past end",
			key:            "unicode_ansi",
			widthToLeft:    6,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       "",
		},
		{
			name:           "unicode_ansi with continuation at end",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      5,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("A💖") + "..", // bit of an edge cases, seems fine
		},
		{
			name:           "unicode_ansi with continuation at start",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       redBg.Render("..") + "中é",
		},
		{
			name:           "unicode_ansi with highlight whole word",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "A💖",
			highlightStyle: greenBg,
			expected:       greenBg.Render("A💖") + "中é",
		},
		{
			name:           "unicode_ansi with highlight partial word",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "A",
			highlightStyle: greenBg,
			expected:       greenBg.Render("A") + redBg.Render("💖") + "中é",
		},
		{
			name:           "unicode_ansi with highlight across buffer boundary",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "💖中",
			highlightStyle: greenBg,
			expected:       redBg.Render("A") + greenBg.Render("💖中") + "é",
		},
		{
			name:           "unicode_ansi with highlight and overlapping continuation",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "..",
			toHighlight:    "💖",
			highlightStyle: greenBg,
			expected:       greenBg.Render("..") + "中é",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range getEquivalentLineBuffers()[tt.key] {
				var highlights []Highlight
				if tt.toHighlight != "" {
					highlights = ExtractHighlights([]string{eq.Content()}, tt.toHighlight, tt.highlightStyle)
				}
				actual, _ := eq.Take(tt.widthToLeft, tt.takeWidth, tt.continuation, highlights)
				internal.CmpStr(t, tt.expected, actual, fmt.Sprintf("for %s", eq.Repr()))
			}
		})
	}
}

func TestMultiLineBuffer_NumWrappedLines(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wrapWidth int
		expected  int
	}{
		{
			name:      "none no width",
			key:       "none",
			wrapWidth: 0,
			expected:  0,
		},
		{
			name:      "none with width",
			key:       "none",
			wrapWidth: 5,
			expected:  1,
		},
		{
			name:      "hello world negative width",
			key:       "hello world", // 11 width
			wrapWidth: -1,
			expected:  0,
		},
		{
			name:      "hello world zero width",
			key:       "hello world", // 11 width
			wrapWidth: 0,
			expected:  0,
		},
		{
			name:      "hello world wrap 1",
			key:       "hello world", // 11 width
			wrapWidth: 1,
			expected:  11,
		},
		{
			name:      "hello world wrap 5",
			key:       "hello world", // 11 width
			wrapWidth: 5,
			expected:  3,
		},
		{
			name:      "hello world wrap 11",
			key:       "hello world", // 11 width
			wrapWidth: 11,
			expected:  1,
		},
		{
			name:      "hello world wrap 12",
			key:       "hello world", // 11 width
			wrapWidth: 12,
			expected:  1,
		},
		{
			name:      "ansi wrap 5",
			key:       "ansi", // 11 width
			wrapWidth: 5,
			expected:  3,
		},
		{
			name:      "unicode_ansi wrap 3",
			key:       "unicode_ansi", // 6 width
			wrapWidth: 3,
			expected:  2,
		},
		{
			name:      "unicode_ansi wrap 6",
			key:       "unicode_ansi", // 6 width
			wrapWidth: 6,
			expected:  1,
		},
		{
			name:      "unicode_ansi wrap 7",
			key:       "unicode_ansi", // 6 width
			wrapWidth: 7,
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range getEquivalentLineBuffers()[tt.key] {
				actual := eq.NumWrappedLines(tt.wrapWidth)
				if actual != tt.expected {
					t.Errorf("expected %d, got %d for line buffer %s with wrap width %d", tt.expected, actual, eq.Repr(), tt.wrapWidth)
				}
			}
		})
	}
}
