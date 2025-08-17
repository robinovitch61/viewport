package linebuffer

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

func TestLineBuffer_reapplyAnsi(t *testing.T) {
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
			original:        redBg.Render("A💖") + "中é",
			truncated:       "A💖中é",
			truncByteOffset: 0,
			expected:        redBg.Render("A💖") + "中é",
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

func TestLineBuffer_highlightLine(t *testing.T) {
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
			highlightStyle: redFg,
			expected:       "",
		},
		{
			name:           "no highlight",
			line:           "hello",
			highlight:      "",
			highlightStyle: redFg,
			expected:       "hello",
		},
		{
			name:           "highlight",
			line:           "hello",
			highlight:      "ell",
			highlightStyle: redFg,
			expected:       "h" + redFg.Render("ell") + "o",
		},
		{
			name:           "highlight already styled line",
			line:           redFg.Render("first line"),
			highlight:      "first",
			highlightStyle: blueBg,
			expected:       blueBg.Render("first") + redFg.Render(" line"),
		},
		{
			name:           "highlight already partially styled line",
			line:           "hi a " + redFg.Render("styled line") + " cool " + redFg.Render("and styled") + " more",
			highlight:      "style",
			highlightStyle: blueBg,
			expected:       "hi a " + blueBg.Render("style") + redFg.Render("d line") + " cool " + redFg.Render("and ") + blueBg.Render("style") + redFg.Render("d") + " more",
		},
		{
			name:           "dont highlight ansi escape codes themselves",
			line:           redFg.Render("hi"),
			highlight:      "38",
			highlightStyle: blueBg,
			expected:       redFg.Render("hi"),
		},
		{
			name:           "single letter in partially styled line",
			line:           "line " + redFg.Render("red") + " e again",
			highlight:      "e",
			highlightStyle: blueBg,
			expected:       "lin" + blueBg.Render("e") + " " + redFg.Render("r") + blueBg.Render("e") + redFg.Render("d") + " " + blueBg.Render("e") + " again",
		},
		{
			name:           "super long line",
			line:           strings.Repeat("python generator code world world world code text test code words random words generator hello python generator", 10000),
			highlight:      "e",
			highlightStyle: redFg,
			expected:       strings.Repeat("python g"+redFg.Render("e")+"n"+redFg.Render("e")+"rator cod"+redFg.Render("e")+" world world world cod"+redFg.Render("e")+" t"+redFg.Render("e")+"xt t"+redFg.Render("e")+"st cod"+redFg.Render("e")+" words random words g"+redFg.Render("e")+"n"+redFg.Render("e")+"rator h"+redFg.Render("e")+"llo python g"+redFg.Render("e")+"n"+redFg.Render("e")+"rator", 10000),
		},
		{
			name:           "start and end",
			line:           "my line",
			highlight:      "line",
			highlightStyle: redFg,
			start:          0,
			end:            2,
			expected:       "my line",
		},
		{
			name:           "start and end ansi, in range",
			line:           blueBg.Render("my line"),
			highlight:      "my",
			highlightStyle: redFg,
			start:          0,
			end:            2,
			expected:       redFg.Render("my") + blueBg.Render(" line"),
		},
		{
			name:           "start and end ansi, out of range",
			line:           blueBg.Render("my line"),
			highlight:      "my",
			highlightStyle: redFg,
			start:          2,
			end:            4,
			expected:       blueBg.Render("my line"),
		},
		{
			name:           "ansi across multiple styles",
			line:           redBg.Render("hello") + " " + blueBg.Render("world"),
			highlight:      "lo wo",
			highlightStyle: greenBg,
			start:          0,
			end:            11,
			expected:       redBg.Render("hel") + greenBg.Render("lo wo") + blueBg.Render("rld"),
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
		plainLine      string // full line without ANSI
		segmentStart   int
		segmentEnd     int
		expected       string
	}{
		{
			name:           "empty",
			styledSegment:  "",
			toHighlight:    "",
			highlightStyle: redFg,
			plainLine:      "",
			segmentStart:   0,
			segmentEnd:     0,
			expected:       "",
		},
		{
			name:           "no highlight",
			styledSegment:  "hello",
			toHighlight:    "",
			highlightStyle: redFg,
			plainLine:      "hello",
			segmentStart:   0,
			segmentEnd:     5,
			expected:       "hello",
		},
		{
			name:           "simple highlight",
			styledSegment:  "hello",
			toHighlight:    "ell",
			highlightStyle: redFg,
			plainLine:      "hello",
			segmentStart:   0,
			segmentEnd:     5,
			expected:       "h\x1b[38;2;255;0;0mell" + RST + "o",
		},
		{
			name:           "highlight with existing style",
			styledSegment:  "\x1b[38;2;255;0;0mfirst line" + RST,
			toHighlight:    "first",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			plainLine:      "first line",
			segmentStart:   0,
			segmentEnd:     10,
			expected:       "\x1b[38;2;0;0;255mfirst" + RST + "\x1b[38;2;255;0;0m line" + RST,
		},
		{
			name:           "left overflow",
			styledSegment:  "ello world",
			toHighlight:    "hello",
			highlightStyle: redFg,
			plainLine:      "hello world",
			segmentStart:   1,
			segmentEnd:     11,
			expected:       "\x1b[38;2;255;0;0mello" + RST + " world",
		},
		{
			name:           "right overflow",
			styledSegment:  "hello wo",
			toHighlight:    "world",
			highlightStyle: redFg,
			plainLine:      "hello world",
			segmentStart:   0,
			segmentEnd:     8,
			expected:       "hello \x1b[38;2;255;0;0mwo" + RST,
		},
		{
			name:           "both overflow with existing style",
			styledSegment:  "\x1b[38;2;255;0;0mello wor" + RST,
			toHighlight:    "hello world",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			plainLine:      "hello world",
			segmentStart:   1,
			segmentEnd:     9,
			expected:       "\x1b[38;2;255;0;0mello wor" + RST,
		},
		{
			name:           "no match in segment",
			styledSegment:  "middle",
			toHighlight:    "outside",
			highlightStyle: redFg,
			plainLine:      "outside middle outside",
			segmentStart:   8,
			segmentEnd:     14,
			expected:       "middle",
		},
		{
			name:           "across ansi styles",
			styledSegment:  redBg.Render("hello") + " " + blueBg.Render("world"),
			toHighlight:    "lo wo",
			highlightStyle: greenBg,
			plainLine:      "hello world",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       redBg.Render("hel") + greenBg.Render("lo wo") + blueBg.Render("rld"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			toHighlight := HighlightData{
				StringToHighlight: tt.toHighlight,
				IsRegex:           false,
			}
			result := highlightString(
				tt.styledSegment,
				toHighlight,
				tt.highlightStyle,
				tt.plainLine,
				tt.segmentStart,
				tt.segmentEnd,
			)
			internal.CmpStr(t, tt.expected, result)
		})
	}
}

func TestHighlightStringRegex(t *testing.T) {
	for _, tt := range []struct {
		name           string
		styledSegment  string // segment with ANSI codes
		regexPattern   string
		highlightStyle lipgloss.Style
		plainLine      string // full line without ANSI
		segmentStart   int
		segmentEnd     int
		expected       string
	}{
		{
			name:           "simple regex match",
			styledSegment:  "hello world",
			regexPattern:   "wo.ld",
			highlightStyle: redFg,
			plainLine:      "hello world",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       "hello \x1b[38;2;255;0;0mworld" + RST,
		},
		{
			name:           "multiple regex matches",
			styledSegment:  "foo bar foo",
			regexPattern:   "foo",
			highlightStyle: redFg,
			plainLine:      "foo bar foo",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       "\x1b[38;2;255;0;0mfoo" + RST + " bar \x1b[38;2;255;0;0mfoo" + RST,
		},
		{
			name:           "no regex matches",
			styledSegment:  "hello world",
			regexPattern:   "xyz",
			highlightStyle: redFg,
			plainLine:      "hello world",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       "hello world",
		},
		{
			name:           "regex with existing ansi style",
			styledSegment:  "\x1b[38;2;255;0;0mhello world" + RST,
			regexPattern:   "world",
			highlightStyle: lipgloss.NewStyle().Foreground(blue),
			plainLine:      "hello world",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       "\x1b[38;2;255;0;0mhello " + RST + "\x1b[38;2;0;0;255mworld" + RST,
		},
		{
			name:           "regex across ansi styles",
			styledSegment:  redBg.Render("hello") + " " + blueBg.Render("world"),
			regexPattern:   "lo wo",
			highlightStyle: greenBg,
			plainLine:      "hello world",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       redBg.Render("hel") + greenBg.Render("lo wo") + blueBg.Render("rld"),
		},
		{
			name:           "case sensitive regex",
			styledSegment:  "Hello World",
			regexPattern:   "hello",
			highlightStyle: redFg,
			plainLine:      "Hello World",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       "Hello World",
		},
		{
			name:           "word boundary regex",
			styledSegment:  "hello world hello",
			regexPattern:   "\\bhello\\b",
			highlightStyle: redFg,
			plainLine:      "hello world hello",
			segmentStart:   0,
			segmentEnd:     17,
			expected:       "\x1b[38;2;255;0;0mhello" + RST + " world \x1b[38;2;255;0;0mhello" + RST,
		},
		{
			name:           "character class regex",
			styledSegment:  "abc 123 def",
			regexPattern:   "[0-9]+",
			highlightStyle: redFg,
			plainLine:      "abc 123 def",
			segmentStart:   0,
			segmentEnd:     11,
			expected:       "abc \x1b[38;2;255;0;0m123" + RST + " def",
		},
		{
			name:           "partial segment match",
			styledSegment:  "middle",
			regexPattern:   "mid",
			highlightStyle: redFg,
			plainLine:      "prefix middle suffix",
			segmentStart:   7,
			segmentEnd:     13,
			expected:       "\x1b[38;2;255;0;0mmid" + RST + "dle",
		},
		{
			name:           "empty segment",
			styledSegment:  "",
			regexPattern:   "test",
			highlightStyle: redFg,
			plainLine:      "",
			segmentStart:   0,
			segmentEnd:     0,
			expected:       "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			regex := regexp.MustCompile(tt.regexPattern)
			toHighlight := HighlightData{
				RegexPatternToHighlight: regex,
				IsRegex:                 true,
			}
			result := highlightString(
				tt.styledSegment,
				toHighlight,
				tt.highlightStyle,
				tt.plainLine,
				tt.segmentStart,
				tt.segmentEnd,
			)
			internal.CmpStr(t, tt.expected, result)
		})
	}
}

func TestLineBuffer_overflowsLeft(t *testing.T) {
	tests := []struct {
		name         string
		str          string
		startByteIdx int
		substr       string
		wantBool     bool
		wantInt      int
	}{
		{
			name:         "basic overflow case",
			str:          "my str here",
			startByteIdx: 3,
			substr:       "my str",
			wantBool:     true,
			wantInt:      6,
		},
		{
			name:         "no overflow case",
			str:          "my str here",
			startByteIdx: 6,
			substr:       "my str",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "empty string",
			str:          "",
			startByteIdx: 0,
			substr:       "test",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "empty substring",
			str:          "test string",
			startByteIdx: 0,
			substr:       "",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "startByteIdx out of bounds",
			str:          "test",
			startByteIdx: 10,
			substr:       "test",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "exact full match",
			str:          "hello world",
			startByteIdx: 0,
			substr:       "hello world",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "partial overflow at end",
			str:          "hello world",
			startByteIdx: 9,
			substr:       "dd",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "case sensitivity test - no match",
			str:          "Hello World",
			startByteIdx: 0,
			substr:       "hello",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "multiple character same overflow",
			str:          "aaaa",
			startByteIdx: 1,
			substr:       "aaa",
			wantBool:     true,
			wantInt:      3,
		},
		{
			name:         "multiple character same overflow but difference",
			str:          "aaaa",
			startByteIdx: 1,
			substr:       "baaa",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "special characters",
			str:          "test!@#$",
			startByteIdx: 4,
			substr:       "st!@#",
			wantBool:     true,
			wantInt:      7,
		},
		{
			name:         "false if does not overflow",
			str:          "some string",
			startByteIdx: 1,
			substr:       "ome",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "one char overflow",
			str:          "some string",
			startByteIdx: 1,
			substr:       "some",
			wantBool:     true,
			wantInt:      4,
		},
		// 世 is 3 bytes
		// 界 is 3 bytes
		// 🌟 is 4 bytes
		// "世界🌟世界🌟"[3:13] = "界🌟世"
		{
			name:         "unicode with ansi left not overflowing",
			str:          "世界🌟世界🌟",
			startByteIdx: 0,
			substr:       "世界🌟世",
			wantBool:     false,
			wantInt:      0,
		},
		{
			name:         "unicode with ansi left overflow 1 byte",
			str:          "世界🌟世界🌟",
			startByteIdx: 1,
			substr:       "世界🌟世",
			wantBool:     true,
			wantInt:      13,
		},
		{
			name:         "unicode with ansi left overflow 2 bytes",
			str:          "世界🌟世界🌟",
			startByteIdx: 2,
			substr:       "世界🌟世",
			wantBool:     true,
			wantInt:      13,
		},
		{
			name:         "unicode with ansi left overflow full rune",
			str:          "世界🌟世界🌟",
			startByteIdx: 3,
			substr:       "世界🌟世",
			wantBool:     true,
			wantInt:      13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotInt := overflowsLeft(tt.str, tt.startByteIdx, tt.substr)
			if gotBool != tt.wantBool || gotInt != tt.wantInt {
				t.Errorf("overflowsLeft(%q, %d, %q) = (%v, %d), want (%v, %d)",
					tt.str, tt.startByteIdx, tt.substr, gotBool, gotInt, tt.wantBool, tt.wantInt)
			}
		})
	}
}

func TestLineBuffer_overflowsRight(t *testing.T) {
	tests := []struct {
		name       string
		s          string
		endByteIdx int
		substr     string
		wantBool   bool
		wantInt    int
	}{
		{
			name:       "example 1",
			s:          "my str here",
			endByteIdx: 3,
			substr:     "y str",
			wantBool:   true,
			wantInt:    1,
		},
		{
			name:       "example 2",
			s:          "my str here",
			endByteIdx: 3,
			substr:     "y strong",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "example 3",
			s:          "my str here",
			endByteIdx: 6,
			substr:     "tr here",
			wantBool:   true,
			wantInt:    4,
		},
		{
			name:       "empty string",
			s:          "",
			endByteIdx: 0,
			substr:     "test",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "empty substring",
			s:          "test string",
			endByteIdx: 0,
			substr:     "",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "end index out of bounds",
			s:          "test",
			endByteIdx: 10,
			substr:     "test",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "exact full match",
			s:          "hello world",
			endByteIdx: 11,
			substr:     "hello world",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "case sensitivity test - no match",
			s:          "Hello World",
			endByteIdx: 4,
			substr:     "hello",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "multiple character same overflow",
			s:          "aaaa",
			endByteIdx: 2,
			substr:     "aaa",
			wantBool:   true,
			wantInt:    0,
		},
		{
			name:       "multiple character same overflow but difference",
			s:          "aaaa",
			endByteIdx: 2,
			substr:     "aaab",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "false if does not overflow",
			s:          "some string",
			endByteIdx: 5,
			substr:     "ome ",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "one char overflow",
			s:          "some string",
			endByteIdx: 5,
			substr:     "ome s",
			wantBool:   true,
			wantInt:    1,
		},
		// 世 is 3 bytes
		// 界 is 3 bytes
		// 🌟 is 4 bytes
		// "世界🌟世界🌟"[3:10] = "界🌟"
		{
			name:       "unicode with ansi no overflow",
			s:          "世界🌟世界🌟",
			endByteIdx: 13,
			substr:     "界🌟世",
			wantBool:   false,
			wantInt:    0,
		},
		{
			name:       "unicode with ansi overflow right one byte",
			s:          "世界🌟世界🌟",
			endByteIdx: 12,
			substr:     "界🌟世",
			wantBool:   true,
			wantInt:    3,
		},
		{
			name:       "unicode with ansi overflow right two bytes",
			s:          "世界🌟世界🌟",
			endByteIdx: 11,
			substr:     "界🌟世",
			wantBool:   true,
			wantInt:    3,
		},
		{
			name:       "unicode with ansi overflow right full rune",
			s:          "世界🌟世界🌟",
			endByteIdx: 10,
			substr:     "界🌟世",
			wantBool:   true,
			wantInt:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBool, gotInt := overflowsRight(tt.s, tt.endByteIdx, tt.substr)
			if gotBool != tt.wantBool || gotInt != tt.wantInt {
				t.Errorf("overflowsRight(%q, %d, %q) = (%v, %d), want (%v, %d)",
					tt.s, tt.endByteIdx, tt.substr, gotBool, gotInt, tt.wantBool, tt.wantInt)
			}
		})
	}
}

func TestLineBuffer_replaceStartWithContinuation(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		continuation string
		expected     string
	}{
		{
			name:         "empty",
			s:            "",
			continuation: "",
			expected:     "",
		},
		{
			name:         "empty continuation",
			s:            "my string",
			continuation: "",
			expected:     "my string",
		},
		{
			name:         "simple",
			s:            "my string",
			continuation: "...",
			expected:     "...string",
		},
		{
			name:         "ansi from start",
			s:            "\x1b[31mmy string" + RST,
			continuation: "...",
			expected:     "\x1b[31m...string" + RST,
		},
		{
			name:         "ansi overlaps continuation",
			s:            "m\x1b[31my string" + RST,
			continuation: "...",
			expected:     ".\x1b[31m..string" + RST,
		},
		{
			name: "unicode",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖中é",
			continuation: "...",
			expected:     "...中é",
		},
		{
			name: "unicode leading combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "é💖中",
			continuation: "...",
			expected:     "...中",
		},
		{
			name: "unicode combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "💖é💖中",
			continuation: "...",
			expected:     "...💖中",
		},
		{
			name: "unicode width overlap",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "中💖中é",
			continuation: "...",
			expected:     "..💖中é", // continuation shrinks by 1
		},
		{
			name: "unicode start",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖中é",
			continuation: "...",
			expected:     "...中é",
		},
		{
			name: "unicode start ansi",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            redBg.Render("A💖") + "中é",
			continuation: "...",
			expected:     redBg.Render("...") + "中é",
		},
		{
			name: "unicode almost start ansi",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A" + redBg.Render("💖") + "中é",
			continuation: "...",
			expected:     "." + redBg.Render("..") + "中é",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if r := replaceStartWithContinuation(tt.s, []rune(tt.continuation)); r != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, r)
			}
		})
	}
}

func TestLineBuffer_replaceEndWithContinuation(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		continuation string
		expected     string
	}{
		{
			name:         "empty",
			s:            "",
			continuation: "",
			expected:     "",
		},
		{
			name:         "empty continuation",
			s:            "my string",
			continuation: "",
			expected:     "my string",
		},
		{
			name:         "simple",
			s:            "my string",
			continuation: "...",
			expected:     "my str...",
		},
		{
			name:         "ansi from end",
			s:            "\x1b[31mmy string" + RST,
			continuation: "...",
			expected:     "\x1b[31mmy str..." + RST,
		},
		{
			name:         "ansi overlaps continuation",
			s:            "\x1b[31mmy strin" + RST + "g",
			continuation: "...",
			expected:     "\x1b[31mmy str.." + RST + ".",
		},
		{
			name: "unicode",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖中é",
			continuation: "...",
			expected:     "A💖...",
		},
		{
			name: "unicode trailing combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖中é",
			continuation: "...",
			expected:     "A💖...",
		},
		{
			name: "unicode combined",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖é中",
			continuation: "...",
			expected:     "A💖...",
		},
		{
			name: "unicode width overlap",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "💖中",
			continuation: "...",
			expected:     "💖..", // continuation shrinks by 1
		},
		{
			name: "unicode end",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖中é",
			continuation: "...",
			expected:     "A💖...",
		},
		{
			name: "unicode end ansi",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A💖" + redBg.Render("中é"),
			continuation: "...",
			expected:     "A💖" + redBg.Render("..."),
		},
		{
			name: "unicode almost end ansi",
			// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
			s:            "A" + redBg.Render("💖中") + "é",
			continuation: "...",
			expected:     "A" + redBg.Render("💖..") + ".",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if r := replaceEndWithContinuation(tt.s, []rune(tt.continuation)); r != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, r)
			}
		})
	}
}

func TestLineBuffer_getNonAnsiBytes(t *testing.T) {
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
			s:            "abc" + redBg.Render("def") + "ghi",
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
			s:            "A💖" + redBg.Render("中") + "é",
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

func TestLineBuffer_getBytesLeftOfWidth(t *testing.T) {
	tests := []struct {
		name           string
		buffers        []LineBuffer
		nBytes         int
		startBufferIdx int
		widthToLeft    int
		expected       string
		shouldPanic    bool
	}{
		{
			name:           "empty buffers",
			buffers:        nil,
			nBytes:         1,
			startBufferIdx: 0,
			widthToLeft:    0,
			expected:       "",
		},
		{
			name:           "negative bytes",
			buffers:        []LineBuffer{New("abc")},
			nBytes:         -1,
			startBufferIdx: 0,
			widthToLeft:    1,
			shouldPanic:    true,
		},
		{
			name:           "zero bytes",
			buffers:        []LineBuffer{New("abc")},
			nBytes:         0,
			startBufferIdx: 0,
			widthToLeft:    1,
			expected:       "",
		},
		{
			name:           "buffer index out of bounds",
			buffers:        []LineBuffer{New("abc")},
			nBytes:         1,
			startBufferIdx: 1,
			widthToLeft:    0,
			expected:       "",
		},
		{
			name:           "single buffer full content",
			buffers:        []LineBuffer{New("abc")},
			nBytes:         3,
			startBufferIdx: 0,
			widthToLeft:    3,
			expected:       "abc",
		},
		{
			name:           "single buffer partial content",
			buffers:        []LineBuffer{New("abc")},
			nBytes:         2,
			startBufferIdx: 0,
			widthToLeft:    2,
			expected:       "ab",
		},
		{
			name: "multiple buffers full content",
			buffers: []LineBuffer{
				New("abc"),
				New("def"),
			},
			nBytes:         6,
			startBufferIdx: 1,
			widthToLeft:    3,
			expected:       "abcdef",
		},
		{
			name: "multiple buffers partial content",
			buffers: []LineBuffer{
				New("abc"),
				New("def"),
			},
			nBytes:         4,
			startBufferIdx: 1,
			widthToLeft:    2,
			expected:       "bcde",
		},
		{
			name: "ignore ansi codes",
			buffers: []LineBuffer{
				New("a" + redBg.Render("b") + "c"),
				New(redBg.Render("def")),
			},
			nBytes:         5,
			startBufferIdx: 1,
			widthToLeft:    3,
			expected:       "bcdef",
		},
		// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
		{
			name: "unicode characters",
			buffers: []LineBuffer{
				New("A💖中"),
				New("é"),
			},
			nBytes:         10,
			startBufferIdx: 1,
			widthToLeft:    1,
			expected:       "💖中é",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assertPanic(t, func() {
					getBytesLeftOfWidth(tt.nBytes, tt.buffers, tt.startBufferIdx, tt.widthToLeft)
				})
				return
			}

			if got := getBytesLeftOfWidth(tt.nBytes, tt.buffers, tt.startBufferIdx, tt.widthToLeft); got != tt.expected {
				t.Errorf("getBytesLeftOfWidth() = %v, want %v", []byte(got), []byte(tt.expected))
			}
		})
	}
}

func TestLineBuffer_getBytesRightOfWidth(t *testing.T) {
	tests := []struct {
		name         string
		buffers      []LineBuffer
		nBytes       int
		endBufferIdx int
		widthToRight int
		expected     string
		shouldPanic  bool
	}{
		{
			name:         "empty buffers",
			buffers:      nil,
			nBytes:       1,
			endBufferIdx: 0,
			widthToRight: 0,
			expected:     "",
		},
		{
			name:         "negative bytes",
			buffers:      []LineBuffer{New("abc")},
			nBytes:       -1,
			endBufferIdx: 0,
			widthToRight: 1,
			shouldPanic:  true,
		},
		{
			name:         "zero bytes",
			buffers:      []LineBuffer{New("abc")},
			nBytes:       0,
			endBufferIdx: 0,
			widthToRight: 1,
			expected:     "",
		},
		{
			name:         "buffer index out of bounds",
			buffers:      []LineBuffer{New("abc")},
			nBytes:       1,
			endBufferIdx: 1,
			widthToRight: 0,
			expected:     "",
		},
		{
			name:         "single buffer full content",
			buffers:      []LineBuffer{New("abc")},
			nBytes:       3,
			endBufferIdx: 0,
			widthToRight: 3,
			expected:     "abc",
		},
		{
			name:         "single buffer partial content",
			buffers:      []LineBuffer{New("abc")},
			nBytes:       2,
			endBufferIdx: 0,
			widthToRight: 2,
			expected:     "bc",
		},
		{
			name: "multiple buffers full content",
			buffers: []LineBuffer{
				New("abc"),
				New("def"),
			},
			nBytes:       6,
			endBufferIdx: 0,
			widthToRight: 3,
			expected:     "abcdef",
		},
		{
			name: "multiple buffers partial content",
			buffers: []LineBuffer{
				New("abc"),
				New("def"),
			},
			nBytes:       4,
			endBufferIdx: 0,
			widthToRight: 2,
			expected:     "bcde",
		},
		{
			name: "ignore ansi codes",
			buffers: []LineBuffer{
				New("a" + redBg.Render("b") + "c"),
				New(redBg.Render("def")),
			},
			nBytes:       5,
			endBufferIdx: 0,
			widthToRight: 2,
			expected:     "bcdef",
		},
		// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
		{
			name: "unicode characters",
			buffers: []LineBuffer{
				New("A💖中"),
				New("é"),
			},
			nBytes:       10,
			endBufferIdx: 0,
			widthToRight: 4,
			expected:     "💖中é",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assertPanic(t, func() {
					getBytesRightOfWidth(tt.nBytes, tt.buffers, tt.endBufferIdx, tt.widthToRight)
				})
				return
			}

			if got := getBytesRightOfWidth(tt.nBytes, tt.buffers, tt.endBufferIdx, tt.widthToRight); got != tt.expected {
				t.Errorf("getBytesRightOfWidth() = %v, want %v", []byte(got), []byte(tt.expected))
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
