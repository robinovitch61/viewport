package filterableviewport

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/robinovitch61/bubbleo/internal"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
)

var (
	// TODO LEO: replace these with MakeKeyMsg from internal package
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
	//focusedStyle   = lipgloss.NewStyle()
	//unfocusedStyle = lipgloss.NewStyle()
	matchStyles = MatchStyles{
		Focused:   focusedStyle,
		Unfocused: unfocusedStyle,
	}
	filterableViewportStyles = Styles{
		CursorStyle: cursorStyle,
		Match:       matchStyles,
	}
)

// Note: this won't be necessary in future charm library versions
func init() {
	// Force TrueColor profile for consistent test output
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func makeFilterableViewport(
	width int,
	height int,
	vpOptions []viewport.Option[viewport.Item],
	fvOptions []Option[viewport.Item],
) *Model[viewport.Item] {
	// use default viewport test styles, will be overridden by options if passed in
	defaultTestVpStylesOption := viewport.WithStyles[viewport.Item](viewportStyles)
	vpOptions = append([]viewport.Option[viewport.Item]{defaultTestVpStylesOption}, vpOptions...)

	// use default filterable viewport test styles, will be overridden by options if passed in
	defaultTestFvStylesOption := WithStyles[viewport.Item](filterableViewportStyles)
	fvOptions = append([]Option[viewport.Item]{defaultTestFvStylesOption}, fvOptions...)

	vp := viewport.New[viewport.Item](width, height, vpOptions...)
	return New[viewport.Item](vp, fvOptions...)
}

func TestNew(t *testing.T) {
	// TODO LEO: use this in other tests
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
			WithEmptyText[viewport.Item]("No Filter"),
		},
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
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNewLongText(t *testing.T) {
	fv := makeFilterableViewport(
		10, // emptyText is longer than this
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
			WithEmptyText[viewport.Item]("No Filter Present"),
		},
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
		footerStyle.Render("66% (2/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestNewWidthHeight(t *testing.T) {
	fv := makeFilterableViewport(
		25,
		8,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	if fv.GetWidth() != 0 {
		t.Errorf("expected width 0 for negative input, got %d", fv.GetWidth())
	}
	if fv.GetHeight() != 0 {
		t.Errorf("expected height 0 for negative input, got %d", fv.GetHeight())
	}
	internal.CmpStr(t, "", fv.View())
}

func TestSetWidth(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetWidth(30)
	if fv.GetWidth() != 30 {
		t.Errorf("expected width 30, got %d", fv.GetWidth())
	}
}

func TestSetHeight(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetHeight(6)
	if fv.GetHeight() != 6 {
		t.Errorf("expected height 6, got %d", fv.GetHeight())
	}
}

func TestFilterFocused_Initial(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	if fv.FilterFocused() {
		t.Error("filter should not be focused initially")
	}
}

func TestEmptyContent(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
			WithEmptyText[viewport.Item]("No filter"),
		},
	)
	fv.SetContent([]viewport.Item{})
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No filter",
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestWithMatchesOnly_True(t *testing.T) {
	fv := makeFilterableViewport(
		80,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
			WithMatchingItemsOnly[viewport.Item](true),
		},
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
			WithMatchingItemsOnly[viewport.Item](false),
		},
	)
	fv.SetContent(stringsToItems([]string{
		"apple",
		"banana",
		"cherry",
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithCanToggleMatchingItemsOnly[viewport.Item](true),
		},
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithCanToggleMatchingItemsOnly[viewport.Item](false),
		},
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
			WithEmptyText[viewport.Item]("No Filter"),
		},
	)
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
	fv := makeFilterableViewport(
		40,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
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
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing filter key")
	}
}

func TestRegexFilterKeyFocus(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	if !fv.FilterFocused() {
		t.Error("filter should be focused after pressing regex filter key")
	}
}

func TestApplyFilterKey(t *testing.T) {
	fv := makeFilterableViewport(
		40,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeAKeyMsg)
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
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
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestRegexFilter_ValidPattern(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
		},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana", "apricot"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	fv, _ = fv.Update(typeAKeyMsg)
	fv, _ = fv.Update(typePKeyMsg)
	fv, _ = fv.Update(typePlusKeyMsg)
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
		},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(regexFilterKeyMsg)
	fv, _ = fv.Update(typeLeftBracketKeyMsg)
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeXKeyMsg)
	fv, _ = fv.Update(typeYKeyMsg)
	fv, _ = fv.Update(typeZKeyMsg)
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithKeyMap[viewport.Item](customKeyMap),
		},
	)
	fv.SetContent(stringsToItems([]string{"test"}))
	fv, _ = fv.Update(filterKeyMsg) // should not match custom key
	if fv.FilterFocused() {
		t.Error("filter should not be focused with custom keymap")
	}
}

func TestViewportControls(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		3,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetContent(stringsToItems([]string{"line1", "line2", "line3"}))
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithEmptyText[viewport.Item]("No filter applied"),
		},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithPrefixText[viewport.Item]("Filter:"),
		},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
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
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{
			WithCanToggleMatchingItemsOnly[viewport.Item](true),
		},
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
		"[exact] ponN/r" + cursorStyle.Render(" ") + " (no matches)",
		"apple",
		"book",
		footerStyle.Render("50% (2/4)"),
	})
	internal.CmpStr(t, expectedViewAfterO, fv.View())
}

func TestMatchNavigationWithNoMatches(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		4,
		[]viewport.Option[viewport.Item]{},
		[]Option[viewport.Item]{},
	)
	fv.SetContent(stringsToItems([]string{"apple", "banana"}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeXKeyMsg)
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

func TestMatchNavigationWithAllItemsWrapped(t *testing.T) {
	fv := makeFilterableViewport(
		60,
		5,
		[]viewport.Option[viewport.Item]{
			viewport.WithWrapText[viewport.Item](true),
		},
		[]Option[viewport.Item]{
			WithStyles[viewport.Item](Styles{
				Match: matchStyles,
			}),
			WithMatchingItemsOnly[viewport.Item](false),
		},
	)
	fv.SetContent(stringsToItems([]string{
		"hi there",
		"hi over there",
		"no match",
	}))
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"hi there",
		"hi over there",
		"no match",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "there" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] there  (1/2 matches on 2 items)",
		"hi " + focusedStyle.Render("there"),
		"hi over " + unfocusedStyle.Render("there"),
		"no match",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] there  (2/2 matches on 2 items)",
		"hi " + unfocusedStyle.Render("there"),
		"hi over " + focusedStyle.Render("there"),
		"no match",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

func TestMatchNavigationWithMatchingItemsOnlyWrapped(t *testing.T) {
	fv := makeFilterableViewport(
		60,
		5,
		[]viewport.Option[viewport.Item]{
			viewport.WithWrapText[viewport.Item](true),
		},
		[]Option[viewport.Item]{
			WithStyles[viewport.Item](Styles{
				Match: matchStyles,
			}),
			WithMatchingItemsOnly[viewport.Item](true),
			WithCanToggleMatchingItemsOnly[viewport.Item](false),
		},
	)
	fv.SetContent(stringsToItems([]string{
		"hi there",
		"hi over there",
		"no match",
	}))
	expected := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"No Filter",
		"hi there",
		"hi over there",
		"no match",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expected, fv.View())

	fv, _ = fv.Update(filterKeyMsg)
	for _, c := range "there" {
		fv, _ = fv.Update(internal.MakeKeyMsg(c))
	}
	fv, _ = fv.Update(applyFilterKeyMsg)
	expectedFirstMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] there  (1/2 matches on 2 items) showing matches only",
		"hi " + focusedStyle.Render("there"),
		"hi over " + unfocusedStyle.Render("there"),
	})
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	expectedSecondMatch := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"[exact] there  (2/2 matches on 2 items) showing matches only",
		"hi " + unfocusedStyle.Render("there"),
		"hi over " + focusedStyle.Render("there"),
	})
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(nextMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedSecondMatch, fv.View())

	fv, _ = fv.Update(prevMatchKeyMsg)
	internal.CmpStr(t, expectedFirstMatch, fv.View())
}

//func TestMatchNavigationNoWrap(t *testing.T) {
//	// TODO LEO
//}
//
//func TestMatchNavigationSelectionEnabled(t *testing.T) {
//	// TODO LEO
//}

func TestMatchNavigationManyMatchesWrapped(t *testing.T) {
	fv := makeFilterableViewport(
		100,
		50,
		[]viewport.Option[viewport.Item]{
			viewport.WithWrapText[viewport.Item](true),
		},
		[]Option[viewport.Item]{},
	)
	numAs := 10000
	fv.SetContent(stringsToItems([]string{
		strings.Repeat("a", numAs),
	}))
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(typeAKeyMsg)
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

// TODO LEO: add test for 10k character 'a' in a single wrapped line and filter is 'a' - very slow right now

// TODO LEO: add tests for match navigation with matches

// TODO LEO: add test for match navigation showing only matches

// TODO LEO: add test for when wrapped item goes off screen and focused match in the item is off screen (currently shows top lines item and not focused match)

// TODO LEO: add test that updating filter itself scrolls/pans screen to first match without needing to press n/N

// TODO LEO: test for multiple regex matches in a single line

func stringsToItems(vals []string) []viewport.Item {
	items := make([]viewport.Item, len(vals))
	for i, s := range vals {
		items[i] = viewport.Item{LineBuffer: linebuffer.New(s)}
	}
	return items
}
