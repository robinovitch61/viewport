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

func TestFilterFocusedInitial(t *testing.T) {
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

func TestWithMatchesOnlyTrue(t *testing.T) {
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
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnlyFalse(t *testing.T) {
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
func TestWithCanToggleMatchesOnlyTrue(t *testing.T) {
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
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithCanToggleMatchesOnlyFalse(t *testing.T) {
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
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('p'))
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] p  (no matches)",
		"test",
		"",
		footerStyle.Render("100% (1/1)"),
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

func TestRegexFilterValidPattern(t *testing.T) {
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

func TestRegexFilterInvalidPattern(t *testing.T) {
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

func TestStyleOverlay(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetSelectionEnabled(true)

	fv.SetObjects(stringsToItems([]string{
		"apple pie",
		internal.RedFg.Render("apple") + " pie " + internal.BlueFg.Render("yum"),
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "apple pie" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	// match highlighting overrides both selection style and styled sections
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] apple pie  (1/2 matches on 2 items)",
		focusedStyle.Render("apple pie"),
		unfocusedStyle.Render("apple pie") + " " + internal.BlueFg.Render("yum"),
		footerStyle.Render("50% (1/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	// move selection down to second item
	fv, _ = fv.Update(downKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] apple pie  (1/2 matches on 2 items)",
		focusedStyle.Render("apple pie"),
		unfocusedStyle.Render("apple pie") + selectedItemStyle.Render(" ") + internal.BlueFg.Render("yum"),
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilterMultipleMatchesInSingleLine(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		6,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"the cat sat on the mat",
		"dog",
		"another the and the end",
	}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	// use regex pattern \bthe\b to match whole word "the"
	for _, c := range "\\bthe\\b" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	// should focus on first match in first line
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: \\bthe\\b  (1/4 matches on 2 items)",
		focusedStyle.Render("the") + " cat sat on " + unfocusedStyle.Render("the") + " mat",
		"dog",
		"another " + unfocusedStyle.Render("the") + " and " + unfocusedStyle.Render("the") + " end",
		"",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	// navigate to second match (still in first line)
	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: \\bthe\\b  (2/4 matches on 2 items)",
		unfocusedStyle.Render("the") + " cat sat on " + focusedStyle.Render("the") + " mat",
		"dog",
		"another " + unfocusedStyle.Render("the") + " and " + unfocusedStyle.Render("the") + " end",
		"",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	// navigate to third match (third line, first match)
	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedThirdMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: \\bthe\\b  (3/4 matches on 2 items)",
		unfocusedStyle.Render("the") + " cat sat on " + unfocusedStyle.Render("the") + " mat",
		"dog",
		"another " + focusedStyle.Render("the") + " and " + unfocusedStyle.Render("the") + " end",
		"",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedThirdMatch, fv.View())

	// navigate to fourth match (third line, second match)
	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedFourthMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[regex] Filter: \\bthe\\b  (4/4 matches on 2 items)",
		unfocusedStyle.Render("the") + " cat sat on " + unfocusedStyle.Render("the") + " mat",
		"dog",
		"another " + unfocusedStyle.Render("the") + " and " + focusedStyle.Render("the") + " end",
		"",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedFourthMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedFourthMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedThirdMatch, fv.View())
}

func TestNoMatchesShowsNoMatchesText(t *testing.T) {
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

func TestApplyEmptyFilterShowsWhenEmptyText(t *testing.T) {
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

func TestEditingEmptyFilterShowsEditingMessage(t *testing.T) {
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
		"",
		footerStyle.Render("100% (1/1)"),
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

func TestMatchNavigationWrapLineOffset(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		5,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		strings.Repeat("a", 100) + "goose" + strings.Repeat("a", 100),
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "goose" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] goose  (1...",
		strings.Repeat("a", 20),
		strings.Repeat("a", 20),
		focusedStyle.Render("goose") + strings.Repeat("a", 15),
		footerStyle.Render("99% (1/1)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestMatchNavigationWrappedLinesWithMatches(t *testing.T) {
	fv := makeFilterableViewport(
		4,
		6,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		strings.Repeat("a", 10),
		strings.Repeat("b", 15),
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "aaa" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		focusedStyle.Render("aaa") + unfocusedStyle.Render("a"),
		unfocusedStyle.Render("aa") + unfocusedStyle.Render("aa"),
		unfocusedStyle.Render("a") + "a",
		"bbbb",
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("aaa") + focusedStyle.Render("a"),
		focusedStyle.Render("aa") + unfocusedStyle.Render("aa"),
		unfocusedStyle.Render("a") + "a",
		"bbbb",
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(cancelFilterKeyMsg)
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "bbb" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		"aaaa",
		"aaaa",
		"aa",
		focusedStyle.Render("bbb") + unfocusedStyle.Render("b"),
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		"aaaa",
		"aa",
		unfocusedStyle.Render("bbb") + focusedStyle.Render("b"),
		focusedStyle.Render("bb") + unfocusedStyle.Render("bb"),
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestMatchNavigationWrappedLinesWithWrappedMatches(t *testing.T) {
	fv := makeFilterableViewport(
		4,
		5,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		strings.Repeat("a", 10),
		strings.Repeat("a", 15),
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for range 5 {
		fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		focusedStyle.Render("aaaa"),
		focusedStyle.Render("a") + unfocusedStyle.Render("aaa"),
		unfocusedStyle.Render("aa"),
		footerStyle.Render("5..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("aaaa"),
		unfocusedStyle.Render("a") + focusedStyle.Render("aaa"),
		focusedStyle.Render("aa"),
		footerStyle.Render("5..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("aa"),
		focusedStyle.Render("aaaa"),
		focusedStyle.Render("a") + unfocusedStyle.Render("aaa"),
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("aaaa"),
		unfocusedStyle.Render("a") + focusedStyle.Render("aaa"),
		focusedStyle.Render("aa") + unfocusedStyle.Render("aa"),
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("a") + unfocusedStyle.Render("aaa"),
		unfocusedStyle.Render("aa") + focusedStyle.Render("aa"),
		focusedStyle.Render("aaa"),
		footerStyle.Render("1..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("a") + focusedStyle.Render("aaa"),
		focusedStyle.Render("aa") + unfocusedStyle.Render("aa"),
		unfocusedStyle.Render("aaa"),
		footerStyle.Render("1..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		focusedStyle.Render("aaaa"),
		focusedStyle.Render("a") + unfocusedStyle.Render("aaa"),
		unfocusedStyle.Render("aa") + unfocusedStyle.Render("aa"),
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("a") + focusedStyle.Render("aaa"),
		focusedStyle.Render("aa"),
		unfocusedStyle.Render("aaaa"),
		footerStyle.Render("9..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		focusedStyle.Render("aaaa"),
		focusedStyle.Render("a") + unfocusedStyle.Render("aaa"),
		unfocusedStyle.Render("aa"),
		footerStyle.Render("5..."),
	})
	internal.CmpStr(t, expected, fv.View())

	// rollover
	fv, _ = fv.Update(prevMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		unfocusedStyle.Render("a") + unfocusedStyle.Render("aaa"),
		unfocusedStyle.Render("aa") + focusedStyle.Render("aa"),
		focusedStyle.Render("aaa"),
		footerStyle.Render("1..."),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[...",
		focusedStyle.Render("aaaa"),
		focusedStyle.Render("a") + unfocusedStyle.Render("aaa"),
		unfocusedStyle.Render("aa"),
		footerStyle.Render("5..."),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestMatchNavigationNoWrap(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		6,
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
		"",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] goose  (2/3 matches...",
		"...k duck duck duck duck " + unfocusedStyle.Render("goose"),
		focusedStyle.Render("...se") + " duck duck duck duck duck",
		"...ck duck duck duck duck duck",
		"",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedThirdMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] goose  (3/3 matches...",
		"duck duck duck duck duck du...",
		"duck duck duck duck duck " + unfocusedStyle.Render("go..."),
		focusedStyle.Render("goose") + " duck duck duck duck d...",
		"",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedThirdMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

func TestMatchNavigationNoWrapPanning(t *testing.T) {
	fv := makeFilterableViewport(
		10,
		3,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		strings.Repeat("a", 32),
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for range 4 {
		fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedLeftmostMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		focusedStyle.Render("aaaa") + unfocusedStyle.Render("aaa.") + unfocusedStyle.Render(".."),
		footerStyle.Render("100% (1/1)"),
	})

	internal.CmpStr(t, expectedLeftmostMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		unfocusedStyle.Render("aaaa") + focusedStyle.Render("aaa.") + unfocusedStyle.Render(".."),
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedTravelingRight := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		unfocusedStyle.Render("..") + unfocusedStyle.Render(".aaa") + focusedStyle.Render("a..."),
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedTravelingRight, fv.View())

	for range 4 {
		fv, _ = fv.Update(nextMatchKeyMsg)
		internal.CmpStr(t, expectedTravelingRight, fv.View())
	}

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedRightmostMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		unfocusedStyle.Render("..") + unfocusedStyle.Render(".aaa") + focusedStyle.Render("aaaa"),
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedRightmostMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		unfocusedStyle.Render("..") + focusedStyle.Render(".aaa") + unfocusedStyle.Render("aaaa"),
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	expectedTravelingLeft := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		focusedStyle.Render("...a") + unfocusedStyle.Render("aaa.") + unfocusedStyle.Render(".."),
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedTravelingLeft, fv.View())

	for range 4 {
		fv, _ = fv.Update(prevMatchKeyMsg)
		internal.CmpStr(t, expectedTravelingLeft, fv.View())
	}

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedLeftmostMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedRightmostMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedLeftmostMatch, fv.View())
}

func TestMatchNavigationNoWrapUnicode(t *testing.T) {
	fv := makeFilterableViewport(
		32,
		3,
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
		"💖💖💖💖💖💖💖💖 " + focusedStyle.Render("hi") + " aaaaaaaaa...",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

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

func TestMatchNavigationManyMatchesWrapPerformance(t *testing.T) {
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
	internal.RunWithTimeout(t, runTest, 200*time.Millisecond)
}

func TestScrollingWithManyHighlightedMatchesPerformance(t *testing.T) {
	runTest := func(t *testing.T) {
		width := 80
		height := 20
		fv := makeFilterableViewport(
			width,
			height,
			[]viewport.Option[object]{
				viewport.WithWrapText[object](false),
			},
			[]Option[object]{},
		)

		numItems := height * 5
		items := make([]string, numItems)
		for i := range items {
			items[i] = strings.Repeat("a", width)
		}
		fv.SetObjects(stringsToItems(items))

		// everything on screen highlighted
		fv, _ = fv.Update(filterKeyMsg)
		fv, _ = fv.Update(internal.MakeKeyMsg('a'))
		fv, _ = fv.Update(applyFilterKeyMsg)

		firstView := fv.View()
		if !strings.Contains(firstView, focusedStyle.Render("a")) {
			t.Fatal("expected focused match in initial view")
		}

		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		for i := 0; i < height; i++ {
			fv, _ = fv.Update(downMsg)
			view := fv.View()

			// after first scroll, focused match should go out of view
			// but unfocused matches should still be visible
			if i > 0 && strings.Contains(view, focusedStyle.Render("a")) {
				t.Errorf("focused match should be out of view after scrolling %d times", i+1)
			}
			if !strings.Contains(view, unfocusedStyle.Render("a")) {
				t.Errorf("unfocused matches should still be visible after scrolling %d times", i+1)
			}
		}
	}
	internal.RunWithTimeout(t, runTest, 220*time.Millisecond)
}

func TestScrollingWithManyHighlightedMatchesPerformanceSelectionEnabled(t *testing.T) {
	runTest := func(t *testing.T) {
		width := 80
		height := 20
		fv := makeFilterableViewport(
			width,
			height,
			[]viewport.Option[object]{
				viewport.WithWrapText[object](false),
				viewport.WithSelectionEnabled[object](true),
			},
			[]Option[object]{},
		)

		numItems := height * 5
		items := make([]string, numItems)
		for i := range items {
			items[i] = strings.Repeat("a", width)
		}
		fv.SetObjects(stringsToItems(items))

		// everything on screen highlighted
		fv, _ = fv.Update(filterKeyMsg)
		fv, _ = fv.Update(internal.MakeKeyMsg('a'))
		fv, _ = fv.Update(applyFilterKeyMsg)

		firstView := fv.View()
		if !strings.Contains(firstView, focusedStyle.Render("a")) {
			t.Fatal("expected focused match in initial view")
		}

		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		// with selection enabled, the viewport keeps the selected item (with focused match) in view
		// height - 2 accounts for header and footer lines, leaving content lines
		contentLines := height - 2
		for i := 0; i < height; i++ {
			fv, _ = fv.Update(downMsg)
			view := fv.View()

			// for first (contentLines - 1) scrolls, focused match stays in view
			// after that, selection scrolls past visible area
			if i < contentLines-1 {
				if !strings.Contains(view, focusedStyle.Render("a")) {
					t.Errorf("focused match should stay in view after moving selection down %d times", i+1)
				}
			} else {
				if strings.Contains(view, focusedStyle.Render("a")) {
					t.Errorf("focused match should be out of view after moving selection down %d times", i+1)
				}
			}

			// unfocused matches should always be visible
			if !strings.Contains(view, unfocusedStyle.Render("a")) {
				t.Errorf("unfocused matches should still be visible after moving selection down %d times", i+1)
			}
		}
	}
	internal.RunWithTimeout(t, runTest, 200*time.Millisecond)
}

func TestMatchNavigationWithSelectionEnabled(t *testing.T) {
	fv := makeFilterableViewport(
		40,
		5,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
			viewport.WithSelectionEnabled[object](true),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"apple pie",
		"banana bread",
		"apple cake",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "apple" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] apple  (1/2 matches on 2 items)",
		focusedStyle.Render("apple") + selectedItemStyle.Render(" pie"),
		"banana bread",
		unfocusedStyle.Render("apple") + " cake",
		footerStyle.Render("33% (1/3)"),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] apple  (2/2 matches on 2 items)",
		unfocusedStyle.Render("apple") + " pie",
		"banana bread",
		focusedStyle.Render("apple") + selectedItemStyle.Render(" cake"),
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

func TestMatchNavigationWithSelectionEnabledWrap(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		6,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
			viewport.WithSelectionEnabled[object](true),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"the quick brown fox",
		"jumped over the lazy dog",
		"the end",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "the" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] the  (1/3...",
		focusedStyle.Render("the") + selectedItemStyle.Render(" quick brown fox"),
		"jumped over " + unfocusedStyle.Render("the") + " lazy",
		" dog",
		unfocusedStyle.Render("the") + " end",
		footerStyle.Render("33% (1/3)"),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] the  (2/3...",
		unfocusedStyle.Render("the") + " quick brown fox",
		selectedItemStyle.Render("jumped over ") + focusedStyle.Render("the") + selectedItemStyle.Render(" lazy"),
		selectedItemStyle.Render(" dog"),
		unfocusedStyle.Render("the") + " end",
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedThirdMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] the  (3/3...",
		unfocusedStyle.Render("the") + " quick brown fox",
		"jumped over " + unfocusedStyle.Render("the") + " lazy",
		" dog",
		focusedStyle.Render("the") + selectedItemStyle.Render(" end"),
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedThirdMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedSecondMatch, fv.View())
}

func TestMatchNavigationWithSelectionEnabledWrapScrolling(t *testing.T) {
	fv := makeFilterableViewport(
		5,
		4,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
			viewport.WithSelectionEnabled[object](true),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"long long long long ",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "long " {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedTopFocused := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[e...",
		focusedStyle.Render("long "),
		unfocusedStyle.Render("long "),
		footerStyle.Render("10..."),
	})
	internal.CmpStr(t, expectedTopFocused, fv.View())

	for range 2 {
		fv, _ = fv.Update(nextMatchKeyMsg)
		expectedBottomFocused := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
			"[e...",
			unfocusedStyle.Render("long "),
			focusedStyle.Render("long "),
			footerStyle.Render("10..."),
		})
		internal.CmpStr(t, expectedBottomFocused, fv.View())
	}

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedTopFocused, fv.View())
}

func TestToggleWrap(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		6,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"the quick brown fox jumped over the lazy dog",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "lazy" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	// at first the match is in view
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] lazy  (1/...",
		"...ped over the " + focusedStyle.Render("l..."),
		"",
		"",
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// when we toggle wrapping here, the match happens to still be in view, but we don't force that
	// otherwise there would be surprising jumps if the user is scrolled away from the current match and toggles wrap
	fv.SetWrapText(true)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] lazy  (1/...",
		"the quick brown fox ",
		"jumped over the " + focusedStyle.Render("lazy"),
		" dog",
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// the match is out of view here, demonstrating the above comment
	fv.SetWrapText(false)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] lazy  (1/...",
		"the quick brown f...",
		"",
		"",
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestApplyFilterScrollsToFirstMatch(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"match here",
		"line 8",
	}))
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"line 1",
		"line 2",
		"line 3",
		footerStyle.Render("37% (3/8)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "match" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (1/1 matches...",
		"line 5",
		"line 6",
		focusedStyle.Render("match") + " here",
		footerStyle.Render("87% (7/8)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(cancelFilterKeyMsg)
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "lin" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] lin  (1/7 matches o...",
		focusedStyle.Render("lin") + "e 1",
		unfocusedStyle.Render("lin") + "e 2",
		unfocusedStyle.Render("lin") + "e 3",
		footerStyle.Render("37% (3/8)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] lin  (2/7 matches o...",
		unfocusedStyle.Render("lin") + "e 1",
		focusedStyle.Render("lin") + "e 2",
		unfocusedStyle.Render("lin") + "e 3",
		footerStyle.Render("37% (3/8)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('e'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] line  (1/7 matches ...",
		focusedStyle.Render("line") + " 1",
		unfocusedStyle.Render("line") + " 2",
		unfocusedStyle.Render("line") + " 3",
		footerStyle.Render("37% (3/8)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestSetObjectsPreservesMatchIndex(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"match one",
		"match two",
		"match three",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "match" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (2/3 matches...",
		unfocusedStyle.Render("match") + " one",
		focusedStyle.Render("match") + " two",
		unfocusedStyle.Render("match") + " three",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// add a new item - should stay on match 2, now 2/4
	fv.SetObjects(stringsToItems([]string{
		"match one",
		"match new",
		"match two",
		"match three",
	}))

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (2/4 matches...",
		unfocusedStyle.Render("match") + " one",
		focusedStyle.Render("match") + " new",
		unfocusedStyle.Render("match") + " two",
		footerStyle.Render("75% (3/4)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestAppendObjectsPreservesMatchIndex(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"match one",
		"match two",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "match" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	fv, _ = fv.Update(nextMatchKeyMsg)
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (2/2 matches...",
		unfocusedStyle.Render("match") + " one",
		focusedStyle.Render("match") + " two",
		"",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// append new items - should stay on match 2, now 2/4
	fv.AppendObjects(stringsToItems([]string{
		"match three",
		"match four",
	}))

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (2/4 matches...",
		unfocusedStyle.Render("match") + " one",
		focusedStyle.Render("match") + " two",
		unfocusedStyle.Render("match") + " three",
		footerStyle.Render("75% (3/4)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestAppendObjectsWithNil(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"item one",
		"item two",
	}))

	// appending nil should not crash or change objects
	fv.AppendObjects(nil)

	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"item one",
		"item two",
		"",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestAppendObjectsRespectsMatchLimit(t *testing.T) {
	fv := makeFilterableViewport(
		40,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithMaxMatchLimit[object](5),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"match one",
		"match two",
		"match three",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "match" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	// 3 matches, under limit
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (1/3 matches on 3 items)",
		focusedStyle.Render("match") + " one",
		unfocusedStyle.Render("match") + " two",
		unfocusedStyle.Render("match") + " three",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// append 3 more items, which will exceed the limit of 5
	fv.AppendObjects(stringsToItems([]string{
		"match four",
		"match five",
		"match six",
	}))

	// should now show limit exceeded message and all items
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (5+ matches on 6+ items)",
		"match one",
		"match two",
		"match three",
		footerStyle.Render("50% (3/6)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestAppendObjectsIncrementalWithMatchingItemsOnly(t *testing.T) {
	fv := makeFilterableViewport(
		40,
		6,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithMatchingItemsOnly[object](true),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"match one",
		"nothing here",
		"match two",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "match" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	// should show only matching items
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (1/2 matches on 2 item...",
		focusedStyle.Render("match") + " one",
		unfocusedStyle.Render("match") + " two",
		"",
		"",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// append mixed items (some matching, some not)
	fv.AppendObjects(stringsToItems([]string{
		"nothing",
		"match three",
		"also nothing",
		"match four",
	}))

	// should show only matching items, including new matches
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (1/4 matches on 4 item...",
		focusedStyle.Render("match") + " one",
		unfocusedStyle.Render("match") + " two",
		unfocusedStyle.Render("match") + " three",
		unfocusedStyle.Render("match") + " four",
		footerStyle.Render("100% (4/4)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestVerticalPadding(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		10,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
		},
		[]Option[object]{
			WithVerticalPad[object](2),
		},
	)

	// create many items so we can test padding
	items := make([]string, 50)
	for i := 0; i < 50; i++ {
		if i == 10 || i == 20 || i == 30 {
			items[i] = fmt.Sprintf("match item %d", i)
		} else {
			items[i] = fmt.Sprintf("item %d", i)
		}
	}
	fv.SetObjects(stringsToItems(items))

	// apply filter to find "match"
	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "match" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	// first match at item 10 should have at least 2 lines above and below
	// with 8 content lines and verticalPad=2, it shows items 5-12
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (1/3 matches...",
		"item 5",
		"item 6",
		"item 7",
		"item 8",
		"item 9",
		focusedStyle.Render("match") + " item 10",
		"item 11",
		"item 12",
		footerStyle.Render("26% (13/50)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	// navigate to second match at item 20
	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (2/3 matches...",
		"item 15",
		"item 16",
		"item 17",
		"item 18",
		"item 19",
		focusedStyle.Render("match") + " item 20",
		"item 21",
		"item 22",
		footerStyle.Render("46% (23/50)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestHorizontalPadding(t *testing.T) {
	fv := makeFilterableViewport(
		10,
		5,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
		},
		[]Option[object]{
			WithHorizontalPad[object](3),
		},
	)

	fv.SetObjects(stringsToItems([]string{
		"short goose text with some more words here",
		"another goose line with extra padding test",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "goose" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	// first match attempted padding of 3 on each side
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		".." + focusedStyle.Render(".oose") + "...",
		"... " + unfocusedStyle.Render("goo..") + ".",
		"",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	// second match attempted padding of 3 on each side
	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact]...",
		unfocusedStyle.Render("...se") + " t...",
		".." + focusedStyle.Render(".oose") + "...",
		"",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestMatchNavigationWithVerticalPadding(t *testing.T) {
	h := 34
	fv := makeFilterableViewport(
		100,
		h,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{
			WithVerticalPad[object](10),
		},
	)

	nItems := 50
	items := make([]string, nItems)
	for i := 0; i < nItems; i++ {
		items[i] = "hi"
	}
	fv.SetObjects(stringsToItems(items))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "hi" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedStrings := []string{
		"[exact] hi  (1/50 matches on 50 items)",
		focusedStyle.Render("hi"),
	}
	for i := 0; i < h-3; i++ { // -3 for header, focused line, & footer
		expectedStrings = append(expectedStrings, unfocusedStyle.Render("hi"))
	}
	expectedStrings = append(expectedStrings, footerStyle.Render("64% (32/50)"))
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), expectedStrings)
	internal.CmpStr(t, expectedView, fv.View())

	// go to bottom match, then previous match 21 times to reach the 10 padding above
	fv, _ = fv.Update(prevMatchKeyMsg)
	nPrev := 21
	for i := 0; i < nPrev; i++ {
		fv, _ = fv.Update(prevMatchKeyMsg)
	}
	expectedStrings = []string{"[exact] hi  (29/50 matches on 50 items)"}
	for i := 0; i < 10; i++ {
		expectedStrings = append(expectedStrings, unfocusedStyle.Render("hi"))
	}
	expectedStrings = append(expectedStrings, focusedStyle.Render("hi"))
	for i := 0; i < h-10-3; i++ {
		expectedStrings = append(expectedStrings, unfocusedStyle.Render("hi"))
	}
	expectedStrings = append(expectedStrings, footerStyle.Render("100% (50/50)"))
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), expectedStrings)
	internal.CmpStr(t, expectedView, fv.View())

	// next previous match should keep 10 lines above and scroll one up
	fv, _ = fv.Update(prevMatchKeyMsg)
	expectedStrings = []string{"[exact] hi  (28/50 matches on 50 items)"}
	for i := 0; i < 10; i++ {
		expectedStrings = append(expectedStrings, unfocusedStyle.Render("hi"))
	}
	expectedStrings = append(expectedStrings, focusedStyle.Render("hi"))
	for i := 0; i < h-10-3; i++ {
		expectedStrings = append(expectedStrings, unfocusedStyle.Render("hi"))
	}
	expectedStrings = append(expectedStrings, footerStyle.Render("98% (49/50)"))
	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), expectedStrings)
	internal.CmpStr(t, expectedView, fv.View())
}

func TestMatchNavigationRolloverWithVerticalPadding(t *testing.T) {
	fv := makeFilterableViewport(
		100,
		10,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](true),
		},
		[]Option[object]{
			WithVerticalPad[object](10),
		},
	)
	fv.SetSelectionEnabled(true)

	nItems := 20
	items := make([]string, nItems)
	for i := 0; i < nItems; i++ {
		items[i] = "hi"
	}
	fv.SetObjects(stringsToItems(items))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "hi" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] hi  (1/20 matches on 20 items)",
		focusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		footerStyle.Render("5% (1/20)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	// previous match (last one)
	fv, _ = fv.Update(prevMatchKeyMsg)
	expectedViewAfterScroll := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] hi  (20/20 matches on 20 items)",
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		unfocusedStyle.Render("hi"),
		focusedStyle.Render("hi"),
		footerStyle.Render("100% (20/20)"),
	})
	internal.CmpStr(t, expectedViewAfterScroll, fv.View())
}

func stringsToItems(vals []string) []object {
	items := make([]object, len(vals))
	for i, s := range vals {
		items[i] = object{item: item.NewItem(s)}
	}
	return items
}

func TestSelectionAndFocusedMatchAfterItemsChange(t *testing.T) {
	fv := makeFilterableViewport(
		100,
		5,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
			viewport.WithSelectionEnabled[object](true),
		},
		[]Option[object]{},
	)

	initialItems := []string{
		"1 2",
		"1 2",
		"1 2",
		"1 2",
		"1 2",
	}
	fv.SetObjects(stringsToItems(initialItems))

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('1'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	// focus second match
	fv, _ = fv.Update(nextMatchKeyMsg)

	// move selection to third item
	fv, _ = fv.Update(downKeyMsg)

	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] 1  (2/5 matches on 5 items)",
		unfocusedStyle.Render("1") + " 2",
		focusedStyle.Render("1") + " 2",
		unfocusedStyle.Render("1") + selectedItemStyle.Render(" 2"),
		footerStyle.Render("60% (3/5)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// add a new item
	initialItems = append(initialItems, "1 2")
	fv.SetObjects(stringsToItems(initialItems))

	// neither match nor selection should change
	expected = strings.ReplaceAll(expected, "2/5 matches on 5", "2/6 matches on 6")
	expected = strings.ReplaceAll(expected, "60% (3/5)", "50% (3/6)")
	internal.CmpStr(t, expected, fv.View())

	// changing match should change selection too
	fv, _ = fv.Update(nextMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] 1  (3/6 matches on 6 items)",
		unfocusedStyle.Render("1") + " 2",
		unfocusedStyle.Render("1") + " 2",
		focusedStyle.Render("1") + selectedItemStyle.Render(" 2"),
		footerStyle.Render("50% (3/6)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] 1  (2/6 matches on 6 items)",
		unfocusedStyle.Render("1") + " 2",
		focusedStyle.Render("1") + selectedItemStyle.Render(" 2"),
		unfocusedStyle.Render("1") + " 2",
		footerStyle.Render("33% (2/6)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestCurrentMatchNotCenteredAfterItemsChange(t *testing.T) {
	fv := makeFilterableViewport(
		100,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{},
	)

	initialItems := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
	}
	fv.SetObjects(stringsToItems(initialItems))

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('1'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] 1  (1/1 matches on 1 items)",
		focusedStyle.Render("1"),
		"2",
		footerStyle.Render("33% (2/6)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// scroll so focused match out of view
	fv, _ = fv.Update(downKeyMsg)
	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] 1  (1/1 matches on 1 items)",
		"2",
		"3",
		footerStyle.Render("50% (3/6)"),
	})
	internal.CmpStr(t, expected, fv.View())

	initialItems = append(initialItems, "7", "8", "9")
	fv.SetObjects(stringsToItems(initialItems))

	newExpected := strings.ReplaceAll(expected, "50% (3/6)", "33% (3/9)")
	internal.CmpStr(t, newExpected, fv.View())
}

func TestMaxMatchLimit(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		6,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithMaxMatchLimit[object](5),
			WithMatchingItemsOnly[object](true), // Should be ignored when limit exceeded
		},
	)

	items := []string{
		"apple apple",
		"apple apple",
		"apple apple",
		"apple apple",
		"apple apple",
		"banana",
	}
	fv.SetObjects(stringsToItems(items))

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "app" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: app  (5+ matches on 3+ items)",
		"apple apple",
		"apple apple",
		"apple apple",
		"apple apple",
		footerStyle.Render("66% (4/6)"),
	})
	internal.CmpStr(t, expectedView, fv.View())

	// view should be unchanged by navigating matches when limit exceeded
	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedView, fv.View())

	// clear search filter
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(cancelFilterKeyMsg)

	if fv.matchLimitExceeded {
		t.Error("matchLimitExceeded should be false after clearing filter")
	}

	// filter that doesn't exceed limit
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('b'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedView = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: b  (1/1 matches on 1 items) showing matches only",
		focusedStyle.Render("b") + "anana",
		"",
		"",
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestMaxMatchLimitWithAppendObjects(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		3,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithMaxMatchLimit[object](3),
		},
	)

	items := []string{
		"a",
		"bbb",
	}
	fv.SetObjects(stringsToItems(items))

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: a  (1/1 matches on 1 items)",
		focusedStyle.Render("a"),
		footerStyle.Render("50% (1/2)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// append new items that cause match limit to be exceeded
	fv.AppendObjects(stringsToItems([]string{"aaa", "aaa"}))

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: a  (3+ matches on 2+ items)",
		"a",
		footerStyle.Render("25% (1/4)"),
	})
	internal.CmpStr(t, expected, fv.View())
}

func TestMaxMatchLimitUnlimited(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		6,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithMaxMatchLimit[object](0), // unlimited
		},
	)

	fv.SetObjects(stringsToItems([]string{
		"apple apple",
	}))

	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('a'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] Filter: a  (1/2 matches on 1 items)",
		focusedStyle.Render("a") + "pple " + unfocusedStyle.Render("a") + "pple",
		"",
		"",
		"",
		footerStyle.Render("100% (1/1)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestToggleWrap_DoesNotJumpToMatchWhenScrolledAway(t *testing.T) {
	fv := makeFilterableViewport(
		30,
		5,
		[]viewport.Option[object]{
			viewport.WithWrapText[object](false),
			viewport.WithSelectionEnabled[object](true),
		},
		[]Option[object]{},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
		"line 4",
		"line 5",
		"line 6",
		"match here",
		"line 8",
	}))

	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		selectedItemStyle.Render("line 1"),
		"line 2",
		"line 3",
		footerStyle.Render("12% (1/8)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "match" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (1/1 matches...",
		"line 5",
		"line 6",
		focusedStyle.Render("match") + " here",
		footerStyle.Render("12% (1/8)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(internal.MakeKeyMsg('g'))

	expected = internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] match  (1/1 matches...",
		selectedItemStyle.Render("line 1"),
		"line 2",
		"line 3",
		footerStyle.Render("12% (1/8)"),
	})
	internal.CmpStr(t, expected, fv.View())

	// toggling wrap should not change view
	fv.SetWrapText(true)
	internal.CmpStr(t, expected, fv.View())
	fv.SetWrapText(false)
	internal.CmpStr(t, expected, fv.View())
}
