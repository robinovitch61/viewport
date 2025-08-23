package filterableviewport

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/bubbleo/internal"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
)

var (
	filterKeyMsg          = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	regexFilterKeyMsg     = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	applyFilterKeyMsg     = tea.KeyMsg{Type: tea.KeyEnter}
	cancelFilterKeyMsg    = tea.KeyMsg{Type: tea.KeyEsc}
	toggleMatchesKeyMsg   = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	nextMatchKeyMsg       = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	prevMatchKeyMsg       = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	downKeyMsg            = tea.KeyMsg{Type: tea.KeyDown}
	typeAKeyMsg           = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	typePKeyMsg           = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	typePlusKeyMsg        = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}}
	typeLeftBracketKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}}
	typeXKeyMsg           = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	typeYKeyMsg           = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	typeZKeyMsg           = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
)

func TestNew(t *testing.T) {
	fv := New[viewport.Item](
		20,
		4,
		WithPrefixText[viewport.Item]("Filter:"),
		WithEmptyText[viewport.Item]("No Filter"),
	)
	fv.SetContent(stringsToItems([]string{
		"Line 1",
		"Line 2",
		"Line 3",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"Line 1",
		"Line 2",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNewLongText(t *testing.T) {
	fv := New[viewport.Item](
		10, // emptyText is longer than this
		4,
		WithPrefixText[viewport.Item]("Filter:"),
		WithEmptyText[viewport.Item]("No Filter Present"),
	)
	fv.SetContent(stringsToItems([]string{
		"Line 1",
		"Line 2",
		"Line 3",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filt...",
		"Line 1",
		"Line 2",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNewWidthHeight(t *testing.T) {
	fv := New[viewport.Item](25, 8)
	if fv.GetWidth() != 25 {
		t.Errorf("expected width 25, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 8 {
		t.Errorf("expected height 8, got %d", fv.GetHeight())
	}
}

func TestZeroDimensions(t *testing.T) {
	fv := New[viewport.Item](0, 0)
	if fv.GetWidth() != 0 {
		t.Errorf("expected width 0, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 0 {
		t.Errorf("expected height 0, got %d", fv.GetHeight())
	}
	internal.CmpStr(t, "", fv.View())
}

func TestNegativeDimensions(t *testing.T) {
	fv := New[viewport.Item](-5, -3)
	if fv.GetWidth() != 0 {
		t.Errorf("expected width 0 for negative input, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 0 {
		t.Errorf("expected height 0 for negative input, got %d", fv.GetHeight())
	}
	internal.CmpStr(t, "", fv.View())
}

func TestSetWidth(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetWidth(30)
	if fv.GetWidth() != 30 {
		t.Errorf("expected width 30, got %d", fv.GetWidth())
	}
}

func TestSetHeight(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetHeight(6)
	if fv.GetHeight() != 6 {
		t.Errorf("expected height 6, got %d", fv.GetHeight())
	}
}

func TestFilterFocused_Initial(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	if fv.FilterFocused() {
		t.Error("filter should not be focused initially")
	}
}

func TestEmptyContent(t *testing.T) {
	fv := New[viewport.Item](20, 4, WithPrefixText[viewport.Item]("Filter:"), WithEmptyText[viewport.Item]("No filter"))
	fv.SetContent([]viewport.Item{})
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No filter",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnly_True(t *testing.T) {
	fv := New[viewport.Item](
		80,
		4,
		WithPrefixText[viewport.Item]("Filter:"),
		WithMatchingItemsOnly[viewport.Item](true),
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: p  (1/2 matches on 1 items) showing matches only",
		"apple",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnly_False(t *testing.T) {
	fv := New[viewport.Item](
		80,
		4,
		WithPrefixText[viewport.Item]("Filter:"),
		WithMatchingItemsOnly[viewport.Item](false),
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: p  (1/2 matches on 1 items)",
		"apple",
		"banana",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}
func TestWithCanToggleMatchesOnly_True(t *testing.T) {
	fv := New[viewport.Item](
		80,
		4,
		WithCanToggleMatchingItemsOnly[viewport.Item](true),
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (1/2 matches on 1 items)",
		"apple",
		"banana",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(toggleMatchesKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (1/2 matches on 1 items) showing matches only",
		"apple",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithCanToggleMatchesOnly_False(t *testing.T) {
	fv := New[viewport.Item](
		80,
		4,
		WithCanToggleMatchingItemsOnly[viewport.Item](false),
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (1/2 matches on 1 items)",
		"apple",
		"banana",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(toggleMatchesKeyMsg)
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNilContent(t *testing.T) {
	fv := New[viewport.Item](20, 4, WithPrefixText[viewport.Item]("Filter:"), WithEmptyText[viewport.Item]("No Filter"))
	fv.SetContent(nil)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"",
		"",
		"",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestDefaultText(t *testing.T) {
	fv := New[viewport.Item](40, 4)
	fv.SetContent(stringsToItems([]string{"test"}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"test",
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (no matches)",
		"test",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterKeyFocus(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing filter key")
	}
}

func TestRegexFilterKeyFocus(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing regex filter key")
	}
}

func TestApplyFilterKey(t *testing.T) {
	fv := New[viewport.Item](40, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeAKeyMsg)
	fv, _ = fv.Update(applyFilterKeyMsg)
	if fv.FilterFocused() {
		t.Error("filter should not be focused after applying filter")
	}
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] a  (1/4 matches on 2 items)",
		"apple",
		"banana",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestCancelFilterKey(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeAKeyMsg)
	fv, _ = fv.Update(cancelFilterKeyMsg)
	if fv.FilterFocused() {
		t.Error("filter should not be focused after canceling")
	}
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"apple",
		"banana",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilter_ValidPattern(t *testing.T) {
	fv := New[viewport.Item](
		50,
		4,
		WithPrefixText[viewport.Item]("Filter:"),
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana", "apricot"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	fv, _ = fv.Update(typeAKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	fv, _ = fv.Update(typePlusKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: ap+  (1/2 matches on 2 items)",
		"apple",
		"banana",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilter_InvalidPattern(t *testing.T) {
	fv := New[viewport.Item](
		50,
		4,
		WithPrefixText[viewport.Item]("Filter:"),
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	fv, _ = fv.Update(typeLeftBracketKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: [  (no matches)",
		"apple",
		"banana",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNoMatches_ShowsNoMatchesText(t *testing.T) {
	fv := New[viewport.Item](50, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeXKeyMsg)
	fv, _ = fv.Update(typeYKeyMsg)
	fv, _ = fv.Update(typeZKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] xyz  (no matches)",
		"apple",
		"banana",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithKeyMap(t *testing.T) {
	customKeyMap := DefaultKeyMap()
	customKeyMap.FilterKey = key.NewBinding(key.WithKeys("g"))
	fv := New[viewport.Item](20, 4, WithKeyMap[viewport.Item](customKeyMap))
	fv.SetContent(stringsToItems([]string{"test"}))
	fv, _ = fv.Update(filterKeyMsg) // should not match custom key
	if fv.FilterFocused() {
		t.Error("filter should not be focused with custom keymap")
	}
}

func TestViewportControls(t *testing.T) {
	fv := New[viewport.Item](20, 3)
	fv.SetContent(stringsToItems([]string{"line1", "line2", "line3"}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"line1",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(downKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"line2",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestApplyEmptyFilter_ShowsWhenEmptyText(t *testing.T) {
	fv := New[viewport.Item](30, 4, WithEmptyText[viewport.Item]("No filter applied"))
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No filter applied",
		"apple",
		"banana",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestEditingEmptyFilter_ShowsEditingMessage(t *testing.T) {
	fv := New[viewport.Item](50, 4, WithPrefixText[viewport.Item]("Filter:"))
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter:   type to filter                  ",
		"apple                                             ",
		"banana                                            ",
		"100% (2/2)                                        ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestSpecialKeysWhileFiltering(t *testing.T) {
	fv := New[viewport.Item](
		80,
		4,
		WithCanToggleMatchingItemsOnly[viewport.Item](true),
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"book",
		"food",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	fv, _ = fv.Update(toggleMatchesKeyMsg) // 'o'
	fv, _ = fv.Update(nextMatchKeyMsg)     // 'n'
	fv, _ = fv.Update(prevMatchKeyMsg)     // 'N'
	fv, _ = fv.Update(filterKeyMsg)        // '/'
	fv, _ = fv.Update(regexFilterKeyMsg)   // 'r'
	expectedViewAfterO := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] ponN/r  (no matches)",
		"apple",
		"book",
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedViewAfterO, fv.View())
}

func TestMatchNavigationWithNoMatches(t *testing.T) {
	fv := New[viewport.Item](50, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeXKeyMsg)
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] x  (no matches)",
		"apple",
		"banana",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedView, fv.View())
}

// TODO LEO: add tests for match navigation with matches

// TODO LEO: add test for match navigation showing only matches

// TODO LEO: add test for when wrapped item goes off screen and focused match in the item is off screen (currently shows top lines item and not focused match)

// TODO LEO: test for multiple regex matches in a single line

func stringsToItems(vals []string) []viewport.Item {
	items := make([]viewport.Item, len(vals))
	for i, s := range vals {
		items[i] = viewport.Item{LineBuffer: linebuffer.New(s)}
	}
	return items
}
