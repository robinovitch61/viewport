package item

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/robinovitch61/bubbleo/internal"

	"github.com/charmbracelet/lipgloss"
)

func getEquivalentItems() map[string][]Item {
	return map[string][]Item{
		"none": {},
		"hello world": {
			NewItem("hello world"),
			NewMulti(NewItem("hello world")),
			NewMulti(
				NewItem("hello"),
				NewItem(" world"),
			),
			NewMulti(
				NewItem("hel"),
				NewItem("lo "),
				NewItem("wo"),
				NewItem("rld"),
			),
			NewMulti(
				NewItem("h"),
				NewItem("e"),
				NewItem("l"),
				NewItem("l"),
				NewItem("o"),
				NewItem(" "),
				NewItem("w"),
				NewItem("o"),
				NewItem("r"),
				NewItem("l"),
				NewItem("d"),
			),
		},
		"ansi": {
			NewItem(internal.RedBg.Render("hello") + " " + internal.BlueBg.Render("world")),
			NewMulti(NewItem(internal.RedBg.Render("hello") + " " + internal.BlueBg.Render("world"))),
			NewMulti(
				NewItem(internal.RedBg.Render("hello")+" "),
				NewItem(internal.BlueBg.Render("world")),
			),
			NewMulti(
				NewItem(internal.RedBg.Render("hello")),
				NewItem(" "),
				NewItem(internal.BlueBg.Render("world")),
			),
		},
		"unicode_ansi": {
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
			NewItem(internal.RedBg.Render("A💖") + "中é"),
			NewMulti(NewItem(internal.RedBg.Render("A💖") + "中é")),
			NewMulti(
				NewItem(internal.RedBg.Render("A💖")),
				NewItem("中"),
				NewItem("é"),
			),
		}}
}

func TestMultiItem_Width(t *testing.T) {
	for _, eq := range getEquivalentItems() {
		for _, item := range eq {
			if item.Width() != eq[0].Width() {
				t.Errorf("expected %d, got %d for item %s", eq[0].Width(), item.Width(), item.repr())
			}
		}
	}
}

func TestMultiItem_Content(t *testing.T) {
	for _, eq := range getEquivalentItems() {
		for _, item := range eq {
			if item.Content() != eq[0].Content() {
				t.Errorf("expected %q, got %q for item %s", eq[0].Content(), item.Content(), item.repr())
			}
		}
	}
}

func TestMultiItem_Take(t *testing.T) {
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
			highlightStyle: internal.RedBg,
			expected:       internal.RedBg.Render("hello") + " world",
		},
		{
			name:           "hello world with highlight across boundary",
			key:            "hello world",
			widthToLeft:    3,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: internal.RedBg,
			expected:       internal.RedBg.Render("lo wo") + "r",
		},
		{
			name:           "hello world with highlight and middle continuation",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "..",
			toHighlight:    "lo ",
			highlightStyle: internal.RedBg,
			expected:       ".." + internal.RedBg.Render("lo ") + "..",
		},
		{
			name:           "hello world with highlight and overlapping continuation",
			key:            "hello world",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "lo ",
			highlightStyle: internal.RedBg,
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
			expected:       internal.RedBg.Render("hello") + " " + internal.BlueBg.Render("w"),
		},
		{
			name:           "ansi start at 1",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       internal.RedBg.Render("ello") + " " + internal.BlueBg.Render("wo"),
		},
		{
			name:           "ansi end",
			key:            "ansi",
			widthToLeft:    10,
			takeWidth:      3,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       internal.BlueBg.Render("d"),
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
			expected:       internal.RedBg.Render("hell.") + "." + internal.BlueBg.Render("."),
		},
		{
			name:           "ansi with continuation at start",
			key:            "ansi",
			widthToLeft:    4,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       internal.RedBg.Render(".") + "." + internal.BlueBg.Render(".orld"),
		},
		{
			name:           "ansi with continuation both ends",
			key:            "ansi",
			widthToLeft:    2,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       internal.RedBg.Render("...") + " " + internal.BlueBg.Render("..."),
		},
		{
			name:           "ansi with highlight whole word",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "hello",
			highlightStyle: internal.GreenBg,
			expected:       internal.GreenBg.Render("hello") + " " + internal.BlueBg.Render("world"),
		},
		{
			name:           "ansi with highlight partial word",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "ell",
			highlightStyle: internal.GreenBg,
			expected:       internal.RedBg.Render("h") + internal.GreenBg.Render("ell") + internal.RedBg.Render("o") + " " + internal.BlueBg.Render("world"),
		},
		{
			name:           "ansi with highlight across boundary",
			key:            "ansi",
			widthToLeft:    0,
			takeWidth:      11,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: internal.GreenBg,
			expected:       internal.RedBg.Render("hel") + internal.GreenBg.Render("lo wo") + internal.BlueBg.Render("rld"),
		},
		{
			name:           "ansi with highlight and middle continuation",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "..",
			toHighlight:    "lo ",
			highlightStyle: internal.GreenBg,
			expected:       internal.RedBg.Render("..") + internal.GreenBg.Render("lo ") + internal.BlueBg.Render(".."),
		},
		{
			name:           "ansi with highlight and overlapping continuation",
			key:            "ansi",
			widthToLeft:    1,
			takeWidth:      7,
			continuation:   "...",
			toHighlight:    "lo ",
			highlightStyle: internal.GreenBg,
			expected:       internal.RedBg.Render("..") + internal.GreenBg.Render(".o.") + internal.BlueBg.Render(".."),
		},
		{
			name:           "unicode_ansi start at 0",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       internal.RedBg.Render("A💖") + "中é",
		},
		{
			name:           "unicode_ansi start at 1",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       internal.RedBg.Render("💖") + "中é",
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
			expected:       internal.RedBg.Render("A💖") + "..", // bit of an edge cases, seems fine
		},
		{
			name:           "unicode_ansi with continuation at start",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "...",
			toHighlight:    "",
			highlightStyle: lipgloss.NewStyle(),
			expected:       internal.RedBg.Render("..") + "中é",
		},
		{
			name:           "unicode_ansi with highlight whole word",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "A💖",
			highlightStyle: internal.GreenBg,
			expected:       internal.GreenBg.Render("A💖") + "中é",
		},
		{
			name:           "unicode_ansi with highlight partial word",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "A",
			highlightStyle: internal.GreenBg,
			expected:       internal.GreenBg.Render("A") + internal.RedBg.Render("💖") + "中é",
		},
		{
			name:           "unicode_ansi with highlight across boundary",
			key:            "unicode_ansi",
			widthToLeft:    0,
			takeWidth:      6,
			continuation:   "",
			toHighlight:    "💖中",
			highlightStyle: internal.GreenBg,
			expected:       internal.RedBg.Render("A") + internal.GreenBg.Render("💖中") + "é",
		},
		{
			name:           "unicode_ansi with highlight and overlapping continuation",
			key:            "unicode_ansi",
			widthToLeft:    1,
			takeWidth:      5,
			continuation:   "..",
			toHighlight:    "💖",
			highlightStyle: internal.GreenBg,
			expected:       internal.GreenBg.Render("..") + "中é",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range getEquivalentItems()[tt.key] {
				byteRanges := eq.ExtractExactMatches(tt.toHighlight)
				highlights := toHighlights(byteRanges, tt.highlightStyle)
				actual, _ := eq.Take(tt.widthToLeft, tt.takeWidth, tt.continuation, highlights)
				internal.CmpStr(t, tt.expected, actual, fmt.Sprintf("for %s", eq.repr()))
			}
		})
	}
}

func TestMultiItem_NumWrappedLines(t *testing.T) {
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
			for _, eq := range getEquivalentItems()[tt.key] {
				actual := eq.NumWrappedLines(tt.wrapWidth)
				if actual != tt.expected {
					t.Errorf("expected %d, got %d for item %s with wrap width %d", tt.expected, actual, eq.repr(), tt.wrapWidth)
				}
			}
		})
	}
}

func TestMultiItem_ExtractExactMatches(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		exactMatch string
		expected   []Match
	}{
		{
			name:       "hello world empty exact match",
			key:        "hello world",
			exactMatch: "",
			expected:   []Match{},
		},
		{
			name:       "hello world no matches",
			key:        "hello world",
			exactMatch: "xyz",
			expected:   []Match{},
		},
		{
			name:       "hello world single match hello",
			key:        "hello world",
			exactMatch: "hello",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   5,
					},
				},
			},
		},
		{
			name:       "hello world single match world",
			key:        "hello world",
			exactMatch: "world",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 6,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 6,
						End:   11,
					},
				},
			},
		},
		{
			name:       "hello world match full content",
			key:        "hello world",
			exactMatch: "hello world",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   11,
					},
				},
			},
		},
		{
			name:       "hello world partial match lo wo",
			key:        "hello world",
			exactMatch: "lo wo",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   8,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   8,
					},
				},
			},
		},
		{
			name:       "hello world single character match l",
			key:        "hello world",
			exactMatch: "l",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 2,
						End:   3,
					},
					WidthRange: WidthRange{
						Start: 2,
						End:   3,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   4,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   4,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 9,
						End:   10,
					},
					WidthRange: WidthRange{
						Start: 9,
						End:   10,
					},
				},
			},
		},
		{
			name:       "hello world overlapping matches ll",
			key:        "hello world",
			exactMatch: "ll",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 2,
						End:   4,
					},
					WidthRange: WidthRange{
						Start: 2,
						End:   4,
					},
				},
			},
		},
		{
			name:       "hello world case sensitive Hello",
			key:        "hello world",
			exactMatch: "Hello",
			expected:   []Match{},
		},
		{
			name:       "ansi match hello",
			key:        "ansi",
			exactMatch: "hello",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   5,
					},
				},
			},
		},
		{
			name:       "ansi match world",
			key:        "ansi",
			exactMatch: "world",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 6,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 6,
						End:   11,
					},
				},
			},
		},
		{
			name:       "ansi match across boundary lo wo",
			key:        "ansi",
			exactMatch: "lo wo",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   8,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   8,
					},
				},
			},
		},
		{
			name:       "unicode_ansi match A💖",
			key:        "unicode_ansi",
			exactMatch: "A💖",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   3,
					},
				},
			},
		},
		{
			name:       "unicode_ansi match 中é",
			key:        "unicode_ansi",
			exactMatch: "中é",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 5,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   6,
					},
				},
			},
		},
		{
			name:       "unicode_ansi match single character A",
			key:        "unicode_ansi",
			exactMatch: "A",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   1,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   1,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, eq := range getEquivalentItems()[tt.key] {
				matches := eq.ExtractExactMatches(tt.exactMatch)

				if len(matches) != len(tt.expected) {
					t.Errorf("for item %s: expected %d matches, got %d", eq.repr(), len(tt.expected), len(matches))
					return
				}

				for i, expected := range tt.expected {
					match := matches[i]

					if match.ByteRange.Start != expected.ByteRange.Start || match.ByteRange.End != expected.ByteRange.End {
						t.Errorf("for item %s, match %d: expected byte range Start=%d End=%d, got Start=%d End=%d",
							eq.repr(), i, expected.ByteRange.Start, expected.ByteRange.End, match.ByteRange.Start, match.ByteRange.End)
					}

					if match.WidthRange.Start != expected.WidthRange.Start || match.WidthRange.End != expected.WidthRange.End {
						t.Errorf("for item %s, match %d: expected width range Start=%d End=%d, got Start=%d End=%d",
							eq.repr(), i, expected.WidthRange.Start, expected.WidthRange.End, match.WidthRange.Start, match.WidthRange.End)
					}
				}
			}
		})
	}
}

func TestMultiItem_ExtractRegexMatches(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		regexPattern string
		expected     []Match
		expectError  bool
	}{
		{
			name:         "hello world no matches",
			key:          "hello world",
			regexPattern: "xyz",
			expected:     []Match{},
		},
		{
			name:         "hello world simple word match",
			key:          "hello world",
			regexPattern: "world",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 6,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 6,
						End:   11,
					},
				},
			},
		},
		{
			name:         "hello world word boundary match",
			key:          "hello world",
			regexPattern: `\bworld\b`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 6,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 6,
						End:   11,
					},
				},
			},
		},
		{
			name:         "hello world character class match l",
			key:          "hello world",
			regexPattern: `l`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 2,
						End:   3,
					},
					WidthRange: WidthRange{
						Start: 2,
						End:   3,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   4,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   4,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 9,
						End:   10,
					},
					WidthRange: WidthRange{
						Start: 9,
						End:   10,
					},
				},
			},
		},
		{
			name:         "hello world case insensitive pattern",
			key:          "hello world",
			regexPattern: `(?i)HELLO`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   5,
					},
				},
			},
		},
		{
			name:         "hello world across boundary lo wo",
			key:          "hello world",
			regexPattern: `lo wo`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   8,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   8,
					},
				},
			},
		},
		{
			name:         "hello world capturing groups",
			key:          "hello world",
			regexPattern: `(hello) (world)`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   11,
					},
				},
			},
		},
		{
			name:         "hello world dot metacharacter",
			key:          "hello world",
			regexPattern: `l.o`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 2,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 2,
						End:   5,
					},
				},
			},
		},
		{
			name:         "hello world anchored pattern start",
			key:          "hello world",
			regexPattern: `^hello`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   5,
					},
				},
			},
		},
		{
			name:         "hello world anchored pattern end",
			key:          "hello world",
			regexPattern: `world$`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 6,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 6,
						End:   11,
					},
				},
			},
		},
		{
			name:         "ansi match hello",
			key:          "ansi",
			regexPattern: "hello",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   5,
					},
				},
			},
		},
		{
			name:         "ansi match across boundary",
			key:          "ansi",
			regexPattern: "lo wo",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   8,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   8,
					},
				},
			},
		},
		{
			name:         "unicode_ansi match A with unicode",
			key:          "unicode_ansi",
			regexPattern: "A💖",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   3,
					},
				},
			},
		},
		{
			name:         "unicode_ansi match 中é",
			key:          "unicode_ansi",
			regexPattern: "中é",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 5,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 3,
						End:   6,
					},
				},
			},
		},
		{
			name:         "unicode_ansi match unicode across boundary",
			key:          "unicode_ansi",
			regexPattern: "💖中",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 1,
						End:   8,
					},
					WidthRange: WidthRange{
						Start: 1,
						End:   5,
					},
				},
			},
		},
		{
			name:         "unicode_ansi wildcard match",
			key:          "unicode_ansi",
			regexPattern: ".💖",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   5,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   3,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := regexp.Compile(tt.regexPattern)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error compiling regex: %v", err)
				return
			}

			for _, eq := range getEquivalentItems()[tt.key] {
				matches := eq.ExtractRegexMatches(regex)

				if len(matches) != len(tt.expected) {
					t.Errorf("for item %s: expected %d matches, got %d", eq.repr(), len(tt.expected), len(matches))
					return
				}

				for i, expected := range tt.expected {
					match := matches[i]

					if match.ByteRange.Start != expected.ByteRange.Start || match.ByteRange.End != expected.ByteRange.End {
						t.Errorf("for item %s, match %d: expected byte range Start=%d End=%d, got Start=%d End=%d",
							eq.repr(), i, expected.ByteRange.Start, expected.ByteRange.End, match.ByteRange.Start, match.ByteRange.End)
					}

					if match.WidthRange.Start != expected.WidthRange.Start || match.WidthRange.End != expected.WidthRange.End {
						t.Errorf("for item %s, match %d: expected width range Start=%d End=%d, got Start=%d End=%d",
							eq.repr(), i, expected.WidthRange.Start, expected.WidthRange.End, match.WidthRange.Start, match.WidthRange.End)
					}
				}
			}
		})
	}
}

func toHighlights(matches []Match, style lipgloss.Style) []Highlight {
	var highlights []Highlight
	for _, match := range matches {
		highlights = append(highlights, Highlight{
			ByteRangeUnstyledContent: match.ByteRange,
			Style:                    style,
		})
	}
	return highlights
}
