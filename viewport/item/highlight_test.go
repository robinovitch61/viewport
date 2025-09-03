package item

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/internal"
)

func TestExtractHighlightsSubstring(t *testing.T) {
	tests := []struct {
		name           string
		items          []string
		exactMatch     string
		highlightStyle lipgloss.Style
		expected       []Highlight
	}{
		{
			name:           "empty items",
			items:          []string{},
			exactMatch:     "test",
			highlightStyle: internal.RedBg,
			expected:       []Highlight{},
		},
		{
			name:           "empty exact match",
			items:          []string{"hello", "world"},
			exactMatch:     "",
			highlightStyle: internal.RedBg,
			expected:       []Highlight{},
		},
		{
			name:           "no matches",
			items:          []string{"hell"},
			exactMatch:     "lo",
			highlightStyle: internal.RedBg,
			expected:       []Highlight{},
		},
		{
			name:           "no matches multiple items",
			items:          []string{"hello", "world"},
			exactMatch:     "",
			highlightStyle: internal.RedBg,
			expected:       []Highlight{},
		},
		{
			name:           "single match",
			items:          []string{"hello world"},
			exactMatch:     "world",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "multiple matches in single item",
			items:          []string{"hello world world"},
			exactMatch:     "world",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           internal.RedBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   17,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "matches across multiple items",
			items:          []string{"hello world", "test world", "world end"},
			exactMatch:     "world",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           internal.RedBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 5,
					EndByteOffset:   10,
					Style:           internal.RedBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "overlapping potential matches",
			items:          []string{"aaa"},
			exactMatch:     "aa",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   2,
					Style:           internal.RedBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 1,
					EndByteOffset:   3,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "case sensitive",
			items:          []string{"Hello HELLO hello"},
			exactMatch:     "hello",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   17,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "unicode characters",
			items:          []string{"世界 hello 🌟"},
			exactMatch:     "hello",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 7,
					EndByteOffset:   12,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "single character match",
			items:          []string{"abcabc"},
			exactMatch:     "a",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   1,
					Style:           internal.RedBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 3,
					EndByteOffset:   4,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "match at beginning and end",
			items:          []string{"test middle test"},
			exactMatch:     "test",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           internal.RedBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   16,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "simple ansi sequence stripped",
			items:          []string{"\x1b[31mhello\x1b[0m world"},
			exactMatch:     "hello",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "ansi sequence in middle of text",
			items:          []string{"start \x1b[32mmiddle\x1b[0m end"},
			exactMatch:     "middle",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   12,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "match spans across ansi codes",
			items:          []string{"wo\x1b[31mrl\x1b[0md"},
			exactMatch:     "world",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "multiple ansi sequences",
			items:          []string{"\x1b[31m\x1b[1mhello\x1b[0m\x1b[0m world \x1b[32mtest\x1b[0m"},
			exactMatch:     "hello",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.RedBg,
				},
			},
		},
		{
			name:           "multiple matches with ansi codes",
			items:          []string{"\x1b[31mtest\x1b[0m middle \x1b[32mtest\x1b[0m"},
			exactMatch:     "test",
			highlightStyle: internal.RedBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           internal.RedBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   16,
					Style:           internal.RedBg,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractHighlights(tt.items, tt.exactMatch, tt.highlightStyle)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d highlights, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				actual := result[i]
				if actual.ItemIndex != expected.ItemIndex ||
					actual.StartByteOffset != expected.StartByteOffset ||
					actual.EndByteOffset != expected.EndByteOffset {
					t.Errorf("highlight %d: expected ItemIndex=%d StartByteOffset=%d EndByteOffset=%d, got ItemIndex=%d StartByteOffset=%d EndByteOffset=%d",
						i, expected.ItemIndex, expected.StartByteOffset, expected.EndByteOffset,
						actual.ItemIndex, actual.StartByteOffset, actual.EndByteOffset)
				}
			}
		})
	}
}

func TestExtractHighlightsRegexMatch(t *testing.T) {
	tests := []struct {
		name           string
		items          []string
		regexPattern   string
		highlightStyle lipgloss.Style
		expected       []Highlight
		expectError    bool
	}{
		{
			name:           "empty items",
			items:          []string{},
			regexPattern:   "test",
			highlightStyle: internal.BlueBg,
			expected:       []Highlight{},
		},
		{
			name:           "invalid regex",
			items:          []string{"hello", "world"},
			regexPattern:   "[",
			highlightStyle: internal.BlueBg,
			expected:       nil,
			expectError:    true,
		},
		{
			name:           "no matches",
			items:          []string{"hello", "world"},
			regexPattern:   "xyz",
			highlightStyle: internal.BlueBg,
			expected:       []Highlight{},
		},
		{
			name:           "simple word match",
			items:          []string{"hello world"},
			regexPattern:   "world",
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "word boundary match",
			items:          []string{"hello world worldly"},
			regexPattern:   `\bworld\b`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "digit pattern",
			items:          []string{"line 123 has numbers 456"},
			regexPattern:   `\d+`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 5,
					EndByteOffset:   8,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 21,
					EndByteOffset:   24,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "case insensitive pattern",
			items:          []string{"Hello HELLO hello"},
			regexPattern:   `(?i)hello`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   17,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "multiple items with matches",
			items:          []string{"error: failed", "warning: issue", "error: timeout"},
			regexPattern:   `error:`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   6,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   6,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "capturing groups",
			items:          []string{"user: john", "user: jane"},
			regexPattern:   `user: (\w+)`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   10,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 0,
					EndByteOffset:   10,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "dot metacharacter",
			items:          []string{"a1b", "a.b", "axb"},
			regexPattern:   `a.b`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   3,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 0,
					EndByteOffset:   3,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   3,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "anchored pattern",
			items:          []string{"start middle", "middle end", "start end"},
			regexPattern:   `^start`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "unicode with regex",
			items:          []string{"世界 test 🌟", "test 世界"},
			regexPattern:   `test`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 7,
					EndByteOffset:   11,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "overlapping matches not possible with regex",
			items:          []string{"aaa"},
			regexPattern:   `aa`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   2,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "regex match with ansi codes stripped",
			items:          []string{"\x1b[31mhello\x1b[0m world \x1b[32m123\x1b[0m"},
			regexPattern:   `\d+`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   15,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "word boundary with ansi",
			items:          []string{"\x1b[31mworld\x1b[0m worldly"},
			regexPattern:   `\bworld\b`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "case insensitive with ansi",
			items:          []string{"\x1b[31mHELLO\x1b[0m \x1b[32mhello\x1b[0m"},
			regexPattern:   `(?i)hello`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "anchored pattern with ansi at start",
			items:          []string{"\x1b[31mstart\x1b[0m middle", "middle \x1b[31mstart\x1b[0m"},
			regexPattern:   `^start`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           internal.BlueBg,
				},
			},
		},
		{
			name:           "complex ansi sequences with multiple colors",
			items:          []string{"\x1b[31;44;1mtest\x1b[0m normal \x1b[32mtext\x1b[0m"},
			regexPattern:   `test|text`,
			highlightStyle: internal.BlueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           internal.BlueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   16,
					Style:           internal.BlueBg,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractHighlightsRegexMatch(tt.items, tt.regexPattern, tt.highlightStyle)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d highlights, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				actual := result[i]
				if actual.ItemIndex != expected.ItemIndex ||
					actual.StartByteOffset != expected.StartByteOffset ||
					actual.EndByteOffset != expected.EndByteOffset {
					t.Errorf("highlight %d: expected ItemIndex=%d StartByteOffset=%d EndByteOffset=%d, got ItemIndex=%d StartByteOffset=%d EndByteOffset=%d",
						i, expected.ItemIndex, expected.StartByteOffset, expected.EndByteOffset,
						actual.ItemIndex, actual.StartByteOffset, actual.EndByteOffset)
				}
			}
		})
	}
}
