package item

import (
	"regexp"
	"strings"
	"testing"

	"github.com/robinovitch61/bubbleo/internal"

	"github.com/muesli/termenv"

	"github.com/charmbracelet/lipgloss"
)

// Note: this won't be necessary in future charm library versions
func init() {
	// Force TrueColor profile for consistent test output
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestAnsi_reapplyAnsi(t *testing.T) {
	tests := []struct {
		name            string
		original        string
		truncated       string
		truncByteOffset int
		expected        string
	}{
		{
			name:            "no ansi, no offset",
			original:        "1234567890123456789012345",
			truncated:       "12345",
			truncByteOffset: 0,
			expected:        "12345",
		},
		{
			name:            "no ansi, offset",
			original:        "1234567890123456789012345",
			truncated:       "2345",
			truncByteOffset: 1,
			expected:        "2345",
		},
		{
			name:            "multi ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1" + RST + "\x1b[38;2;0;0;255m2" + RST + "\x1b[38;2;255;0;0m3" + RST + "45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1" + RST + "\x1b[38;2;0;0;255m2" + RST + "\x1b[38;2;255;0;0m3" + RST,
		},
		{
			name:            "surrounding ansi, no offset",
			original:        "\x1b[38;2;255;0;0m12345" + RST,
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m123" + RST,
		},
		{
			name:            "surrounding ansi, offset",
			original:        "\x1b[38;2;255;0;0m12345" + RST,
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m234" + RST,
		},
		{
			name:            "left ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1" + RST + "2345",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1" + RST + "23",
		},
		{
			name:            "left ansi, offset",
			original:        "\x1b[38;2;255;0;0m12" + RST + "345",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m2" + RST + "34",
		},
		{
			name:            "right ansi, no offset",
			original:        "1" + "\x1b[38;2;255;0;0m2345" + RST,
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "1" + "\x1b[38;2;255;0;0m23" + RST,
		},
		{
			name:            "right ansi, offset",
			original:        "12" + "\x1b[38;2;255;0;0m345" + RST,
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "2" + "\x1b[38;2;255;0;0m34" + RST,
		},
		{
			name:            "left and right ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1" + RST + "2" + "\x1b[38;2;255;0;0m345" + RST,
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1" + RST + "2" + "\x1b[38;2;255;0;0m3" + RST,
		},
		{
			name:            "left and right ansi, offset",
			original:        "\x1b[38;2;255;0;0m12" + RST + "3" + "\x1b[38;2;255;0;0m45" + RST,
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m2" + RST + "3" + "\x1b[38;2;255;0;0m4" + RST,
		},
		{
			name:            "truncated right ansi, no offset",
			original:        "\x1b[38;2;255;0;0m1" + RST + "234" + "\x1b[38;2;255;0;0m5" + RST,
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1" + RST + "23",
		},
		{
			name:            "truncated right ansi, offset",
			original:        "\x1b[38;2;255;0;0m12" + RST + "34" + "\x1b[38;2;255;0;0m5" + RST,
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[38;2;255;0;0m2" + RST + "34",
		},
		{
			name:            "truncated left ansi, offset",
			original:        "\x1b[38;2;255;0;0m1" + RST + "23" + "\x1b[38;2;255;0;0m45" + RST,
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "23" + "\x1b[38;2;255;0;0m4" + RST,
		},
		{
			name:            "nested color sequences",
			original:        "\x1b[31m1\x1b[32m2\x1b[33m3" + RST + RST + RST + "45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[31m1\x1b[32m2\x1b[33m3" + RST,
		},
		{
			name:            "nested color sequences with offset",
			original:        "\x1b[31m1\x1b[32m2\x1b[33m3" + RST + RST + RST + "45",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[32m2\x1b[33m3" + RST + "4",
		},
		{
			name:            "nested style sequences",
			original:        "\x1b[1m1\x1b[4m2\x1b[3m3" + RST + RST + RST + "45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[1m1\x1b[4m2\x1b[3m3" + RST,
		},
		{
			name:            "mixed nested sequences",
			original:        "\x1b[31m1\x1b[1m2\x1b[4;32m3" + RST + RST + RST + "45",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[1m2\x1b[4;32m3" + RST + "4",
		},
		{
			name:            "deeply nested sequences",
			original:        "\x1b[31m1\x1b[1m2\x1b[4m3\x1b[32m4" + RST + RST + RST + RST + "5",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[31m1\x1b[1m2\x1b[4m3" + RST,
		},
		{
			name:            "partial nested sequences",
			original:        "1\x1b[31m2\x1b[1m3\x1b[4m4" + RST + RST + RST + "5",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m2\x1b[1m3\x1b[4m4" + RST,
		},
		{
			name:            "overlapping nested sequences",
			original:        "\x1b[31m1\x1b[1m2" + RST + "3\x1b[4m4" + RST + "5",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[1m2" + RST + "3\x1b[4m4" + RST,
		},
		{
			name:            "complex RGB nested sequences",
			original:        "\x1b[38;2;255;0;0m1\x1b[1m2\x1b[38;2;0;255;0m3" + RST + RST + "45",
			truncated:       "123",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;255;0;0m1\x1b[1m2\x1b[38;2;0;255;0m3" + RST,
		},
		{
			name:            "nested sequences with background colors",
			original:        "\x1b[31;44m1\x1b[1m2\x1b[32;45m3" + RST + RST + "45",
			truncated:       "234",
			truncByteOffset: 1,
			expected:        "\x1b[31;44m\x1b[1m2\x1b[32;45m3" + RST + "4",
		},
		{
			name:            "emoji basic",
			original:        "1️⃣2️⃣3️⃣4️⃣5️⃣",
			truncated:       "1️⃣2️⃣3️⃣",
			truncByteOffset: 0,
			expected:        "1️⃣2️⃣3️⃣",
		},
		{
			name:            "emoji with ansi",
			original:        "\x1b[31m1️⃣\x1b[32m2️⃣\x1b[33m3️⃣" + RST,
			truncated:       "1️⃣2️⃣",
			truncByteOffset: 0,
			expected:        "\x1b[31m1️⃣\x1b[32m2️⃣" + RST,
		},
		{
			name:            "chinese characters",
			original:        "你好世界星星",
			truncated:       "你好世",
			truncByteOffset: 0,
			expected:        "你好世",
		},
		{
			name:            "simple with ansi and offset",
			original:        "\x1b[31ma\x1b[32mb\x1b[33mc" + RST + "de",
			truncated:       "bcd",
			truncByteOffset: 1,
			expected:        "\x1b[31m\x1b[32mb\x1b[33mc" + RST + "d",
		},
		{
			name:            "chinese with ansi and offset",
			original:        "\x1b[31m你\x1b[32m好\x1b[33m世" + RST + "界星",
			truncated:       "好世界",
			truncByteOffset: 3, // 你 is 3 bytes
			expected:        "\x1b[31m\x1b[32m好\x1b[33m世" + RST + "界",
		},
		{
			name:            "lots of leading empty ansi",
			original:        "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST + "\x1b[38;2;255;0;0mr" + RST,
			truncated:       "r",
			truncByteOffset: 10,
			expected:        "\x1b[38;2;255;0;0mr" + RST,
		},
		{
			name:            "complex ansi, no offset",
			original:        "\x1b[38;2;0;0;255msome " + RST + "\x1b[38;2;255;0;0mred" + RST + "\x1b[38;2;0;0;255m t" + RST,
			truncated:       "some red t",
			truncByteOffset: 0,
			expected:        "\x1b[38;2;0;0;255msome " + RST + "\x1b[38;2;255;0;0mred" + RST + "\x1b[38;2;0;0;255m t" + RST,
		},
		{
			name:            "unicode with ansi",
			original:        internal.RedBg.Render("A💖") + "中é",
			truncated:       "A💖中é",
			truncByteOffset: 0,
			expected:        internal.RedBg.Render("A💖") + "中é",
		},
	}

	ansiRegex := regexp.MustCompile("\x1b\\[[0-9;]*m")

	toUInt32 := func(indexes [][]int) [][]uint32 {
		uint32Indexes := make([][]uint32, len(indexes))
		for i, idx := range indexes {
			uint32Indexes[i] = []uint32{clampIntToUint32(idx[0]), clampIntToUint32(idx[1])}
		}
		return uint32Indexes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ansiCodeIndexes := toUInt32(ansiRegex.FindAllStringIndex(tt.original, -1))
			actual := reapplyAnsi(tt.original, tt.truncated, tt.truncByteOffset, ansiCodeIndexes)
			internal.CmpStr(t, tt.expected, actual)
		})
	}
}

func TestAnsi_highlightLine(t *testing.T) {
	for _, tt := range []struct {
		name           string
		line           string
		highlight      string
		highlightStyle lipgloss.Style
		start          int
		end            int
		expected       string
	}{
		{
			name:           "empty",
			line:           "",
			highlight:      "",
			highlightStyle: internal.RedFg,
			expected:       "",
		},
		{
			name:           "no highlight",
			line:           "hello",
			highlight:      "",
			highlightStyle: internal.RedFg,
			expected:       "hello",
		},
		{
			name:           "highlight",
			line:           "hello",
			highlight:      "ell",
			highlightStyle: internal.RedFg,
			expected:       "h" + internal.RedFg.Render("ell") + "o",
		},
		{
			name:           "highlight already styled line",
			line:           internal.RedFg.Render("first line"),
			highlight:      "first",
			highlightStyle: internal.BlueBg,
			expected:       internal.BlueBg.Render("first") + internal.RedFg.Render(" line"),
		},
		{
			name:           "highlight already partially styled line",
			line:           "hi a " + internal.RedFg.Render("styled line") + " cool " + internal.RedFg.Render("and styled") + " more",
			highlight:      "style",
			highlightStyle: internal.BlueBg,
			expected:       "hi a " + internal.BlueBg.Render("style") + internal.RedFg.Render("d line") + " cool " + internal.RedFg.Render("and ") + internal.BlueBg.Render("style") + internal.RedFg.Render("d") + " more",
		},
		{
			name:           "dont highlight ansi escape codes themselves",
			line:           internal.RedFg.Render("hi"),
			highlight:      "38",
			highlightStyle: internal.BlueBg,
			expected:       internal.RedFg.Render("hi"),
		},
		{
			name:           "single letter in partially styled line",
			line:           "line " + internal.RedFg.Render("red") + " e again",
			highlight:      "e",
			highlightStyle: internal.BlueBg,
			expected:       "lin" + internal.BlueBg.Render("e") + " " + internal.RedFg.Render("r") + internal.BlueBg.Render("e") + internal.RedFg.Render("d") + " " + internal.BlueBg.Render("e") + " again",
		},
		{
			name:           "super long line",
			line:           strings.Repeat("python generator code world world world code text test code words random words generator hello python generator", 10000),
			highlight:      "e",
			highlightStyle: internal.RedFg,
			expected:       strings.Repeat("python g"+internal.RedFg.Render("e")+"n"+internal.RedFg.Render("e")+"rator cod"+internal.RedFg.Render("e")+" world world world cod"+internal.RedFg.Render("e")+" t"+internal.RedFg.Render("e")+"xt t"+internal.RedFg.Render("e")+"st cod"+internal.RedFg.Render("e")+" words random words g"+internal.RedFg.Render("e")+"n"+internal.RedFg.Render("e")+"rator h"+internal.RedFg.Render("e")+"llo python g"+internal.RedFg.Render("e")+"n"+internal.RedFg.Render("e")+"rator", 10000),
		},
		{
			name:           "start and end",
			line:           "my line",
			highlight:      "line",
			highlightStyle: internal.RedFg,
			start:          0,
			end:            2,
			expected:       "my line",
		},
		{
			name:           "start and end ansi, in range",
			line:           internal.BlueBg.Render("my line"),
			highlight:      "my",
			highlightStyle: internal.RedFg,
			start:          0,
			end:            2,
			expected:       internal.RedFg.Render("my") + internal.BlueBg.Render(" line"),
		},
		{
			name:           "start and end ansi, out of range",
			line:           internal.BlueBg.Render("my line"),
			highlight:      "my",
			highlightStyle: internal.RedFg,
			start:          2,
			end:            4,
			expected:       internal.BlueBg.Render("my line"),
		},
		{
			name:           "ansi across multiple styles",
			line:           internal.RedBg.Render("hello") + " " + internal.BlueBg.Render("world"),
			highlight:      "lo wo",
			highlightStyle: internal.GreenBg,
			start:          0,
			end:            11,
			expected:       internal.RedBg.Render("hel") + internal.GreenBg.Render("lo wo") + internal.BlueBg.Render("rld"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.start == 0 && tt.end == 0 {
				tt.end = len(tt.line)
			}
			internal.CmpStr(t, tt.expected, highlightLine(tt.line, tt.highlight, tt.highlightStyle, tt.start, tt.end))
		})
	}
}

func TestHighlightString(t *testing.T) {
	for _, tt := range []struct {
		name           string
		styledSegment  string // segment with ANSI codes
		toHighlight    string
		highlightStyle lipgloss.Style
		fullLine       string // full line without ANSI
		segmentStart   int
		segmentEnd     int
		expected       string
	}{
		{
			name:           "empty",
			styledSegment:  "",
			toHighlight:    "",
			highlightStyle: internal.RedFg,
			fullLine:       "",
			segmentStart:   0,
			segmentEnd:     0,
			expected:       "",
		},
		{
			name:           "no highlight",
			styledSegment:  "hello",
			toHighlight:    "",
			highlightStyle: internal.RedFg,
			fullLine:       "hello",
			segmentStart:   0,
			segmentEnd:     5,
			expected:       "hello",
		},
		{
			name:           "simple highlight",
			styledSegment:  "hello",
			toHighlight:    "ell",
			highlightStyle: internal.RedFg,
			fullLine:       "hello",
			segmentStart:   0,
			segmentEnd:     5,
			expected:       "h\x1b[38;2;255;0;0mell" + RST + "o",
		},
		{
			name:           "highlight with existing style",
			styledSegment:  "\x1b[38;2;255;0;0mfirst line" + RST,
			toHighlight:    "first",
			highlightStyle: internal.BlueFg,
			fullLine:       "first line",
			segmentStart:   0,
			segmentEnd:     10,
			expected:       "\x1b[38;2;0;0;255mfirst" + RST + "\x1b[38;2;255;0;0m line" + RST,
		},
		{
			name:           "left overflow",
			styledSegment:  "ello world",
			toHighlight:    "hello",
			highlightStyle: internal.RedFg,
			fullLine:       "hello world",
			segmentStart:   1,
			segmentEnd:     11,
			expected:       "\x1b[38;2;255;0;0mello" + RST + " world",
		},
		{
			name:           "right overflow",
			styledSegment:  "hello wo",
			toHighlight:    "world",
			highlightStyle: internal.RedFg,
			fullLine:       "hello world",
			segmentStart:   0,
			segmentEnd:     8,
			expected:       "hello \x1b[38;2;255;0;0mwo" + RST,
		},
		{
			name:           "both overflow with existing style",
			styledSegment:  "\x1b[38;2;255;0;0mello wor" + RST,
			toHighlight:    "hello world",
			highlightStyle: internal.BlueFg,
			fullLine:       "hello world",
			segmentStart:   1,
			segmentEnd:     9,
			expected:       "\x1b[38;2;0;0;255mello wor" + RST,
		},
		{
			name:           "no match in segment",
			styledSegment:  "middle",
			toHighlight:    "outside",
			highlightStyle: internal.RedFg,
			fullLine:       "outside middle outside",
			segmentStart:   8,
			segmentEnd:     14,
			expected:       "middle",
		},
		{
			name:           "across ansi styles",
			styledSegment:  internal.RedBg.Render("hello") + " " + internal.BlueBg.Render("world"),
			toHighlight:    "lo wo",
			highlightStyle: internal.GreenBg,
			fullLine:       "hello world",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       internal.RedBg.Render("hel") + internal.GreenBg.Render("lo wo") + internal.BlueBg.Render("rld"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var highlights []Highlight
			if tt.toHighlight != "" {
				highlights = ExtractHighlights([]string{tt.fullLine}, tt.toHighlight, tt.highlightStyle)
			}
			result := highlightString(
				tt.styledSegment,
				highlights,
				tt.segmentStart,
				tt.segmentEnd,
			)
			internal.CmpStr(t, tt.expected, result)
		})
	}
}

func TestAnsi_getNonAnsiBytes(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		startByteIdx int
		numBytes     int
		expected     string
		shouldPanic  bool
	}{
		{
			name:         "empty",
			s:            "",
			startByteIdx: 0,
			numBytes:     0,
			expected:     "",
		},
		{
			name:         "negative start panics",
			s:            "a",
			startByteIdx: -1,
			numBytes:     1,
			shouldPanic:  true,
		},
		{
			name:         "zero bytes",
			s:            "abc",
			startByteIdx: 1,
			numBytes:     0,
			expected:     "",
		},
		{
			name:         "negative bytes",
			s:            "abc",
			startByteIdx: 1,
			numBytes:     -1,
			expected:     "",
		},
		{
			name:         "all from start",
			s:            "abc",
			startByteIdx: 0,
			numBytes:     3,
			expected:     "abc",
		},
		{
			name:         "some from start",
			s:            "abc",
			startByteIdx: 0,
			numBytes:     2,
			expected:     "ab",
		},
		{
			name:         "rest from offset",
			s:            "abc",
			startByteIdx: 1,
			numBytes:     2,
			expected:     "bc",
		},
		{
			name:         "some from offset",
			s:            "abc",
			startByteIdx: 1,
			numBytes:     1,
			expected:     "b",
		},
		{
			name:         "ignore ansi",
			s:            "abc" + internal.RedBg.Render("def") + "ghi",
			startByteIdx: 1,
			numBytes:     7,
			expected:     "bcdefgh",
		},
		{
			name: "unicode",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖中é",
			startByteIdx: 1,
			numBytes:     7,
			expected:     "💖中",
		},
		{
			name: "unicode with ansi",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖" + internal.RedBg.Render("中") + "é",
			startByteIdx: 0,
			numBytes:     11,
			expected:     "A💖中é",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assertPanic(t, func() {
					getNonAnsiBytes(tt.s, tt.startByteIdx, tt.numBytes)
				})
				return
			}

			if r := getNonAnsiBytes(tt.s, tt.startByteIdx, tt.numBytes); r != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, r)
			}
		})
	}
}

// testing helper
func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("did not panic as expected")
		}
	}()
	f()
}
