package internal

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// CmpStr compares two strings and fails the test if they are not equal
func CmpStr(t *testing.T, expected, actual string) {
	_, file, line, _ := runtime.Caller(1)
	testName := t.Name()
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("\nTest %q failed at %s:%d\nDiff (-expected +actual):\n%s", testName, file, line, diff)
	}
}

// RunWithTimeout runs a test function with a timeout.
func RunWithTimeout(t *testing.T, runTest func(t *testing.T), timeout time.Duration) {
	t.Helper()

	// warmup runs
	for i := 0; i < 3; i++ {
		runTest(t)
	}

	// actual measured runs
	var durations []time.Duration
	for i := 0; i < 3; i++ {
		done := make(chan struct{})
		var testErr error
		start := time.Now()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					testErr = fmt.Errorf("test panicked: %v", r)
				}
				close(done)
			}()

			subT := &testing.T{}
			runTest(subT)
			if subT.Failed() {
				testErr = fmt.Errorf("test failed in goroutine")
			}
		}()

		select {
		case <-done:
			if testErr != nil {
				t.Fatal(testErr)
			}
			durations = append(durations, time.Since(start))
		case <-time.After(timeout):
			t.Fatalf("Test took too long: %v", timeout)
		}

		runtime.GC()
		time.Sleep(time.Millisecond * 10)
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	median := durations[len(durations)/2]
	t.Logf("Test timing: median=%v min=%v max=%v",
		median, durations[0], durations[len(durations)-1])
}

func Pad(width, height int, lines []string) string {
	var res []string
	for _, line := range lines {
		resLine := line
		numSpaces := width - lipgloss.Width(line)
		if numSpaces > 0 {
			resLine += strings.Repeat(" ", numSpaces)
		}
		res = append(res, resLine)
	}
	numEmptyLines := height - len(lines)
	for i := 0; i < numEmptyLines; i++ {
		res = append(res, strings.Repeat(" ", width))
	}
	return strings.Join(res, "\n")
}
