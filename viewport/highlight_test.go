package viewport

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestExtractHighlightsSubstring(t *testing.T) {
	redBg := lipgloss.NewStyle().Background(lipgloss.Color("#FF0000"))

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
			highlightStyle: redBg,
			expected:       []Highlight{},
		},
		{
			name:           "empty exact match",
			items:          []string{"hello", "world"},
			exactMatch:     "",
			highlightStyle: redBg,
			expected:       []Highlight{},
		},
		{
			name:           "no matches",
			items:          []string{"hell"},
			exactMatch:     "lo",
			highlightStyle: redBg,
			expected:       []Highlight{},
		},
		{
			name:           "no matches multiple items",
			items:          []string{"hello", "world"},
			exactMatch:     "",
			highlightStyle: redBg,
			expected:       []Highlight{},
		},
		{
			name:           "single match",
			items:          []string{"hello world"},
			exactMatch:     "world",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           redBg,
				},
			},
		},
		{
			name:           "multiple matches in single item",
			items:          []string{"hello world world"},
			exactMatch:     "world",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           redBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   17,
					Style:           redBg,
				},
			},
		},
		{
			name:           "matches across multiple items",
			items:          []string{"hello world", "test world", "world end"},
			exactMatch:     "world",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           redBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 5,
					EndByteOffset:   10,
					Style:           redBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           redBg,
				},
			},
		},
		{
			name:           "overlapping potential matches",
			items:          []string{"aaa"},
			exactMatch:     "aa",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   2,
					Style:           redBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 1,
					EndByteOffset:   3,
					Style:           redBg,
				},
			},
		},
		{
			name:           "case sensitive",
			items:          []string{"Hello HELLO hello"},
			exactMatch:     "hello",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   17,
					Style:           redBg,
				},
			},
		},
		{
			name:           "unicode characters",
			items:          []string{"世界 hello 🌟"},
			exactMatch:     "hello",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 7,
					EndByteOffset:   12,
					Style:           redBg,
				},
			},
		},
		{
			name:           "single character match",
			items:          []string{"abcabc"},
			exactMatch:     "a",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   1,
					Style:           redBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 3,
					EndByteOffset:   4,
					Style:           redBg,
				},
			},
		},
		{
			name:           "match at beginning and end",
			items:          []string{"test middle test"},
			exactMatch:     "test",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           redBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   16,
					Style:           redBg,
				},
			},
		},
		{
			name:           "simple ansi sequence stripped",
			items:          []string{"\x1b[31mhello\x1b[0m world"},
			exactMatch:     "hello",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           redBg,
				},
			},
		},
		{
			name:           "ansi sequence in middle of text",
			items:          []string{"start \x1b[32mmiddle\x1b[0m end"},
			exactMatch:     "middle",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   12,
					Style:           redBg,
				},
			},
		},
		{
			name:           "match spans across ansi codes",
			items:          []string{"wo\x1b[31mrl\x1b[0md"},
			exactMatch:     "world",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           redBg,
				},
			},
		},
		{
			name:           "multiple ansi sequences",
			items:          []string{"\x1b[31m\x1b[1mhello\x1b[0m\x1b[0m world \x1b[32mtest\x1b[0m"},
			exactMatch:     "hello",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           redBg,
				},
			},
		},
		{
			name:           "multiple matches with ansi codes",
			items:          []string{"\x1b[31mtest\x1b[0m middle \x1b[32mtest\x1b[0m"},
			exactMatch:     "test",
			highlightStyle: redBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           redBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   16,
					Style:           redBg,
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
	blueBg := lipgloss.NewStyle().Background(lipgloss.Color("#0000FF"))

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
			highlightStyle: blueBg,
			expected:       []Highlight{},
		},
		{
			name:           "invalid regex",
			items:          []string{"hello", "world"},
			regexPattern:   "[",
			highlightStyle: blueBg,
			expected:       nil,
			expectError:    true,
		},
		{
			name:           "no matches",
			items:          []string{"hello", "world"},
			regexPattern:   "xyz",
			highlightStyle: blueBg,
			expected:       []Highlight{},
		},
		{
			name:           "simple word match",
			items:          []string{"hello world"},
			regexPattern:   "world",
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "word boundary match",
			items:          []string{"hello world worldly"},
			regexPattern:   `\bworld\b`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "digit pattern",
			items:          []string{"line 123 has numbers 456"},
			regexPattern:   `\d+`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 5,
					EndByteOffset:   8,
					Style:           blueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 21,
					EndByteOffset:   24,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "case insensitive pattern",
			items:          []string{"Hello HELLO hello"},
			regexPattern:   `(?i)hello`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           blueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           blueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   17,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "multiple items with matches",
			items:          []string{"error: failed", "warning: issue", "error: timeout"},
			regexPattern:   `error:`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   6,
					Style:           blueBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   6,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "capturing groups",
			items:          []string{"user: john", "user: jane"},
			regexPattern:   `user: (\w+)`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   10,
					Style:           blueBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 0,
					EndByteOffset:   10,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "dot metacharacter",
			items:          []string{"a1b", "a.b", "axb"},
			regexPattern:   `a.b`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   3,
					Style:           blueBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 0,
					EndByteOffset:   3,
					Style:           blueBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   3,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "anchored pattern",
			items:          []string{"start middle", "middle end", "start end"},
			regexPattern:   `^start`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           blueBg,
				},
				{
					ItemIndex:       2,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "unicode with regex",
			items:          []string{"世界 test 🌟", "test 世界"},
			regexPattern:   `test`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 7,
					EndByteOffset:   11,
					Style:           blueBg,
				},
				{
					ItemIndex:       1,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "overlapping matches not possible with regex",
			items:          []string{"aaa"},
			regexPattern:   `aa`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   2,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "regex match with ansi codes stripped",
			items:          []string{"\x1b[31mhello\x1b[0m world \x1b[32m123\x1b[0m"},
			regexPattern:   `\d+`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   15,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "word boundary with ansi",
			items:          []string{"\x1b[31mworld\x1b[0m worldly"},
			regexPattern:   `\bworld\b`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "case insensitive with ansi",
			items:          []string{"\x1b[31mHELLO\x1b[0m \x1b[32mhello\x1b[0m"},
			regexPattern:   `(?i)hello`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           blueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 6,
					EndByteOffset:   11,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "anchored pattern with ansi at start",
			items:          []string{"\x1b[31mstart\x1b[0m middle", "middle \x1b[31mstart\x1b[0m"},
			regexPattern:   `^start`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   5,
					Style:           blueBg,
				},
			},
		},
		{
			name:           "complex ansi sequences with multiple colors",
			items:          []string{"\x1b[31;44;1mtest\x1b[0m normal \x1b[32mtext\x1b[0m"},
			regexPattern:   `test|text`,
			highlightStyle: blueBg,
			expected: []Highlight{
				{
					ItemIndex:       0,
					StartByteOffset: 0,
					EndByteOffset:   4,
					Style:           blueBg,
				},
				{
					ItemIndex:       0,
					StartByteOffset: 12,
					EndByteOffset:   16,
					Style:           blueBg,
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
