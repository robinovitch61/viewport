package item

import (
	"regexp"
	"strings"
	"testing"

	"github.com/robinovitch61/bubbleo/internal"

	"github.com/charmbracelet/lipgloss/v2"
)

func TestSingle_Width(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected int
	}{
		{
			name:     "empty",
			s:        "",
			expected: 0,
		},
		{
			name:     "simple",
			s:        "1234567890",
			expected: 10,
		},
		{
			name:     "unicode",
			s:        "世界🌟世界a",
			expected: 11,
		},
		{
			name:     "ansi",
			s:        "\x1b[38;2;255;0;0mhi" + RST,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.s)
			if actual := item.Width(); actual != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, actual)
			}
		})
	}
}

func TestSingle_Content(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		expected string
	}{
		{
			name:     "empty",
			s:        "",
			expected: "",
		},
		{
			name:     "simple",
			s:        "1234567890",
			expected: "1234567890",
		},
		{
			name:     "unicode",
			s:        "世界🌟世界",
			expected: "世界🌟世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.s)
			if actual := item.Content(); actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestSingle_Take(t *testing.T) {
	tests := []struct {
		name           string
		s              string
		width          int
		continuation   string
		toHighlight    string
		highlightStyle lipgloss.Style
		startWidth     int
		numTakes       int
		expected       []string
	}{
		{
			name:         "empty",
			s:            "",
			width:        10,
			continuation: "",
			startWidth:   0,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "simple",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   0,
			numTakes:     1,
			expected:     []string{"1234567890"},
		},
		{
			name:         "negative widthToLeft",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   -1,
			numTakes:     1,
			expected:     []string{"1234567890"},
		},
		{
			name:         "seek",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   3,
			numTakes:     1,
			expected:     []string{"4567890"},
		},
		{
			name:         "seek to end",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   10,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "seek past end",
			s:            "1234567890",
			width:        10,
			continuation: "",
			startWidth:   11,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "continuation",
			s:            "1234567890",
			width:        7,
			continuation: "...",
			startWidth:   2,
			numTakes:     1,
			expected:     []string{"...6..."},
		},
		{
			name:         "continuation past end",
			s:            "1234567890",
			width:        10,
			continuation: "...",
			startWidth:   11,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "unicode",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   0,
			numTakes:     1,
			expected:     []string{"世界🌟世界"},
		},
		{
			name:         "unicode seek past first rune",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   2,
			numTakes:     1,
			expected:     []string{"界🌟世界🌟"},
		},
		{
			name:         "unicode seek past first 2 runes",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   3,
			numTakes:     1,
			expected:     []string{"🌟世界🌟"},
		},
		{
			name:         "unicode seek past all but 1 rune",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   10,
			numTakes:     1,
			expected:     []string{"🌟"},
		},
		{
			name:         "unicode seek almost to end",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   11,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "unicode seek to end",
			s:            "世界🌟世界🌟",
			width:        10,
			continuation: "",
			startWidth:   12,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "unicode insufficient width",
			s:            "世界🌟世界🌟",
			width:        1,
			continuation: "",
			startWidth:   2,
			numTakes:     1,
			expected:     []string{""},
		},
		{
			name:         "no ansi, no continuation, no width",
			s:            "12345678901234",
			width:        0,
			continuation: "",
			numTakes:     3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "no ansi, continuation, no width",
			s:            "12345678901234",
			width:        0,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "no ansi, no continuation, width 1",
			s:            "12345678901234",
			width:        1,
			continuation: "",
			numTakes:     3,
			expected: []string{
				"1",
				"2",
				"3",
			},
		},
		{
			name:         "no ansi, continuation, width 1",
			s:            "12345678901234",
			width:        1,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				".",
				".",
				".",
			},
		},
		{
			name:         "no ansi, no continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "",
			numTakes:     4,
			expected: []string{
				"12345",
				"67890",
				"1234",
				"",
			},
		},
		{
			name:         "no ansi, continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "...",
			numTakes:     4,
			expected: []string{
				"12...",
				".....",
				"...4",
				"",
			},
		},
		{
			name:         "no ansi, no continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "",
			numTakes:     4,
			expected: []string{
				"12345",
				"67890",
				"1234",
				"",
			},
		},
		{
			name:         "no ansi, continuation",
			s:            "12345678901234",
			width:        5,
			continuation: "...",
			numTakes:     4,
			expected: []string{
				"12...",
				".....",
				"...4",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, no width",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        0,
			continuation: "",
			numTakes:     3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, no width",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        0,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 1",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        1,
			continuation: "",
			numTakes:     3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 1",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        1,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"",
				"",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 2",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        2,
			continuation: "",
			numTakes:     4,
			expected: []string{
				"世",
				"界",
				"🌟",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 2",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        2,
			continuation: "...",
			numTakes:     4,
			expected: []string{
				"..",
				"..",
				"..",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 3",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        3,
			continuation: "",
			numTakes:     4,
			expected: []string{
				"世",
				"界",
				"🌟",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 3",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        3,
			continuation: "...",
			numTakes:     4,
			expected: []string{
				"..",
				"..",
				"..",
				"",
			},
		},
		{
			name:         "double width unicode, no continuation, width 4",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        4,
			continuation: "",
			numTakes:     3,
			expected: []string{
				"世界",
				"🌟",
				"",
			},
		},
		{
			name:         "double width unicode, continuation, width 3",
			s:            "世界🌟", // each of these takes up 2 terminal cells
			width:        4,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"世..",
				"..",
				"",
			},
		},
		{
			name:         "width equal to continuation",
			s:            "1234567890",
			width:        3,
			continuation: "...",
			numTakes:     4,
			expected: []string{
				"...",
				"...",
				"...",
				".",
			},
		},
		{
			name:         "width slightly bigger than continuation",
			s:            "1234567890",
			width:        4,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"1...",
				"....",
				"..",
			},
		},
		{
			name:         "width double continuation 1",
			s:            "123456789012345678",
			width:        6,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"123...",
				"......",
				"...678",
			},
		},
		{
			name:         "width double continuation 2",
			s:            "1234567890123456789",
			width:        6,
			continuation: "...",
			numTakes:     4,
			expected: []string{
				"123...",
				"......",
				"......",
				".",
			},
		},
		{
			name:         "small string",
			s:            "hi",
			width:        3,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"hi"},
		},
		{
			name:         "continuation longer than width",
			s:            "1234567890123456789012345",
			width:        1,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"."},
		},
		{
			name:         "twice the continuation longer than width",
			s:            "1234567",
			width:        5,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"12..."},
		},
		{
			name:         "sufficient width",
			s:            "1234567890123456789012345",
			width:        30,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"1234567890123456789012345"},
		},
		{
			name:         "sufficient width, space at end preserved",
			s:            "1234567890123456789012345     ",
			width:        30,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"1234567890123456789012345     "},
		},
		{
			name:         "insufficient width",
			s:            "1234567890123456789012345",
			width:        15,
			continuation: "...",
			numTakes:     1,
			expected:     []string{"123456789012..."},
		},
		{
			name:         "insufficient width",
			s:            "123456789012345678901234567890123456789012345",
			width:        15,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"123456789012...",
				"...901234567...",
				"...456789012345",
			},
		},
		{
			name:         "ansi simple, no continuation",
			s:            "\x1b[38;2;255;0;0ma really really long line" + RST,
			width:        15,
			continuation: "",
			numTakes:     2,
			expected: []string{
				"\x1b[38;2;255;0;0ma really really" + RST,
				"\x1b[38;2;255;0;0m long line" + RST,
			},
		},
		{
			name:         "ansi simple, continuation",
			s:            "\x1b[38;2;255;0;0m12345678901234567890123456789012345" + RST,
			width:        15,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"\x1b[38;2;255;0;0m123456789012..." + RST,
				"\x1b[38;2;255;0;0m...901234567..." + RST,
				"\x1b[38;2;255;0;0m...45" + RST,
			},
		},
		{
			name:         "inline ansi, no continuation",
			s:            "\x1b[38;2;255;0;0ma" + RST + " really really long line",
			width:        15,
			continuation: "",
			numTakes:     2,
			expected: []string{
				"\x1b[38;2;255;0;0ma" + RST + " really really",
				" long line",
			},
		},
		{
			name:         "inline ansi, continuation",
			s:            "|\x1b[38;2;169;15;15mfl..-1" + RST + "| {\"timestamp\": \"now\"}",
			width:        15,
			continuation: "...",
			numTakes:     3,
			expected: []string{
				"|\x1b[38;2;169;15;15mfl..-1" + RST + "| {\"t...",
				"...mp\": \"now\"}",
				"",
			},
		},
		{
			name:         "ansi short",
			s:            "\x1b[38;2;0;0;255mhi" + RST,
			width:        3,
			continuation: "...",
			numTakes:     1,
			expected: []string{
				"\x1b[38;2;0;0;255mhi" + RST,
			},
		},
		{
			name:         "multi-byte runes",
			s:            "├─flask",
			width:        6,
			continuation: "...",
			numTakes:     1,
			expected: []string{
				"├─f...",
			},
		},
		{
			name:         "multi-byte runes with ansi and continuation",
			s:            "\x1b[38;2;0;0;255m├─flask" + RST,
			width:        6,
			continuation: "...",
			numTakes:     1,
			expected: []string{
				"\x1b[38;2;0;0;255m├─f..." + RST,
			},
		},
		{
			name:         "width exceeds capacity",
			s:            "  │   └─[ ] local-path-provisioner (running for 11d)",
			width:        53,
			continuation: "",
			numTakes:     1,
			expected: []string{
				"  │   └─[ ] local-path-provisioner (running for 11d)",
			},
		},
		{
			name:           "toHighlight, no continuation, no overflow, no ansi",
			s:              "a very normal log",
			width:          15,
			continuation:   "",
			toHighlight:    "very",
			highlightStyle: internal.RedBg,
			numTakes:       1,
			expected: []string{
				"a " + internal.RedBg.Render("very") + " normal l",
			},
		},
		{
			name:           "toHighlight, no continuation, no overflow, no ansi",
			s:              "a very normal log",
			width:          15,
			continuation:   "",
			toHighlight:    "very",
			highlightStyle: internal.RedBg,
			numTakes:       1,
			expected: []string{
				"a " + internal.RedBg.Render("very") + " normal l",
			},
		},
		{
			name:           "toHighlight, continuation, no overflow, no ansi",
			s:              "a very normal log",
			width:          15,
			continuation:   "...",
			toHighlight:    "l l",
			highlightStyle: internal.RedBg,
			numTakes:       1,
			expected: []string{
				"a very norma\x1b[48;2;255;0;0m..." + RST,
			},
		},
		{
			name:           "toHighlight, another continuation, no overflow, no ansi",
			s:              "a very normal log",
			width:          15,
			continuation:   "...",
			toHighlight:    "very",
			highlightStyle: internal.RedBg,
			startWidth:     1,
			numTakes:       1,
			expected: []string{
				".\x1b[48;2;255;0;0m..ry" + RST + " normal...",
			},
		},
		{
			name:           "toHighlight, no continuation, no overflow, no ansi, many matches",
			s:              strings.Repeat("r", 10),
			width:          6,
			continuation:   "",
			toHighlight:    "r",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				strings.Repeat("\x1b[48;2;255;0;0mr"+RST+"", 6),
				strings.Repeat("\x1b[48;2;255;0;0mr"+RST+"", 4),
			},
		},
		{
			name:           "toHighlight, no continuation, no overflow, ansi",
			s:              "\x1b[38;2;0;0;255mhi \x1b[48;2;0;255;0mthere" + RST + " er",
			width:          15,
			continuation:   "",
			toHighlight:    "er",
			highlightStyle: internal.RedBg,
			numTakes:       1,
			expected: []string{
				"\x1b[38;2;0;0;255mhi \x1b[48;2;0;255;0mth" + RST + "\x1b[48;2;255;0;0mer" + RST + "\x1b[38;2;0;0;255m\x1b[48;2;0;255;0me" + RST + " \x1b[48;2;255;0;0mer" + RST,
			},
		},
		{
			name:           "toHighlight, no continuation, overflows left and right, no ansi",
			s:              "hi there re",
			width:          6,
			continuation:   "",
			toHighlight:    "hi there",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				internal.RedBg.Render("hi the"),
				internal.RedBg.Render("re") + " re",
			},
		},
		{
			name:           "toHighlight, no continuation, overflows left and right, ansi",
			s:              "\x1b[38;2;0;0;255mhi there re" + RST,
			width:          6,
			continuation:   "",
			toHighlight:    "hi there",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				"\x1b[48;2;255;0;0mhi the" + RST,
				"\x1b[48;2;255;0;0mre" + RST + "\x1b[38;2;0;0;255m re" + RST,
			},
		},
		{
			name:           "toHighlight, no continuation, another ansi",
			s:              internal.RedBg.Render("hello") + " " + internal.BlueBg.Render("world"),
			width:          11,
			continuation:   "",
			toHighlight:    "lo wo",
			highlightStyle: internal.GreenBg,
			numTakes:       1,
			expected: []string{
				internal.RedBg.Render("hel") + internal.GreenBg.Render("lo wo") + internal.BlueBg.Render("rld"),
			},
		},
		{
			name:           "toHighlight, no continuation, overflows left and right one char, no ansi",
			s:              "hi there re",
			width:          7,
			continuation:   "",
			toHighlight:    "hi there",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				internal.RedBg.Render("hi ther"),
				internal.RedBg.Render("e") + " re",
			},
		},
		{
			name:           "unicode toHighlight, no continuation, no overflow, no ansi",
			s:              "世界🌟世界🌟",
			width:          7,
			continuation:   "",
			toHighlight:    "世界",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				internal.RedBg.Render("世界") + "🌟",
				internal.RedBg.Render("世界") + "🌟",
			},
		},
		{
			name:           "unicode toHighlight, no continuation, overflow, no ansi",
			s:              "世界🌟世界🌟",
			width:          7,
			continuation:   "",
			toHighlight:    "世界🌟世",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				internal.RedBg.Render("世界🌟"),
				internal.RedBg.Render("世") + "界🌟",
			},
		},
		{
			name:           "unicode toHighlight, no continuation, overflow, ansi",
			s:              "\x1b[38;2;0;0;255m世界🌟世界🌟" + RST,
			width:          7,
			continuation:   "",
			toHighlight:    "世界🌟世",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				internal.RedBg.Render("世界🌟"),
				internal.RedBg.Render("世") + "\x1b[38;2;0;0;255m界🌟" + RST,
			},
		},
		{
			name:           "unicode toHighlight, continuation, overflow, ansi",
			s:              "\x1b[38;2;0;0;255m世界🌟世界🌟" + RST,
			width:          7,
			continuation:   "...",
			toHighlight:    "世界🌟世",
			highlightStyle: internal.RedBg,
			numTakes:       2,
			expected: []string{
				"\x1b[48;2;255;0;0m世界.." + RST,
				"\x1b[48;2;255;0;0m.." + RST + "\x1b[38;2;0;0;255m界🌟" + RST,
			},
		},
		{
			name: "unicode with heart exact width",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
			s:            "A💖中é",
			width:        6,
			continuation: "",
			startWidth:   0,
			numTakes:     1,
			expected:     []string{"A💖中é"},
		},
		{
			name: "unicode with heart start continuation",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
			s:            "A💖中é",
			width:        5,
			continuation: "...",
			startWidth:   1,
			numTakes:     1,
			expected:     []string{"..中é"},
		},
		{
			name: "unicode with heart start continuation and ansi",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
			s:            internal.RedBg.Render("A💖") + "中é",
			width:        5,
			continuation: "...",
			startWidth:   1,
			numTakes:     1,
			expected:     []string{internal.RedBg.Render("..") + "中é"},
		},
		{
			name: "unicode combining",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
			s:            "A💖中éA💖中é", // 12w total
			width:        10,
			continuation: "",
			numTakes:     2,
			expected: []string{
				"A💖中éA💖",
				"中é",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.expected) != tt.numTakes {
				t.Fatalf("num expected != num popLefts")
			}
			item := NewItem(tt.s)
			startWidth := tt.startWidth

			byteRanges := item.ExtractExactMatches(tt.toHighlight)
			highlights := toHighlights(byteRanges, tt.highlightStyle)
			for i := 0; i < tt.numTakes; i++ {
				actual, actualWidth := item.Take(startWidth, tt.width, tt.continuation, highlights)
				internal.CmpStr(t, tt.expected[i], actual)
				startWidth += actualWidth
			}
		})
	}
}

func TestSingle_Take_NoAnsiLeak(t *testing.T) {
	// simulates git diff syntax-highlighted output where " is one color and \b another.
	// when highlighting ", ANSI code internals like "38;2;190;132;255m" must not
	// leak as visible text.
	s := "\x1b[38;2;204;204;204m " + RST +
		"\x1b[38;2;152;195;121m\"" + RST +
		"\x1b[38;2;190;132;255m\\b" + RST +
		"\x1b[38;2;152;195;121m\"" + RST +
		"\x1b[38;2;204;204;204m " + RST

	item := NewItem(s)
	byteRanges := item.ExtractExactMatches("\"")
	highlights := toHighlights(byteRanges, internal.RedBg)

	actual, _ := item.Take(0, 80, "", highlights)
	stripped := stripAnsi(actual)
	plain := stripAnsi(s)
	if stripped != plain {
		t.Errorf("ANSI leak detected: stripAnsi(result) = %q, want %q", stripped, plain)
	}
}

func TestSingle_NumWrappedLines(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		wrapWidth int
		expected  int
	}{
		{
			name:      "none no width",
			s:         "none",
			wrapWidth: 0,
			expected:  0,
		},
		{
			name:      "none with width",
			s:         "none",
			wrapWidth: 5,
			expected:  1,
		},
		{
			name:      "hello world negative width",
			s:         "hello world", // 11 width
			wrapWidth: -1,
			expected:  0,
		},
		{
			name:      "hello world zero width",
			s:         "hello world", // 11 width
			wrapWidth: 0,
			expected:  0,
		},
		{
			name:      "hello world wrap 1",
			s:         "hello world", // 11 width
			wrapWidth: 1,
			expected:  11,
		},
		{
			name:      "hello world wrap 5",
			s:         "hello world", // 11 width
			wrapWidth: 5,
			expected:  3,
		},
		{
			name:      "hello world wrap 11",
			s:         "hello world", // 11 width
			wrapWidth: 11,
			expected:  1,
		},
		{
			name:      "hello world wrap 12",
			s:         "hello world", // 11 width
			wrapWidth: 12,
			expected:  1,
		},
		{
			name:      "ansi wrap 5",
			s:         internal.RedBg.Render("hello world"), // 11 width
			wrapWidth: 5,
			expected:  3,
		},
		{
			name:      "unicode_ansi wrap 3",
			s:         internal.RedBg.Render("A💖") + "中é", // 6 width
			wrapWidth: 3,
			expected:  2,
		},
		{
			name:      "unicode_ansi wrap 6",
			s:         internal.RedBg.Render("A💖") + "中é", // 6 width
			wrapWidth: 6,
			expected:  1,
		},
		{
			name:      "unicode_ansi wrap 7",
			s:         internal.RedBg.Render("A💖") + "中é", // 6 width
			wrapWidth: 7,
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.s)
			actual := item.NumWrappedLines(tt.wrapWidth)
			if actual != tt.expected {
				t.Errorf("expected %d, got %d for item %s with wrap width %d", tt.expected, actual, item.repr(), tt.wrapWidth)
			}
		})
	}
}

func TestSingleItem_ExtractExactMatches(t *testing.T) {
	tests := []struct {
		name       string
		s          string
		exactMatch string
		expected   []Match
	}{
		{
			name:       "empty exact match",
			s:          "hello world",
			exactMatch: "",
			expected:   []Match{},
		},
		{
			name:       "no matches",
			s:          "hell",
			exactMatch: "lo",
			expected:   []Match{},
		},
		{
			name:       "single match",
			s:          "hello world",
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
			name:       "multiple matches in single string",
			s:          "hello world world",
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
				{
					ByteRange: ByteRange{
						Start: 12,
						End:   17,
					},
					WidthRange: WidthRange{
						Start: 12,
						End:   17,
					},
				},
			},
		},
		{
			name:       "overlapping matches",
			s:          "aaa",
			exactMatch: "aa",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   2,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   2,
					},
				},
			},
		},
		{
			name:       "sequential matches",
			s:          "aaaa",
			exactMatch: "aa",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   2,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   2,
					},
				},
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
			name:       "case sensitive",
			s:          "Hello HELLO hello",
			exactMatch: "hello",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 12,
						End:   17,
					},
					WidthRange: WidthRange{
						Start: 12,
						End:   17,
					},
				},
			},
		},
		{
			name: "unicode characters",
			// 世 is 3 bytes 2 width, 界 is 3 bytes 2 width, 🌟 is 4 bytes 2 width
			s:          "世界 hello 🌟",
			exactMatch: "界 hello 🌟",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   17,
					},
					WidthRange: WidthRange{
						Start: 2,
						End:   13,
					},
				},
			},
		},
		{
			name:       "single character match",
			s:          "abcabc",
			exactMatch: "a",
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
			},
		},
		{
			name:       "match at beginning and end",
			s:          "test middle test",
			exactMatch: "test",
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   4,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   4,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 12,
						End:   16,
					},
					WidthRange: WidthRange{
						Start: 12,
						End:   16,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := NewItem(tt.s).ExtractExactMatches(tt.exactMatch)

			if len(matches) != len(tt.expected) {
				t.Errorf("expected %d matches, got %d", len(tt.expected), len(matches))
				return
			}

			for i, expected := range tt.expected {
				match := matches[i]

				if match.ByteRange.Start != expected.ByteRange.Start || match.ByteRange.End != expected.ByteRange.End {
					t.Errorf("match %d: expected byte range Start=%d End=%d, got Start=%d End=%d",
						i, expected.ByteRange.Start, expected.ByteRange.End, match.ByteRange.Start, match.ByteRange.End)
				}

				if match.WidthRange.Start != expected.WidthRange.Start || match.WidthRange.End != expected.WidthRange.End {
					t.Errorf("match %d: expected width range Start=%d End=%d, got Start=%d End=%d",
						i, expected.WidthRange.Start, expected.WidthRange.End, match.WidthRange.Start, match.WidthRange.End)
				}
			}
		})
	}
}

func TestSingleItem_ExtractRegexMatches(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		regexPattern string
		expected     []Match
		expectError  bool
	}{
		{
			name:         "invalid regex",
			s:            "hello world",
			regexPattern: "[",
			expected:     nil,
			expectError:  true,
		},
		{
			name:         "no matches",
			s:            "hello world",
			regexPattern: "xyz",
			expected:     []Match{},
		},
		{
			name:         "simple word match",
			s:            "hello world",
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
			name:         "word boundary match",
			s:            "hello world worldly",
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
			name:         "digit pattern",
			s:            "line 123 has numbers 456",
			regexPattern: `\d+`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 5,
						End:   8,
					},
					WidthRange: WidthRange{
						Start: 5,
						End:   8,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 21,
						End:   24,
					},
					WidthRange: WidthRange{
						Start: 21,
						End:   24,
					},
				},
			},
		},
		{
			name:         "case insensitive pattern",
			s:            "Hello HELLO hello",
			regexPattern: `(?i)hello`,
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
				{
					ByteRange: ByteRange{
						Start: 12,
						End:   17,
					},
					WidthRange: WidthRange{
						Start: 12,
						End:   17,
					},
				},
			},
		},
		{
			name:         "capturing groups",
			s:            "user: john and user: jane",
			regexPattern: `user: (\w+)`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   10,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   10,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 15,
						End:   25,
					},
					WidthRange: WidthRange{
						Start: 15,
						End:   25,
					},
				},
			},
		},
		{
			name:         "multiple capturing groups",
			s:            "user: john smith and user: jane doe",
			regexPattern: `user: (\w+) (\w+)`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   16,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   16,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 21,
						End:   35,
					},
					WidthRange: WidthRange{
						Start: 21,
						End:   35,
					},
				},
			},
		},
		{
			name:         "dot metacharacter",
			s:            "a1b a.b axb",
			regexPattern: `a.b`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   3,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   3,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 4,
						End:   7,
					},
					WidthRange: WidthRange{
						Start: 4,
						End:   7,
					},
				},
				{
					ByteRange: ByteRange{
						Start: 8,
						End:   11,
					},
					WidthRange: WidthRange{
						Start: 8,
						End:   11,
					},
				},
			},
		},
		{
			name:         "anchored pattern",
			s:            "start middle end",
			regexPattern: `^start`,
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
			name: "unicode with regex",
			// 世 is 3 bytes 2 width, 界 is 3 bytes 2 width, 🌟 is 4 bytes 2 width
			s:            "世界 test 🌟 and test 世界",
			regexPattern: `界 test 🌟`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 3,
						End:   16,
					},
					WidthRange: WidthRange{
						Start: 2,
						End:   12,
					},
				},
			},
		},
		{
			name:         "overlapping matches not possible with regex",
			s:            "aaa",
			regexPattern: `aa`,
			expected: []Match{
				{
					ByteRange: ByteRange{
						Start: 0,
						End:   2,
					},
					WidthRange: WidthRange{
						Start: 0,
						End:   2,
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

			matches := NewItem(tt.s).ExtractRegexMatches(regex)

			if len(matches) != len(tt.expected) {
				t.Errorf("expected %d matches, got %d", len(tt.expected), len(matches))
				return
			}

			for i, expected := range tt.expected {
				match := matches[i]

				if match.ByteRange.Start != expected.ByteRange.Start || match.ByteRange.End != expected.ByteRange.End {
					t.Errorf("match %d: expected byte range Start=%d End=%d, got Start=%d End=%d",
						i, expected.ByteRange.Start, expected.ByteRange.End, match.ByteRange.Start, match.ByteRange.End)
				}

				if match.WidthRange.Start != expected.WidthRange.Start || match.WidthRange.End != expected.WidthRange.End {
					t.Errorf("match %d: expected width range Start=%d End=%d, got Start=%d End=%d",
						i, expected.WidthRange.Start, expected.WidthRange.End, match.WidthRange.Start, match.WidthRange.End)
				}
			}
		})
	}
}

func TestSingle_findRuneIndexWithWidthToLeft(t *testing.T) {
	tests := []struct {
		name            string
		s               string
		widthToLeft     int
		expectedRuneIdx int
		shouldPanic     bool
	}{
		{
			name:            "empty string",
			s:               "",
			widthToLeft:     0,
			expectedRuneIdx: 0,
		},
		{
			name:        "negative widthToLeft",
			s:           "hello",
			widthToLeft: -1,
			shouldPanic: true,
		},
		{
			name:            "single char",
			s:               "a",
			widthToLeft:     1,
			expectedRuneIdx: 1,
		},
		{
			name:            "widthToLeft at end",
			s:               "abc",
			widthToLeft:     3,
			expectedRuneIdx: 3,
		},
		{
			name:        "widthToLeft past total width",
			s:           "a",
			widthToLeft: 2,
			shouldPanic: true,
		},
		{
			name:            "longer",
			s:               "hello",
			widthToLeft:     3,
			expectedRuneIdx: 3,
		},
		{
			name:            "ansi",
			s:               "hi " + internal.RedBg.Render("there") + " leo",
			widthToLeft:     8,
			expectedRuneIdx: 8,
		},
		{
			name: "unicode",
			s:    "A💖中é",
			// A (1w, 1b, 1r), 💖 (2w, 4b, 1r), 中 (2w, 3b, 1r), é (1w, 3b, 2r) = 6w, 11b, 5r
			widthToLeft:     5,
			expectedRuneIdx: 3,
		},
		{
			name: "unicode zero-width",
			s:    "A💖中é",
			// A (1w, 1b, 1r), 💖 (2w, 4b, 1r), 中 (2w, 3b, 1r), é (1w, 3b, 2r) = 6w, 11b, 5r
			widthToLeft:     6,
			expectedRuneIdx: 5,
		},
		{
			name:            "unicode zero-width single char",
			s:               "é",
			widthToLeft:     1,
			expectedRuneIdx: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.s)

			if tt.shouldPanic {
				assertPanic(t, func() {
					item.findRuneIndexWithWidthToLeft(tt.widthToLeft)
				})
				return
			}

			actual := item.findRuneIndexWithWidthToLeft(tt.widthToLeft)
			if actual != tt.expectedRuneIdx {
				t.Errorf("findRuneIndexWithWidthToLeft() got %d, expected %d", actual, tt.expectedRuneIdx)
			}
		})
	}
}

func TestSingle_getByteOffsetAtRuneIdx(t *testing.T) {
	tests := []struct {
		name               string
		s                  string
		runeIdx            int
		expectedByteOffset int
		shouldPanic        bool
	}{
		{
			name:               "empty string",
			s:                  "",
			runeIdx:            0,
			expectedByteOffset: 0,
		},
		{
			name:        "negative runeIdx",
			s:           "hello",
			runeIdx:     -1,
			shouldPanic: true,
		},
		{
			name:               "single char",
			s:                  "a",
			runeIdx:            0,
			expectedByteOffset: 0,
		},
		{
			name:        "runeIdx out of bounds",
			s:           "a",
			runeIdx:     1,
			shouldPanic: true,
		},
		{
			name:               "longer",
			s:                  "hello",
			runeIdx:            3,
			expectedByteOffset: 3,
		},
		{
			name:               "ansi",
			s:                  "hi " + internal.RedBg.Render("there") + " leo",
			runeIdx:            8,
			expectedByteOffset: 8,
		},
		{
			name: "unicode",
			s:    "A💖中é",
			// A (1w, 1b, 1r), 💖 (2w, 4b, 1r), 中 (2w, 3b, 1r), é (1w, 3b, 2r) = 6w, 11b, 5r
			runeIdx:            3, // first rune in é
			expectedByteOffset: 8,
		},
		{
			name: "unicode zero-width",
			s:    "A💖中é",
			// A (1w, 1b, 1r), 💖 (2w, 4b, 1r), 中 (2w, 3b, 1r), é (1w, 3b, 2r) = 6w, 11b, 5r
			runeIdx:            4, // second rune in é
			expectedByteOffset: 9,
		},
		{
			name:               "unicode zero-width single char",
			s:                  "é",
			runeIdx:            1,
			expectedByteOffset: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.s)

			if tt.shouldPanic {
				assertPanic(t, func() {
					item.getByteOffsetAtRuneIdx(tt.runeIdx)
				})
				return
			}

			actual := item.getByteOffsetAtRuneIdx(tt.runeIdx)
			if int(actual) != tt.expectedByteOffset {
				t.Errorf("getByteOffsetAtRuneIdx() got %d, expected %d", actual, tt.expectedByteOffset)
			}
		})
	}
}
