⏺ Detailed Analysis: Why TestMatchNavigationManyMatchesWrapped is Slow

Based on the CPU profiling analysis, here are the exact reasons why this test is slow:

1. The Problem Setup

The test creates a 10,000 character string of repeated 'a' characters (strings.Repeat("a", numAs) where numAs = 10000), then
applies a filter for the character 'a'. This results in 10,000 individual matches that need to be highlighted.

2. Primary Performance Bottlenecks

Bottleneck #1: O(n²) String Processing in highlightLine (62% of CPU time)

- Location: viewport/linebuffer/util.go:107 (highlightLine function)
- CPU Impact: 560ms out of 900ms total (62.22%)
- Root Cause: The function processes each character individually with multiple inefficient operations:
    - strings.HasPrefix() calls on every character: Line 120 calls strings.HasPrefix(styledLine[i:], "\x1b[") for each
      position, creating substring allocations
    - strings.Index() calls for ANSI parsing: Line 123 calls strings.Index(styledLine[i:], "m") repeatedly
    - getNonAnsiBytes() function calls: Line 141 calls this function which itself iterates character-by-character to extract
      non-ANSI content

Bottleneck #2: Excessive Memory Allocation and String Building (18.89% of CPU time)

- Location: Various string operations throughout the highlighting pipeline
- CPU Impact: 170ms (18.89% from runtime.mallocgc)
- Root Cause:
    - 10,000 individual highlight operations each allocating new strings.Builder objects
    - Repeated string concatenation in result.WriteString() calls
    - Memory pressure from creating 10,000+ styled string segments

Bottleneck #3: ANSI Sequence Processing in removeEmptyAnsiSequences (16.67% of CPU time)

- Location: viewport/linebuffer/util.go:654
- CPU Impact: 150ms (16.67%)
- Root Cause: After highlighting each segment, this function scans the entire result string character-by-character to remove
  empty ANSI sequences, leading to another O(n) pass over the data

3. Algorithmic Complexity Issues

Quadratic Behavior:

For 10,000 'a' characters, the code performs:
- 10,000 calls to highlightString() (one per match)
- Each call triggers highlightLine() which scans character-by-character
- Each highlightLine() call processes the styled string with multiple substring operations
- Result: ~10,000 × average_string_length character-level operations

Inefficient Text Processing:

// Line 141 in highlightLine - called 10,000 times
textToCheck := getNonAnsiBytes(styledLine, i, len(highlight))
if textToCheck == highlight {
// highlighting logic
}
This creates a new string allocation for every potential match position.

4. Specific Inefficient Code Paths

Most Expensive Path: highlightString → highlightLine → getNonAnsiBytes

TestMatchNavigationManyMatchesWrapped (650ms)
└── updateMatchingItems (480ms)
└── Viewport.SetContent (470ms)
└── highlightString (620ms) - called ~10,000 times
└── highlightLine (560ms) - processes each 'a' match
├── strings.HasPrefix() - every character
├── strings.Index() - ANSI parsing
├── getNonAnsiBytes() - string extraction
└── removeEmptyAnsiSequences (150ms) - cleanup

Memory Allocation Hotspots:

- Line 112: highlightStyle.Render(highlight) - 10,000 lipgloss render calls
- Line 113: var result strings.Builder - 10,000 builder allocations
- Line 130: activeStyles = append(activeStyles, ansi) - slice growth
- Line 132: result.WriteString(ansi) - repeated string building

5. Why This is Particularly Bad for Wrapped Text

The test uses viewport.WithWrapText[viewport.Item](true), which means:
- The 10,000 character string gets wrapped across multiple display lines
- Each wrapped segment gets processed separately for highlighting
- This multiplies the number of highlighting operations beyond just the 10,000 character matches

Root Cause Summary

The performance bottleneck stems from applying individual character-level highlighting to 10,000 matches using an algorithm
that:
1. Processes each match individually (O(n) matches)
2. Scans character-by-character for each match (O(m) characters per match)
3. Allocates new strings repeatedly (memory pressure)
4. Post-processes ANSI sequences after each highlighting operation

This results in O(n×m) complexity where n=number of matches (10,000) and m=average string length, explaining why a simple
test takes over 1 second to complete.

THE PLAN: leave in place post processing, but improve the performance of the rest.
