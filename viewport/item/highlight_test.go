package item

import (
	"testing"
)

func TestExtractMatches(t *testing.T) {
	tests := []struct {
		name       string
		items      []string
		exactMatch string
		expected   []Match
	}{
		{
			name:       "empty items",
			items:      []string{},
			exactMatch: "test",
			expected:   []Match{},
		},
		{
			name:       "empty exact match",
			items:      []string{"hello", "world"},
			exactMatch: "",
			expected:   []Match{},
		},
		{
			name:       "no matches",
			items:      []string{"hell"},
			exactMatch: "lo",
			expected:   []Match{},
		},
		{
			name:       "no matches multiple items",
			items:      []string{"hello", "world"},
			exactMatch: "",
			expected:   []Match{},
		},
		{
			name:       "single match",
			items:      []string{"hello world"},
			exactMatch: "world",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 6,
					EndByteUnstyledContent:   11,
				},
			},
		},
		{
			name:       "multiple matches in single item",
			items:      []string{"hello world world"},
			exactMatch: "world",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 6,
					EndByteUnstyledContent:   11,
				},
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 12,
					EndByteUnstyledContent:   17,
				},
			},
		},
		{
			name:       "matches across multiple items",
			items:      []string{"hello world", "test world", "world end"},
			exactMatch: "world",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 6,
					EndByteUnstyledContent:   11,
				},
				{
					ItemIndex:                1,
					StartByteUnstyledContent: 5,
					EndByteUnstyledContent:   10,
				},
				{
					ItemIndex:                2,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   5,
				},
			},
		},
		{
			name:       "overlapping potential matches",
			items:      []string{"aaa"},
			exactMatch: "aa",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   2,
				},
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 1,
					EndByteUnstyledContent:   3,
				},
			},
		},
		{
			name:       "case sensitive",
			items:      []string{"Hello HELLO hello"},
			exactMatch: "hello",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 12,
					EndByteUnstyledContent:   17,
				},
			},
		},
		{
			name:       "unicode characters",
			items:      []string{"世界 hello 🌟"},
			exactMatch: "hello",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 7,
					EndByteUnstyledContent:   12,
				},
			},
		},
		{
			name:       "single character match",
			items:      []string{"abcabc"},
			exactMatch: "a",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   1,
				},
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 3,
					EndByteUnstyledContent:   4,
				},
			},
		},
		{
			name:       "match at beginning and end",
			items:      []string{"test middle test"},
			exactMatch: "test",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   4,
				},
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 12,
					EndByteUnstyledContent:   16,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMatches(tt.items, tt.exactMatch)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d highlights, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				actual := result[i]
				if actual.ItemIndex != expected.ItemIndex ||
					actual.StartByteUnstyledContent != expected.StartByteUnstyledContent ||
					actual.EndByteUnstyledContent != expected.EndByteUnstyledContent {
					t.Errorf("highlight %d: expected ItemIndex=%d StartByteUnstyledContent=%d EndByteUnstyledContent=%d, got ItemIndex=%d StartByteUnstyledContent=%d EndByteUnstyledContent=%d",
						i, expected.ItemIndex, expected.StartByteUnstyledContent, expected.EndByteUnstyledContent,
						actual.ItemIndex, actual.StartByteUnstyledContent, actual.EndByteUnstyledContent)
				}
			}
		})
	}
}

func TestExtractMatchesRegex(t *testing.T) {
	tests := []struct {
		name         string
		items        []string
		regexPattern string
		expected     []Match
		expectError  bool
	}{
		{
			name:         "empty items",
			items:        []string{},
			regexPattern: "test",
			expected:     []Match{},
		},
		{
			name:         "invalid regex",
			items:        []string{"hello", "world"},
			regexPattern: "[",
			expected:     nil,
			expectError:  true,
		},
		{
			name:         "no matches",
			items:        []string{"hello", "world"},
			regexPattern: "xyz",
			expected:     []Match{},
		},
		{
			name:         "simple word match",
			items:        []string{"hello world"},
			regexPattern: "world",
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 6,
					EndByteUnstyledContent:   11,
				},
			},
		},
		{
			name:         "word boundary match",
			items:        []string{"hello world worldly"},
			regexPattern: `\bworld\b`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 6,
					EndByteUnstyledContent:   11,
				},
			},
		},
		{
			name:         "digit pattern",
			items:        []string{"line 123 has numbers 456"},
			regexPattern: `\d+`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 5,
					EndByteUnstyledContent:   8,
				},
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 21,
					EndByteUnstyledContent:   24,
				},
			},
		},
		{
			name:         "case insensitive pattern",
			items:        []string{"Hello HELLO hello"},
			regexPattern: `(?i)hello`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   5,
				},
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 6,
					EndByteUnstyledContent:   11,
				},
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 12,
					EndByteUnstyledContent:   17,
				},
			},
		},
		{
			name:         "multiple items with matches",
			items:        []string{"error: failed", "warning: issue", "error: timeout"},
			regexPattern: `error:`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   6,
				},
				{
					ItemIndex:                2,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   6,
				},
			},
		},
		{
			name:         "capturing groups",
			items:        []string{"user: john", "user: jane"},
			regexPattern: `user: (\w+)`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   10,
				},
				{
					ItemIndex:                1,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   10,
				},
			},
		},
		{
			name:         "dot metacharacter",
			items:        []string{"a1b", "a.b", "axb"},
			regexPattern: `a.b`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   3,
				},
				{
					ItemIndex:                1,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   3,
				},
				{
					ItemIndex:                2,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   3,
				},
			},
		},
		{
			name:         "anchored pattern",
			items:        []string{"start middle", "middle end", "start end"},
			regexPattern: `^start`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   5,
				},
				{
					ItemIndex:                2,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   5,
				},
			},
		},
		{
			name:         "unicode with regex",
			items:        []string{"世界 test 🌟", "test 世界"},
			regexPattern: `test`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 7,
					EndByteUnstyledContent:   11,
				},
				{
					ItemIndex:                1,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   4,
				},
			},
		},
		{
			name:         "overlapping matches not possible with regex",
			items:        []string{"aaa"},
			regexPattern: `aa`,
			expected: []Match{
				{
					ItemIndex:                0,
					StartByteUnstyledContent: 0,
					EndByteUnstyledContent:   2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractMatchesRegex(tt.items, tt.regexPattern)

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
					actual.StartByteUnstyledContent != expected.StartByteUnstyledContent ||
					actual.EndByteUnstyledContent != expected.EndByteUnstyledContent {
					t.Errorf("highlight %d: expected ItemIndex=%d StartByteUnstyledContent=%d EndByteUnstyledContent=%d, got ItemIndex=%d StartByteUnstyledContent=%d EndByteUnstyledContent=%d",
						i, expected.ItemIndex, expected.StartByteUnstyledContent, expected.EndByteUnstyledContent,
						actual.ItemIndex, actual.StartByteUnstyledContent, actual.EndByteUnstyledContent)
				}
			}
		})
	}
}
