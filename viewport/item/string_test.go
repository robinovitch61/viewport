package item

import "testing"

func TestString_overflowsLeft(t *testing.T) {
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

func TestString_overflowsRight(t *testing.T) {
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

func TestString_replaceStartWithContinuation(t *testing.T) {
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

func TestString_replaceEndWithContinuation(t *testing.T) {
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

func TestString_getBytesLeftOfWidth(t *testing.T) {
	tests := []struct {
		name         string
		items        []SingleItem
		nBytes       int
		startItemIdx int
		widthToLeft  int
		expected     string
		shouldPanic  bool
	}{
		{
			name:         "empty items",
			items:        nil,
			nBytes:       1,
			startItemIdx: 0,
			widthToLeft:  0,
			expected:     "",
		},
		{
			name:         "negative bytes",
			items:        []SingleItem{New("abc")},
			nBytes:       -1,
			startItemIdx: 0,
			widthToLeft:  1,
			shouldPanic:  true,
		},
		{
			name:         "zero bytes",
			items:        []SingleItem{New("abc")},
			nBytes:       0,
			startItemIdx: 0,
			widthToLeft:  1,
			expected:     "",
		},
		{
			name:         "item index out of bounds",
			items:        []SingleItem{New("abc")},
			nBytes:       1,
			startItemIdx: 1,
			widthToLeft:  0,
			expected:     "",
		},
		{
			name:         "single item full content",
			items:        []SingleItem{New("abc")},
			nBytes:       3,
			startItemIdx: 0,
			widthToLeft:  3,
			expected:     "abc",
		},
		{
			name:         "single item partial content",
			items:        []SingleItem{New("abc")},
			nBytes:       2,
			startItemIdx: 0,
			widthToLeft:  2,
			expected:     "ab",
		},
		{
			name: "multiple items full content",
			items: []SingleItem{
				New("abc"),
				New("def"),
			},
			nBytes:       6,
			startItemIdx: 1,
			widthToLeft:  3,
			expected:     "abcdef",
		},
		{
			name: "multiple items partial content",
			items: []SingleItem{
				New("abc"),
				New("def"),
			},
			nBytes:       4,
			startItemIdx: 1,
			widthToLeft:  2,
			expected:     "bcde",
		},
		{
			name: "ignore ansi codes",
			items: []SingleItem{
				New("a" + redBg.Render("b") + "c"),
				New(redBg.Render("def")),
			},
			nBytes:       5,
			startItemIdx: 1,
			widthToLeft:  3,
			expected:     "bcdef",
		},
		// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
		{
			name: "unicode characters",
			items: []SingleItem{
				New("A💖中"),
				New("é"),
			},
			nBytes:       10,
			startItemIdx: 1,
			widthToLeft:  1,
			expected:     "💖中é",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assertPanic(t, func() {
					getBytesLeftOfWidth(tt.nBytes, tt.items, tt.startItemIdx, tt.widthToLeft)
				})
				return
			}

			if got := getBytesLeftOfWidth(tt.nBytes, tt.items, tt.startItemIdx, tt.widthToLeft); got != tt.expected {
				t.Errorf("getBytesLeftOfWidth() = %v, want %v", []byte(got), []byte(tt.expected))
			}
		})
	}
}

func TestString_getBytesRightOfWidth(t *testing.T) {
	tests := []struct {
		name         string
		items        []SingleItem
		nBytes       int
		endItemIdx   int
		widthToRight int
		expected     string
		shouldPanic  bool
	}{
		{
			name:         "empty items",
			items:        nil,
			nBytes:       1,
			endItemIdx:   0,
			widthToRight: 0,
			expected:     "",
		},
		{
			name:         "negative bytes",
			items:        []SingleItem{New("abc")},
			nBytes:       -1,
			endItemIdx:   0,
			widthToRight: 1,
			shouldPanic:  true,
		},
		{
			name:         "zero bytes",
			items:        []SingleItem{New("abc")},
			nBytes:       0,
			endItemIdx:   0,
			widthToRight: 1,
			expected:     "",
		},
		{
			name:         "item index out of bounds",
			items:        []SingleItem{New("abc")},
			nBytes:       1,
			endItemIdx:   1,
			widthToRight: 0,
			expected:     "",
		},
		{
			name:         "single item full content",
			items:        []SingleItem{New("abc")},
			nBytes:       3,
			endItemIdx:   0,
			widthToRight: 3,
			expected:     "abc",
		},
		{
			name:         "single item partial content",
			items:        []SingleItem{New("abc")},
			nBytes:       2,
			endItemIdx:   0,
			widthToRight: 2,
			expected:     "bc",
		},
		{
			name: "multiple items full content",
			items: []SingleItem{
				New("abc"),
				New("def"),
			},
			nBytes:       6,
			endItemIdx:   0,
			widthToRight: 3,
			expected:     "abcdef",
		},
		{
			name: "multiple items partial content",
			items: []SingleItem{
				New("abc"),
				New("def"),
			},
			nBytes:       4,
			endItemIdx:   0,
			widthToRight: 2,
			expected:     "bcde",
		},
		{
			name: "ignore ansi codes",
			items: []SingleItem{
				New("a" + redBg.Render("b") + "c"),
				New(redBg.Render("def")),
			},
			nBytes:       5,
			endItemIdx:   0,
			widthToRight: 2,
			expected:     "bcdef",
		},
		// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b)
		{
			name: "unicode characters",
			items: []SingleItem{
				New("A💖中"),
				New("é"),
			},
			nBytes:       10,
			endItemIdx:   0,
			widthToRight: 4,
			expected:     "💖中é",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assertPanic(t, func() {
					getBytesRightOfWidth(tt.nBytes, tt.items, tt.endItemIdx, tt.widthToRight)
				})
				return
			}

			if got := getBytesRightOfWidth(tt.nBytes, tt.items, tt.endItemIdx, tt.widthToRight); got != tt.expected {
				t.Errorf("getBytesRightOfWidth() = %v, want %v", []byte(got), []byte(tt.expected))
			}
		})
	}
}
