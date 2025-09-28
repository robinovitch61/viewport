package filterableviewport

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/robinovitch61/bubbleo/internal"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/item"
)

// Note: this won't be necessary in future charm library versions
func init() {
	// Force TrueColor profile for consistent test output
	lipgloss.SetColorProfile(termenv.TrueColor)
}

type object struct {
	item item.Item
}

func (i object) GetItem() item.Item {
	return i.item
}

//func objectsEqual(a, b object) bool {
//	if a.item == nil || b.item == nil {
//		return a.item == b.item
//	}
//	return a.item.Content() == b.item.Content()
//}

var _ viewport.Object = object{}

var (
	filterKeyMsg        = internal.MakeKeyMsg('/')
	regexFilterKeyMsg   = internal.MakeKeyMsg('r')
	applyFilterKeyMsg   = tea.KeyMsg{Type: tea.KeyEnter}
	cancelFilterKeyMsg  = tea.KeyMsg{Type: tea.KeyEsc}
	toggleMatchesKeyMsg = internal.MakeKeyMsg('o')
	nextMatchKeyMsg     = internal.MakeKeyMsg('n')
	prevMatchKeyMsg     = internal.MakeKeyMsg('N')
	downKeyMsg          = tea.KeyMsg{Type: tea.KeyDown}

	footerStyle              = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	highlightStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	highlightStyleIfSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Underline(true)
	selectedItemStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	viewportStyles           = viewport.Styles{
		FooterStyle:              footerStyle,
		HighlightStyle:           highlightStyle,
		HighlightStyleIfSelected: highlightStyleIfSelected,
		SelectedItemStyle:        selectedItemStyle,
	}

	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Reverse(true)
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("11"))
	unfocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("12"))
	matchStyles    = MatchStyles{
		Focused:   focusedStyle,
		Unfocused: unfocusedStyle,
	}
	filterableViewportStyles = Styles{
		CursorStyle: cursorStyle,
		Match:       matchStyles,
	}
)

func makeFilterableViewport(
	width int,
	height int,
	vpOptions []viewport.Option[object],
	fvOptions []Option[object],
) *Model[object] {
	// use default viewport test styles, will be overridden by options if passed in
	defaultTestVpStylesOption := viewport.WithStyles[object](viewportStyles)
	vpOptions = append([]viewport.Option[object]{defaultTestVpStylesOption}, vpOptions...)

	// use default filterable viewport test styles, will be overridden by options if passed in
	defaultTestFvStylesOption := WithStyles[object](filterableViewportStyles)
	fvOptions = append([]Option[object]{defaultTestFvStylesOption}, fvOptions...)

	vp := viewport.New[object](width, height, vpOptions...)
	return New[object](vp, fvOptions...)
}

func TestNew(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithEmptyText[object]("No Filter"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"Line 1",
		"Line 2",
		"Line 3",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"Line 1",
		"Line 2",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNewLongText(t *testing.T) {
	fv := makeFilterableViewport(
		10, // emptyText is longer than this
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithEmptyText[object]("Nada Filter"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"Line 1",
		"Line 2",
		"Line 3",
	}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Nada Fi...",
		"Line 1",
		"Line 2",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNewWidthHeight(t *testing.T) {
	fv := makeFilterableViewport(
		25,
		8,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	if fv.GetWidth() != 25 {
		t.Errorf("expected width 25, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 8 {
		t.Errorf("expected height 8, got %d", fv.GetHeight())
	}
}

func TestZeroDimensions(t *testing.T) {
	fv := makeFilterableViewport(
		0,
		0,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	if fv.GetWidth() != 0 {
		t.Errorf("expected width 0, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 0 {
		t.Errorf("expected height 0, got %d", fv.GetHeight())
	}
	internal.CmpStr(t, "", fv.View())
}

func TestNegativeDimensions(t *testing.T) {
	fv := makeFilterableViewport(
		-5,
		-3,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	if fv.GetWidth() != 0 {
		t.Errorf("expected width 0 for negative input, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 0 {
		t.Errorf("expected height 0 for negative input, got %d", fv.GetHeight())
	}
	internal.CmpStr(t, "", fv.View())
}

func TestSetWidthSetHeight(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetWidth(30)
	if fv.GetWidth() != 30 {
		t.Errorf("expected width 30, got %d", fv.GetWidth())
	}

	fv.SetHeight(6)
	if fv.GetHeight() != 6 {
		t.Errorf("expected height 6, got %d", fv.GetHeight())
	}
}

func TestFilterFocused_Initial(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	if fv.FilterFocused() {
		t.Error("filter should not be focused initially")
	}
}

func TestEmptyContent(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithEmptyText[object]("No filter"),
		},
	)
	fv.SetObjects([]object{})
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No filter",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnly_True(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithMatchingItemsOnly[object](true),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: p" + cursorStyle.Render(" ") + " (1/2 matches on 1 items) showing matches only",
		"a" + focusedStyle.Render("p") + unfocusedStyle.Render("p") + "le",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnly_False(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithMatchingItemsOnly[object](false),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: p" + cursorStyle.Render(" ") + " (1/2 matches on 1 items)",
		"a" + focusedStyle.Render("p") + unfocusedStyle.Render("p") + "le",
		"banana",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}
func TestWithCanToggleMatchesOnly_True(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithCanToggleMatchingItemsOnly[object](true),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (1/2 matches on 1 items)",
		"a" + focusedStyle.Render("p") + unfocusedStyle.Render("p") + "le",
		"banana",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(toggleMatchesKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (1/2 matches on 1 items) showing matches only",
		"a" + focusedStyle.Render("p") + unfocusedStyle.Render("p") + "le",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithCanToggleMatchesOnly_False(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithCanToggleMatchingItemsOnly[object](false),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (1/2 matches on 1 items)",
		"a" + focusedStyle.Render("p") + unfocusedStyle.Render("p") + "le",
		"banana",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(toggleMatchesKeyMsg)
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNilContent(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithEmptyText[object]("No Filter"),
		},
	)
	fv.SetObjects(nil)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestDefaultText(t *testing.T) {
	fv := makeFilterableViewport(
		40,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"test"}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"test",
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (no matches)",
		"test",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterKeyFocus(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing filter key")
	}
}

func TestRegexFilterKeyFocus(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing regex filter key")
	}
}

func TestApplyFilterKey(t *testing.T) {
	fv := makeFilterableViewport(
		40,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	fv, _ = fv.Update(applyFilterKeyMsg)
	if fv.FilterFocused() {
		t.Error("filter should not be focused after applying filter")
	}
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] a  (1/4 matches on 2 items)",
		focusedStyle.Render("a") + "pple",
		"b" + unfocusedStyle.Render("a") + "n" + unfocusedStyle.Render("a") + "n" + unfocusedStyle.Render("a"),
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestCancelFilterKey(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	fv, _ = fv.Update(cancelFilterKeyMsg)
	if fv.FilterFocused() {
		t.Error("filter should not be focused after canceling")
	}
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"apple",
		"banana",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilter_ValidPattern(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
		},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana", "apricot"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	fv, _ = fv.Update(internal.MakeKeyMsg('+'))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: ap+" + cursorStyle.Render(" ") + " (1/2 matches on 2 items)",
		focusedStyle.Render("app") + "le",
		"banana",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilter_InvalidPattern(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
		},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('['))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: [" + cursorStyle.Render(" ") + " (no matches)",
		"apple",
		"banana",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNoMatches_ShowsNoMatchesText(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('x'))
	fv, _ = fv.Update(internal.MakeKeyMsg('y'))
	fv, _ = fv.Update(internal.MakeKeyMsg('z'))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] xyz" + cursorStyle.Render(" ") + " (no matches)",
		"apple",
		"banana",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithKeyMap(t *testing.T) {
	customKeyMap := DefaultKeyMap()
	customKeyMap.FilterKey = key.NewBinding(key.WithKeys("g"))
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithKeyMap[object](customKeyMap),
		},
	)
	fv.SetObjects(stringsToItems([]string{"test"}))
	fv, _ = fv.Update(filterKeyMsg) // should not match custom key
	if fv.FilterFocused() {
		t.Error("filter should not be focused with custom keymap")
	}
}

func TestViewportControls(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		3,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"line1", "line2", "line3"}))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"line1",
		footerStyle.Render("33% (1/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(downKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"line2",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestApplyEmptyFilter_ShowsWhenEmptyText(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithEmptyText[object]("No filter applied"),
		},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No filter applied",
		"apple",
		"banana",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestEditingEmptyFilter_ShowsEditingMessage(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
		},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: " + cursorStyle.Render(" ") + " type to filter",
		"apple",
		"banana",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestSpecialKeysWhileFiltering(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithCanToggleMatchingItemsOnly[object](true),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"apple",
		"book",
		"food",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	fv, _ = fv.Update(toggleMatchesKeyMsg) // 'o'
	fv, _ = fv.Update(nextMatchKeyMsg)     // 'n'
	fv, _ = fv.Update(prevMatchKeyMsg)     // 'N'
	fv, _ = fv.Update(filterKeyMsg)        // '/'
	fv, _ = fv.Update(regexFilterKeyMsg)   // 'r'
	expectedViewAfterO := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] ponN/r" + cursorStyle.Render(" ") + " (no matches)",
		"apple",
		"book",
		footerStyle.Render("50% (2/4)"),
	})
	internal.CmpStr(t, expectedViewAfterO, fv.View())
}

func TestAnsiEscapeCodesNotMatched(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		internal.RedFg.Render("apple"),
		internal.RedFg.Render("book"),
		internal.RedFg.Render("food"),
		internal.RedFg.Render("cherry"),
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "x1b" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] x1b  (no matches)",
		internal.RedFg.Render("apple"),
		internal.RedFg.Render("book"),
		footerStyle.Render("50% (2/4)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestMatchNavigationWithNoMatches(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('x'))
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] x  (no matches)",
		"apple",
		"banana",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedView, fv.View())
}

func TestMatchNavigationWithOverlappingMatches(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{"aaa"}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "aa" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] aa  (1/1 matches on 1 items)",
		focusedStyle.Render("aa") + "a",
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

func TestMatchNavigationWithAllItemsWrap(t *testing.T) {
	fv := makeFilterableViewport(
		7,
		6,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{
			WithStyles[object](Styles{
				Match: matchStyles,
			}),
			WithMatchingItemsOnly[object](false),
			WithEmptyText[object]("None"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"hi there",
		"hi over there",
		"no match",
	}))
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"None",
		"hi ther",
		"e",
		"hi over",
		" there",
		footerStyle.Render("66% ..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "there" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exa...",
		"hi " + focusedStyle.Render("ther"),
		focusedStyle.Render("e"),
		"hi over",
		" " + unfocusedStyle.Render("there"),
		footerStyle.Render("66% ..."),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exa...",
		"hi " + unfocusedStyle.Render("ther"),
		unfocusedStyle.Render("e"),
		"hi over",
		" " + focusedStyle.Render("there"),
		footerStyle.Render("66% ..."),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

func TestMatchNavigationWithMatchingItemsOnlyWrap(t *testing.T) {
	fv := makeFilterableViewport(
		7,
		6,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{
			WithStyles[object](Styles{
				Match: matchStyles,
			}),
			WithMatchingItemsOnly[object](true),
			WithEmptyText[object]("None"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"hi there",
		"hi over there",
		"no match",
	}))
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"None",
		"hi ther",
		"e",
		"hi over",
		" there",
		footerStyle.Render("66% ..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "there" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exa...",
		"hi " + focusedStyle.Render("ther"),
		focusedStyle.Render("e"),
		"hi over",
		" " + unfocusedStyle.Render("there"),
		footerStyle.Render("100%..."),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exa...",
		"hi " + unfocusedStyle.Render("ther"),
		unfocusedStyle.Render("e"),
		"hi over",
		" " + focusedStyle.Render("there"),
		footerStyle.Render("100%..."),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

// TODO LEO: implement scrolling to fix this
//func TestMatchNavigationWrap_Overflow(t *testing.T) {
//	fv := makeFilterableViewport(
//		20,
//		5,
//		[]viewport.Option[viewport.Item]{
//			viewport.WithWrapText[viewport.Item](true),
//		},
//		[]Option[viewport.Item]{},
//	)
//	fv.SetObjects(stringsToItems([]string{
//		strings.Repeat("a", 100) + "goose" + strings.Repeat("a", 100),
//	}))
//	fv, _ = fv.Update(filterKeyMsg)
//	for _, c := range "goose" {
//		fv, _ = fv.Update(internal.MakeKeyMsg(c))
//	}
//	fv, _ = fv.Update(applyFilterKeyMsg)
//	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
//		"[exact] goose  (1...",
//		strings.Repeat("a", 20),
//		focusedStyle.GetItem("goose") + strings.Repeat("a", 15),
//		strings.Repeat("a", 20),
//		footerStyle.GetItem("99% (1/1)"),
//	})
//	internal.CmpStr(t, expected, fv.View())
//}

func TestMatchNavigationNoWrap(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		10,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"duck duck duck duck duck duck duck duck duck duck goose",
		"duck duck duck duck duck goose duck duck duck duck duck",
		"goose duck duck duck duck duck duck duck duck duck duck",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "goose" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] goose  (1/3 matches...",
		"...k duck duck duck duck " + focusedStyle.Render("goose"),
		unfocusedStyle.Render("...se") + " duck duck duck duck duck",
		"...ck duck duck duck duck duck",
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] goose  (2/3 matches...",
		"...duck duck duck duck duck...",
		"...duck duck " + focusedStyle.Render("goose") + " duck duc...",
		"... duck duck duck duck duc...",
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedThirdMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] goose  (3/3 matches...",
		"duck duck duck duck duck du...",
		"duck duck duck duck duck " + unfocusedStyle.Render("go..."),
		focusedStyle.Render("goose") + " duck duck duck duck d...",
	})
	internal.CmpStr(t, expectedThirdMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

func TestMatchNavigationNoWrapUnicode(t *testing.T) {
	fv := makeFilterableViewport(
		32,
		10,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		// a (1w, 1b), 💖 (2w, 4b)
		"💖💖💖💖💖💖💖💖 hi aaaaaaaaaaaaaaaa",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "hi" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] hi  (1/1 matches on 1...",
		"..💖💖💖💖💖💖 " + focusedStyle.Render("hi") + " aaaaaaaaaaa...",
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

//func TestMatchNavigationSelectionEnabled(t *testing.T) {
//	// TODO LEO
//}

func TestMatchNavigationManyMatchesWrap(t *testing.T) {
	fv := makeFilterableViewport(
		100,
		50,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{},
	)
	numAs := 10000
	fv.SetObjects(stringsToItems([]string{
		internal.RedFg.Render(strings.Repeat("a", numAs)),
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	fv, _ = fv.Update(applyFilterKeyMsg)
	firstRows := []string{
		fmt.Sprintf("[exact] a  (1/%d matches on 1 items)", numAs),
		focusedStyle.Render("a") + strings.Repeat(unfocusedStyle.Render("a"), fv.GetWidth()-1),
	}
	rest := make([]string, fv.GetHeight()-3)
	for i := range rest {
		rest[i] = strings.Repeat(unfocusedStyle.Render("a"), fv.GetWidth())
	}
	rest = append(rest, footerStyle.Render("99% (1/1)"))
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), append(firstRows, rest...))
	internal.CmpStr(t, expected, fv.View())
}

func TestMatchNavigationManyMatchesWrapTwoItems(t *testing.T) {
	runTest := func(t *testing.T) {
		fv := makeFilterableViewport(
			100,
			50,
			[]viewport.Option[object]{
				viewport.WithWrapText[object](true),
			},
			[]Option[object]{},
		)
		numAs := 5000
		fv.SetObjects(stringsToItems([]string{
			internal.RedFg.Render(strings.Repeat("a", numAs)),
		}))
		fv, _ = fv.Update(filterKeyMsg)
		fv, _ = fv.Update(internal.MakeKeyMsg('a'))
		fv, _ = fv.Update(applyFilterKeyMsg)
		firstRows := []string{
			fmt.Sprintf("[exact] a  (1/%d matches on 1 items)", numAs),
			focusedStyle.Render("a") + strings.Repeat(unfocusedStyle.Render("a"), fv.GetWidth()-1),
		}
		rest := make([]string, fv.GetHeight()-3)
		for i := range rest {
			rest[i] = strings.Repeat(unfocusedStyle.Render("a"), fv.GetWidth())
		}
		rest = append(rest, footerStyle.Render("99% (1/1)"))
		expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), append(firstRows, rest...))
		internal.CmpStr(t, expected, fv.View())

		numNext := 40
		for i := 0; i < numNext; i++ {
			fv, _ = fv.Update(nextMatchKeyMsg)
		}
		expectedAfterNext := []string{
			fmt.Sprintf("[exact] a  (%d/%d matches on 1 items)", numNext+1, numAs),
			strings.Repeat(unfocusedStyle.Render("a"), numNext) + focusedStyle.Render("a") + strings.Repeat(unfocusedStyle.Render("a"), fv.GetWidth()-numNext-1),
		}
		restAfterNext := make([]string, fv.GetHeight()-3)
		for i := range restAfterNext {
			restAfterNext[i] = strings.Repeat(unfocusedStyle.Render("a"), fv.GetWidth())
		}
		restAfterNext = append(restAfterNext, footerStyle.Render("99% (1/1)"))
		expectedAfterNextView := internal.Pad(fv.GetWidth(), fv.GetHeight(), append(expectedAfterNext, restAfterNext...))
		internal.CmpStr(t, expectedAfterNextView, fv.View())
	}
	internal.RunWithTimeout(t, runTest, 500*time.Millisecond)
}

// TODO LEO: add tests for match navigation with matches

// TODO LEO: add test for match navigation showing only matches

// TODO LEO: add test for when wrapped item goes off screen and focused match in the item is off screen (currently shows top lines item and not focused match)

// TODO LEO: add test that updating filter itself scrolls/pans screen to first match without needing to press n/N

// TODO LEO: test for multiple regex matches in a single line

// TODO LEO: test for when toggling wrap, current match should still be visible

// TODO LEO: with a huge block of sequential matches and wrapping on, highlighting disappears near the middle...

// TODO LEO: add timing test for scrolling through a large number of highlighted matches

func stringsToItems(vals []string) []object {
	items := make([]object, len(vals))
	for i, s := range vals {
		items[i] = object{item: item.NewItem(s)}
	}
	return items
}
