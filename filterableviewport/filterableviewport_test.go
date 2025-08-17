package filterableviewport

import (
	"fmt"
	"strings"
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
		WithText[viewport.Item]("Filter:", "Type to filter..."),
	)
	fv.SetContent(stringsToItems([]string{
		"Line 1",
		"Line 2",
		"Line 3",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Type to filter...",
		"Line 1",
		"Line 2",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNew_LongText(t *testing.T) {
	fv := New[viewport.Item](
		10, // whenEmpty is longer than this
		4,
		WithText[viewport.Item]("Filter:", "Type to filter..."),
	)
	fv.SetContent(stringsToItems([]string{
		"Line 1",
		"Line 2",
		"Line 3",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Type to...",
		"Line 1",
		"Line 2",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
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

func TestGetWidthHeight(t *testing.T) {
	fv := New[viewport.Item](25, 8)
	if fv.GetWidth() != 25 {
		t.Errorf("expected width 25, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 8 {
		t.Errorf("expected height 8, got %d", fv.GetHeight())
	}
}

func TestFilterFocused_Initial(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	if fv.FilterFocused() {
		t.Error("filter should not be focused initially")
	}
}

func TestEmptyContent(t *testing.T) {
	fv := New[viewport.Item](20, 4, WithText[viewport.Item]("Filter:", "No items"))
	fv.SetContent([]viewport.Item{})
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No items",
		"",
		"",
		"",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnly_True(t *testing.T) {
	fv := New[viewport.Item](
		20,
		4,
		WithText[viewport.Item]("Filter:", "Type to filter..."),
		WithMatchesOnly[viewport.Item](true),
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Type to filter...",
		"apple",
		"banana",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnly_False(t *testing.T) {
	fv := New[viewport.Item](
		20,
		4,
		WithText[viewport.Item]("Filter:", "Type to filter..."),
		WithMatchesOnly[viewport.Item](false),
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Type to filter...",
		"apple",
		"banana",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithCanToggleMatchesOnly_False(t *testing.T) {
	fv := New[viewport.Item](
		20,
		4,
		WithCanToggleMatchesOnly[viewport.Item](false),
	)
	if fv.canToggleMatchesOnly {
		t.Error("canToggleMatchesOnly should be false")
	}
}

func TestNegativeDimensions(t *testing.T) {
	fv := New[viewport.Item](-5, -3)
	if fv.GetWidth() != 0 {
		t.Errorf("expected width 0 for negative input, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 1 {
		t.Errorf("expected height 1 for negative input (filter line), got %d", fv.GetHeight())
	}
}

func TestNilContent(t *testing.T) {
	fv := New[viewport.Item](20, 4, WithText[viewport.Item]("Filter:", "Empty"))
	fv.SetContent(nil)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Empty",
		"",
		"",
		"",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestSingleItemContent(t *testing.T) {
	fv := New[viewport.Item](20, 4, WithText[viewport.Item]("Filter:", "Type..."))
	fv.SetContent(stringsToItems([]string{"single item"}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Type...",
		"single item",
		"",
		"",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestManyItemsContent(t *testing.T) {
	fv := New[viewport.Item](15, 3, WithText[viewport.Item]("F:", "Filter"))
	items := make([]string, 100)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}
	fv.SetContent(stringsToItems(items))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Filter",
		"Item 1",
		"1% (1/100)",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestDefaultText(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"test"}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"test",
		"",
		"",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestZeroDimensions(t *testing.T) {
	fv := New[viewport.Item](0, 0)
	if fv.GetWidth() != 0 {
		t.Errorf("expected width 0, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 1 {
		t.Errorf("expected height 1 (filter line), got %d", fv.GetHeight())
	}
}

func TestFilterKey_EnterEditMode(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(filterKeyMsg)

	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing filter key")
	}
}

func TestRegexFilterKey_EnterEditMode(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(regexFilterKeyMsg)

	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing regex filter key")
	}
	if !fv.isRegexMode {
		t.Error("should be in regex mode")
	}
}

func TestApplyFilterKey(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(filterKeyMsg)

	fv, _ = fv.Update(applyFilterKeyMsg)

	if fv.FilterFocused() {
		t.Error("filter should not be focused after applying filter")
	}
}

func TestCancelFilterKey(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(filterKeyMsg)

	fv, _ = fv.Update(cancelFilterKeyMsg)

	if fv.FilterFocused() {
		t.Error("filter should not be focused after canceling")
	}
	if fv.isRegexMode {
		t.Error("should not be in regex mode after canceling")
	}
}

func TestToggleMatchesOnlyKey(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	initialMatchesOnly := fv.matchesOnly

	fv, _ = fv.Update(toggleMatchesKeyMsg)

	if fv.matchesOnly == initialMatchesOnly {
		t.Error("matches only mode should have toggled")
	}
}

func TestToggleMatchesOnlyKey_Disabled(t *testing.T) {
	fv := New[viewport.Item](20, 4, WithCanToggleMatchesOnly[viewport.Item](false))
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	initialMatchesOnly := fv.matchesOnly

	fv, _ = fv.Update(toggleMatchesKeyMsg)

	if fv.matchesOnly != initialMatchesOnly {
		t.Error("matches only mode should not have toggled when disabled")
	}
}

func TestFilterTextInput_TypingInEditMode(t *testing.T) {
	fv := New[viewport.Item](50, 4, WithText[viewport.Item]("Filter:", "Type..."))
	fv.SetContent(stringsToItems([]string{"apple", "banana", "cherry"}))

	fv, _ = fv.Update(filterKeyMsg)

	fv, _ = fv.Update(typeAKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: a  (2/3 matches)                ",
		"apple                                             ",
		"banana                                            ",
		"66% (2/3)                                         ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilter_ValidPattern(t *testing.T) {
	fv := New[viewport.Item](50, 4, WithText[viewport.Item]("Filter:", "Type..."))
	fv.SetContent(stringsToItems([]string{"apple", "banana", "apricot"}))

	fv, _ = fv.Update(regexFilterKeyMsg)

	fv, _ = fv.Update(typeAKeyMsg)

	fv, _ = fv.Update(typePKeyMsg)

	fv, _ = fv.Update(typePlusKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: ap+  (2/3 matches)              ",
		"apple                                             ",
		"banana                                            ",
		"66% (2/3)                                         ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilter_InvalidPattern(t *testing.T) {
	fv := New[viewport.Item](50, 4, WithText[viewport.Item]("Filter:", "Type..."))
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(regexFilterKeyMsg)

	fv, _ = fv.Update(typeLeftBracketKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: [  (no matches)                 ",
		"apple                                             ",
		"banana                                            ",
		"100% (2/2)                                        ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestMatchesOnlyMode_FiltersContent(t *testing.T) {
	fv := New[viewport.Item](50, 5, WithMatchesOnly[viewport.Item](true))
	fv.SetContent(stringsToItems([]string{"apple", "banana", "apricot"}))

	fv, _ = fv.Update(filterKeyMsg)

	fv, _ = fv.Update(typeAKeyMsg)

	fv, _ = fv.Update(typePKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] ap  (2/3 matches) showing matches only  ",
		"apple                                             ",
		"apricot                                           ",
		"                                                  ",
		"                                                  ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNoMatches_ShowsNoMatchesText(t *testing.T) {
	fv := New[viewport.Item](50, 4, WithText[viewport.Item]("Filter:", "Type..."))
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(filterKeyMsg)

	fv, _ = fv.Update(typeXKeyMsg)

	fv, _ = fv.Update(typeYKeyMsg)

	fv, _ = fv.Update(typeZKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: xyz  (no matches)               ",
		"apple                                             ",
		"banana                                            ",
		"100% (2/2)                                        ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithKeyMap_CustomKeys(t *testing.T) {
	customKeyMap := DefaultKeyMap()
	customKeyMap.FilterKey = key.NewBinding(key.WithKeys("ctrl+f"))

	fv := New[viewport.Item](20, 4, WithKeyMap[viewport.Item](customKeyMap))
	fv.SetContent(stringsToItems([]string{"test"}))

	if fv.keyMap.FilterKey.Keys()[0] != "ctrl+f" {
		t.Error("custom key map should be applied")
	}
}

func TestInit_ReturnsNil(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	cmd := fv.Init()
	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestUpdate_PassesThroughToViewport(t *testing.T) {
	fv := New[viewport.Item](20, 4)
	fv.SetContent(stringsToItems([]string{"line1", "line2", "line3"}))

	fv, _ = fv.Update(downKeyMsg)

	view := fv.View()
	if !strings.Contains(view, "line2") {
		t.Error("viewport should respond to navigation keys when not filtering")
	}
}

func TestApplyEmptyFilter_ShowsWhenEmptyText(t *testing.T) {
	fv := New[viewport.Item](30, 4, WithText[viewport.Item]("Filter:", "No filter applied"))
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(filterKeyMsg)

	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No filter applied             ",
		"apple                         ",
		"banana                        ",
		"100% (2/2)                    ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestEditingEmptyFilter_ShowsEditingInterface(t *testing.T) {
	fv := New[viewport.Item](50, 4, WithText[viewport.Item]("Filter:", "No filter applied"))
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))

	fv, _ = fv.Update(filterKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter:   (2/2 matches)                  ",
		"apple                                             ",
		"banana                                            ",
		"100% (2/2)                                        ",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func stringsToItems(vals []string) []viewport.Item {
	items := make([]viewport.Item, len(vals))
	for i, s := range vals {
		items[i] = viewport.Item{LineBuffer: linebuffer.New(s)}
	}
	return items
}
