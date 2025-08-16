package viewport

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/muesli/termenv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport/internal"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
)

// Note: this won't be necessary in future charm library versions
func init() {
	// Force TrueColor profile for consistent test output
	lipgloss.SetColorProfile(termenv.TrueColor)
}

var (
	downKeyMsg       = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	halfPgDownKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	fullPgDownKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	upKeyMsg         = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	halfPgUpKeyMsg   = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	fullPgUpKeyMsg   = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}}
	goToTopKeyMsg    = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	goToBottomKeyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	red              = lipgloss.Color("#ff0000")
	blue             = lipgloss.Color("#0000ff")
	green            = lipgloss.Color("#00ff00")
	selectionStyle   = lipgloss.NewStyle().Foreground(blue)
)

func newViewport(width, height int) *Model[Item] {
	styles := Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	}

	return New[Item](width, height, DefaultKeyMap(), styles)
}

// # SELECTION DISABLED, WRAP OFF

func TestViewport_SelectionOff_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"."})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"..", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"...", "hi", "..."})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really rea...",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first",
		"second",
	})
	if selectionEnabled := vp.GetSelectionEnabled(); selectionEnabled {
		t.Errorf("expected selection to be disabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); wrapText {
		t.Errorf("expected text wrapping to be disabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem != nil {
		t.Errorf("expected selected item to be nil, got %v", selectedItem)
	}
}

func TestViewport_SelectionOff_WrapOff_ShowFooter(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really rea...",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really rea...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle().Foreground(red),
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"1",
		"2",
		"3",
		"\x1b[38;2;255;0;0m75% (3/4)" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"    first li...",
		"          fi...",
		"            ...",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	setContent(vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"long header ...",
		"123456789012345",
		"123456789012...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"123456789012345",
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	doSetContent := func() {
		setContent(vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOff_WrapOff_ScrollToItem(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so last item in view
	vp.ScrollSoItemIdxInView(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so second item in view
	vp.ScrollSoItemIdxInView(1)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_BulkScrolling(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Panning(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	doSetContent := func() {
		setContent(vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long",
			"fourth",
			"fifth line that is fairly long",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"first l...",
		"second ...",
		"third l...",
		"fourth",
		"66% (4/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ne t...",
		"...ine ...",
		"...ne t...",
		".",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ine ...",
		"...ne t...",
		".",
		"...ne t...",
		"83% (5/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...e first",
		"...",
		"...",
		"...",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		"...",
		"...ly long",
		"...",
		"100% (6/6)",
	})
	validate(expectedView)

	// set shorter LineBuffer
	setContent(vp, []string{
		"the first one",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...rst one",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove LineBuffer
	setContent(vp, []string{})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// re-add LineBuffer
	setContent(vp, []string{
		"first",
		"second",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetStringToHighlight("second")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(red),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST,
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_RegexToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetRegexToHighlight(regexp.MustCompile("s.*n"))
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(red),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecon" + linebuffer.RST + "d",
		"\x1b[38;2;255;0;0msecon" + linebuffer.RST + "d",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 7) + strings.Repeat(".", 3),
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 20*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOff_RegexToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetRegexToHighlight(regexp.MustCompile("r"))
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 7) + strings.Repeat(".", 3),
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 20*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOff_StringToHighlightAnsi(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e again",
	})
	vp.SetStringToHighlight("e")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           selectionStyle,
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"lin\x1b[38;2;0;0;255me" + linebuffer.RST + " \x1b[38;2;255;0;0mr" + linebuffer.RST + "\x1b[38;2;0;0;255me" + linebuffer.RST + "\x1b[38;2;255;0;0md" + linebuffer.RST + " \x1b[38;2;0;0;255me" + linebuffer.RST + " again",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetStringToHighlight("38")
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e again",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_RegexToHighlightAnsi(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e again",
	})
	vp.SetRegexToHighlight(regexp.MustCompile("r+e"))
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           selectionStyle,
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line \x1b[38;2;0;0;255mre" + linebuffer.RST + "\x1b[38;2;255;0;0md" + linebuffer.RST + " e again",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetRegexToHighlight(regexp.MustCompile("38"))
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e again",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	setContent(vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           selectionStyle,
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		"A💖\x1b[38;2;0;0;255m中é" + linebuffer.RST,
		"A💖\x1b[38;2;0;0;255m中é" + linebuffer.RST + "...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SetSelectionEnabled_SetsTopVisibleItem(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp.SetSelectionEnabled(true)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"fourth",
		"fifth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP OFF

func TestViewport_SelectionOn_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"."})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"..", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"...",
		"\x1b[38;2;0;0;255mhi" + linebuffer.RST,
		"...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really rea...",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first",
		"second",
	})
	if selectionEnabled := vp.GetSelectionEnabled(); !selectionEnabled {
		t.Errorf("expected selection to be enabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); wrapText {
		t.Errorf("expected text wrapping to be disabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	vp, _ = vp.Update(downKeyMsg)
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 1 {
		t.Errorf("expected selected item index to be 1, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem.Render().Content() != "second" {
		t.Errorf("got unexpected selected item: %v", selectedItem)
	}
}

func TestViewport_SelectionOn_WrapOff_ShowFooter(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really rea...",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really rea..." + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really rea...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle().Foreground(red),
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m1" + linebuffer.RST,
		"2",
		"3",
		"\x1b[38;2;255;0;0m25% (1/4)" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"second line",
		"third line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"second line",
		"third line",
		"fourth line",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m    first li..." + linebuffer.RST,
		"          fi...",
		"            ...",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline1" + linebuffer.RST,
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline2" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_OverflowLine(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"long header ...",
		"\x1b[38;2;0;0;255m123456789012345" + linebuffer.RST,
		"123456789012...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m123456789012345" + linebuffer.RST,
		"123456789012...",
		"123456789012...",
		"123456789012...",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"third",
		"fourth",
		"16% (1/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"third",
		"fourth",
		"33% (2/6)",
	})
	validate(expectedView)

	// scrolling to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOff_ScrollToItem(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// attempting to scroll so selection out of view is no-op
	vp.ScrollSoItemIdxInView(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so third item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"third",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_BulkScrolling(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"fourth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfourth" + linebuffer.RST,
		"fifth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"\x1b[38;2;0;0;255mfourth" + linebuffer.RST,
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_Panning(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long",
			"fourth",
			"fifth line that is fairly long",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"\x1b[38;2;0;0;255mfirst l..." + linebuffer.RST,
		"second ...",
		"third l...",
		"fourth",
		"16% (1/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"\x1b[38;2;0;0;255m...ne t..." + linebuffer.RST,
		"...ine ...",
		"...ne t...",
		".",
		"16% (1/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ne t...",
		"\x1b[38;2;0;0;255m...ine ..." + linebuffer.RST,
		"...ne t...",
		".",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...",
		"\x1b[38;2;0;0;255m...e first" + linebuffer.RST,
		"...",
		"...",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...",
		"...e first",
		"\x1b[38;2;0;0;255m..." + linebuffer.RST,
		"...",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...",
		"...e first",
		"...",
		"\x1b[38;2;0;0;255m..." + linebuffer.RST,
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...e first",
		"...",
		"...",
		"\x1b[38;2;0;0;255m..." + linebuffer.RST,
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		"...",
		"...ly long",
		"\x1b[38;2;0;0;255m..." + linebuffer.RST,
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		"...",
		"\x1b[38;2;0;0;255m...ly long" + linebuffer.RST,
		"...",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		"\x1b[38;2;0;0;255m..." + linebuffer.RST,
		"...ly long",
		"...",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"\x1b[38;2;0;0;255m...ly long" + linebuffer.RST,
		"...",
		"...ly long",
		"...",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"\x1b[38;2;0;0;255m...n mu..." + linebuffer.RST,
		"...ly long",
		"...",
		"...ly long",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"\x1b[38;2;0;0;255m...ly long" + linebuffer.RST,
		"...n mu...",
		"...ly long",
		"...",
		"16% (1/6)",
	})
	validate(expectedView)

	// set shorter LineBuffer
	setContent(vp, []string{
		"the first one",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"\x1b[38;2;0;0;255m...rst one" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_MaintainSelection(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetSelectionComparator(ItemCompareFn)
	setContent(vp, []string{
		"sixth",
		"seventh",
		"eighth",
		"ninth",
		"tenth",
		"eleventh",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"seventh",
		"eighth",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh" + linebuffer.RST,
		"eighth",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer above
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"seventh",
		"eighth",
		"ninth",
		"tenth",
		"eleventh",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh" + linebuffer.RST,
		"eighth",
		"63% (7/11)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer below
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"seventh",
		"eighth",
		"ninth",
		"tenth",
		"eleventh",
		"twelfth",
		"thirteenth",
		"fourteenth",
		"fifteenth",
		"sixteenth",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth",
		"\x1b[38;2;0;0;255mseventh" + linebuffer.RST,
		"eighth",
		"43% (7/16)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyTop(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetTopSticky(true)
	setContent(vp, []string{
		"first",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"first",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyBottom(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"first",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"first",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"first",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyBottomOverflowHeight(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetBottomSticky(true)

	// test covers case where first set LineBuffer to empty, then overflow height
	setContent(vp, []string{})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StickyTopBottom(t *testing.T) {
	w, h := 15, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetTopSticky(true)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"first",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer, top sticky wins out arbitrarily when both set
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"first",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"third",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"second",
		"first",
		"third",
		"fourth",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"third",
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_RemoveLogsWhenSelectionBottom(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)

	// add LineBuffer
	setContent(vp, []string{
		"second",
		"first",
		"third",
		"fourth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfourth" + linebuffer.RST,
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove LineBuffer
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"fourth",
		"fifth",
		"sixth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"fourth",
		"fifth",
		"sixth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"third",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove LineBuffer
	setContent(vp, []string{
		"second",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove all LineBuffer
	setContent(vp, []string{})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer (maintain selection off)
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"third",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StringToHighlight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetStringToHighlight("second")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(green),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the second line",
		"the fourth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first line" + linebuffer.RST,
		"the \x1b[38;2;0;255;0msecond" + linebuffer.RST + " line",
		"the \x1b[38;2;0;255;0msecond" + linebuffer.RST + " line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetStringToHighlight("first")
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe " + linebuffer.RST + "\x1b[38;2;255;0;0mfirst" + linebuffer.RST + "\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"the second line",
		"the second line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"first line",
		"second line",
		"second line",
		"fourth line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;255;0;0mfirst" + linebuffer.RST + "\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"second line",
		"second line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_RegexToHighlight(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetRegexToHighlight(regexp.MustCompile("s+.*nd"))
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(green),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the second line",
		"the fourth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first line" + linebuffer.RST,
		"the \x1b[38;2;0;255;0msecond" + linebuffer.RST + " line",
		"the \x1b[38;2;0;255;0msecond" + linebuffer.RST + " line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetRegexToHighlight(regexp.MustCompile("fir.t"))
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe " + linebuffer.RST + "\x1b[38;2;255;0;0mfirst" + linebuffer.RST + "\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"the second line",
		"the second line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"first line",
		"second line",
		"second line",
		"fourth line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;255;0;0mfirst" + linebuffer.RST + "\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"second line",
		"second line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 7) + "\x1b[38;2;0;0;255m" + strings.Repeat(".", 3) + linebuffer.RST,
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOff_RegexToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetRegexToHighlight(regexp.MustCompile("r"))
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 7) + "\x1b[38;2;0;0;255m" + strings.Repeat(".", 3) + linebuffer.RST,
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOff_AnsiOnSelection(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"line with \x1b[38;2;255;0;0mred" + linebuffer.RST + " text",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mline with " + linebuffer.RST + "\x1b[38;2;255;0;0mred" + linebuffer.RST + "\x1b[38;2;0;0;255m text" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SelectionEmpty(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m " + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ExtraSlash(t *testing.T) {
	w, h := 25, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"|2024|\x1b[38;2;0mfl..lq" + linebuffer.RST + "/\x1b[38;2;0mflask-3" + linebuffer.RST + "|",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m|2024|" + linebuffer.RST + "\x1b[38;2;0mfl..lq" + linebuffer.RST + "\x1b[38;2;0;0;255m/" + linebuffer.RST + "\x1b[38;2;0mflask-3" + linebuffer.RST + "\x1b[38;2;0;0;255m|" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(green),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		"\x1b[38;2;0;0;255mA💖" + linebuffer.RST + "\x1b[38;2;255;0;0m中é" + linebuffer.RST,
		"A💖\x1b[38;2;0;255;0m中é" + linebuffer.RST + "...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # SELECTION DISABLED, WRAP ON

func TestViewport_SelectionOff_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"h"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"he", "ad"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"hea", "der", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"head", "er", "hi", "1..."})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first",
		"second",
	})
	if selectionEnabled := vp.GetSelectionEnabled(); selectionEnabled {
		t.Errorf("expected selection to be disabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); !wrapText {
		t.Errorf("expected text wrapping to be enabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem != nil {
		t.Errorf("expected selected item to be nil, got %v", selectedItem)
	}
}

func TestViewport_SelectionOff_WrapOn_ShowFooter(t *testing.T) {
	w, h := 15, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really really",
		"99% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really really",
		" long line",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really really",
		" long line",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle().Foreground(red),
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"1",
		"2",
		"3",
		"\x1b[38;2;255;0;0m75% (3/4)" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	// trailing space is not trimmed
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"    first line ",
		"",
		"          first",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"long header ove",
		"rflows",
		"123456789012345",
		"123456789012345",
		"6",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"123456789012345",
		"123456789012345",
		"6",
		"123456789012345",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	doSetContent := func() {
		setContent(vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOff_WrapOn_ScrollToItem(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so last item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so second item in view
	vp.ScrollSoItemIdxInView(1)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_BulkScrolling(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		" line",
		"the third ",
		"99% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		"the second",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Panning(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetWrapText(true)
	doSetContent := func() {
		setContent(vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long",
			"fourth",
			"fifth line that is fairly long",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"first line",
		" that is f",
		"airly long",
		"second lin",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		" that is f",
		"airly long",
		"second lin",
		"e that is ",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"airly long",
		"second lin",
		"e that is",
		"even much",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"second lin",
		"e that is ",
		"even much ",
		"longer tha",
		"33% (2/6)",
	})
	validate(expectedView)
}

func TestViewport_SelectionOff_WrapOn_ChangeHeight(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third",
		"99% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"the second",
		" line",
		"the third",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_ChangeContent(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove LineBuffer
	setContent(vp, []string{
		"the first line",
		"the second line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove all LineBuffer
	setContent(vp, []string{})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetStringToHighlight("second")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(red),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST,
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"averylongwordthatwraps",
	})
	vp.SetStringToHighlight("wraps")
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"averylongw",
		"ordthat\x1b[38;2;255;0;0mwra" + linebuffer.RST,
		"\x1b[38;2;255;0;0mps" + linebuffer.RST,
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_RegexToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetRegexToHighlight(regexp.MustCompile("s...nd"))
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(red),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST,
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetWrapText(true)
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 10),
			"99% (1/1)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOn_RegexToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetWrapText(true)
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetRegexToHighlight(regexp.MustCompile("r"))
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;0;255;0mr"+linebuffer.RST+"", 10),
			"99% (1/1)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOn_RegexToHighlightMissesWrap(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetWrapText(true)
		setContent(vp, []string{
			"this is too long and triggers wrapping",
		})
		vp.SetRegexToHighlight(regexp.MustCompile("this.*too"))
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		// regex matches aren't shown for wrapped lines for performance reasons (lines could be extremely long)
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"this is to",
			"o long and",
			" triggers ",
			"99% (1/1)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp.SetRegexToHighlight(regexp.MustCompile("this.*to"))
		expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"\x1b[38;2;0;255;0mthis is to" + linebuffer.RST,
			"o long and",
			" triggers ",
			"99% (1/1)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOn_StringToHighlightAnsi(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e again",
	})
	vp.SetStringToHighlight("e")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           selectionStyle,
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"lin\x1b[38;2;0;0;255me" + linebuffer.RST + " \x1b[38;2;255;0;0mr" + linebuffer.RST + "\x1b[38;2;0;0;255me" + linebuffer.RST + "\x1b[38;2;255;0;0md" + linebuffer.RST + " \x1b[38;2;0;0;255me" + linebuffer.RST,
		" again",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetStringToHighlight("38")
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e",
		" again",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_RegexToHighlightAnsi(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e again",
	})
	vp.SetRegexToHighlight(regexp.MustCompile("e"))
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           selectionStyle,
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"lin\x1b[38;2;0;0;255me" + linebuffer.RST + " \x1b[38;2;255;0;0mr" + linebuffer.RST + "\x1b[38;2;0;0;255me" + linebuffer.RST + "\x1b[38;2;255;0;0md" + linebuffer.RST + " \x1b[38;2;0;0;255me" + linebuffer.RST,
		" again",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// should not highlight the ansi escape codes themselves
	vp.SetRegexToHighlight(regexp.MustCompile("38"))
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line \x1b[38;2;255;0;0mred" + linebuffer.RST + " e",
		" again",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SuperLongWrappedLine(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetWrapText(true)
		setContent(vp, []string{
			"smol",
			strings.Repeat("12345678", 1000000),
			"smol",
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"smol",
			"1234567812",
			"3456781234",
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"1234567812",
			"3456781234",
			"5678123456",
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"3456781234",
			"5678123456",
			"7812345678",
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(goToBottomKeyMsg)
		expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"5678123456",
			"7812345678",
			"smol",
			"100% (3/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 500*time.Millisecond)
}

func TestViewport_SelectionOff_WrapOn_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(green),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		"A💖\x1b[38;2;0;255;0m中é" + linebuffer.RST,
		"A💖\x1b[38;2;0;255;0m中é" + linebuffer.RST + "A💖",
		"\x1b[38;2;0;255;0m中é" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_EnableSelectionShowsTopLineInItem(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"short",
		"this is a very long line",
		"another short line",
		"final line",
	})
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"very long ",
		"line",
		"another sh",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetSelectionEnabled(true)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"\x1b[38;2;0;0;255mthis is a " + linebuffer.RST,
		"\x1b[38;2;0;0;255mvery long " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP ON

func TestViewport_SelectionOn_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"h"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"he", "ad"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{"hea", "der", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"head",
		"er",
		"\x1b[38;2;0;0;255mhi" + linebuffer.RST,
		"1...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_GetConfigs(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first",
		"second",
	})
	if selectionEnabled := vp.GetSelectionEnabled(); !selectionEnabled {
		t.Errorf("expected selection to be enabled, got %v", selectionEnabled)
	}
	if wrapText := vp.GetWrapText(); !wrapText {
		t.Errorf("expected text wrapping to be enabled, got %v", wrapText)
	}
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 0 {
		t.Errorf("expected selected item index to be 0, got %v", selectedItemIdx)
	}
	vp, _ = vp.Update(downKeyMsg)
	if selectedItemIdx := vp.GetSelectedItemIdx(); selectedItemIdx != 1 {
		t.Errorf("expected selected item index to be 1, got %v", selectedItemIdx)
	}
	if selectedItem := vp.GetSelectedItem(); selectedItem.Render().Content() != "second" {
		t.Errorf("got unexpected selected item: %v", selectedItem)
	}
}

func TestViewport_SelectionOn_WrapOn_ShowFooter(t *testing.T) {
	w, h := 15, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first line",
		lipgloss.NewStyle().Foreground(red).Render("second") + " line",
		lipgloss.NewStyle().Foreground(red).Render("a really really long line"),
		lipgloss.NewStyle().Foreground(red).Render("a") + " really really long line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really really",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really really",
		" long line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;255;0;0msecond" + linebuffer.RST + " line",
		"\x1b[38;2;255;0;0ma really really" + linebuffer.RST,
		"\x1b[38;2;255;0;0m long line" + linebuffer.RST,
		"\x1b[38;2;255;0;0ma" + linebuffer.RST + " really really",
		" long line",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle().Foreground(red),
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m1" + linebuffer.RST,
		"2",
		"3",
		"\x1b[38;2;255;0;0m25% (1/4)" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_FooterDisabled(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first line",
		"second line",
		"third line",
		"fourth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"second line",
		"third line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"second line",
		"third line",
		"fourth line",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SpaceAround(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"    first line     ",
		"          first line          ",
		"               first line               ",
	})
	// trailing space is not trimmed
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m    first line " + linebuffer.RST,
		"\x1b[38;2;0;0;255m    " + linebuffer.RST,
		"          first",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_MultiHeader(t *testing.T) {
	w, h := 15, 2
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header1", "header2"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"line1",
		"line2",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline1" + linebuffer.RST,
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"\x1b[38;2;0;0;255mline2" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"\x1b[38;2;0;0;255mline2" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_OverflowLine(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"long header overflows"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"long header ove",
		"rflows",
		"\x1b[38;2;0;0;255m123456789012345" + linebuffer.RST,
		"123456789012345",
		"6",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_OverflowHeight(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"123456789012345",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
		"1234567890123456",
	})
	vp.SetSelectedItemIdx(1)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"123456789012345",
		"\x1b[38;2;0;0;255m123456789012345" + linebuffer.RST,
		"\x1b[38;2;0;0;255m6" + linebuffer.RST,
		"123456789012345",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Scrolling(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(vp, []string{
			"first",
			"second",
			"third",
			"fourth",
			"fifth",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"second",
		"third",
		"fourth",
		"16% (1/6)",
	})
	validate(expectedView)

	// scrolling up past top is no-op
	vp, _ = vp.Update(upKeyMsg)
	validate(expectedView)

	// scrolling down by one
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"\x1b[38;2;0;0;255msecond" + linebuffer.RST,
		"third",
		"fourth",
		"33% (2/6)",
	})
	validate(expectedView)

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"\x1b[38;2;0;0;255mthird" + linebuffer.RST,
		"fourth",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOn_ScrollToItem(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"the second",
		" line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// attempting to scroll so selection out of view is no-op
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"the second",
		" line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"\x1b[38;2;0;0;255mthe second" + linebuffer.RST,
		"\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so third item in view
	vp.ScrollSoItemIdxInView(2)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe second" + linebuffer.RST,
		"\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"the third",
		"line",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_BulkScrolling(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe second" + linebuffer.RST,
		"\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe second" + linebuffer.RST,
		"\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_Panning(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	doSetContent := func() {
		setContent(vp, []string{
			"first line that is fairly long",
			"second line that is even much longer than the first",
			"third line that is fairly long as well",
			"fourth kinda long",
			"fifth kinda long too",
			"sixth",
		})
	}
	validate := func(expectedView string) {
		// set LineBuffer multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;0;0;255m that is f" + linebuffer.RST,
		"\x1b[38;2;0;0;255mairly long" + linebuffer.RST,
		"second lin",
		"16% (1/6)",
	})
	validate(expectedView)

	// pan right
	vp.safelySetXOffset(5)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255msecond lin" + linebuffer.RST,
		"\x1b[38;2;0;0;255me that is " + linebuffer.RST,
		"\x1b[38;2;0;0;255meven much " + linebuffer.RST,
		"\x1b[38;2;0;0;255mlonger tha" + linebuffer.RST,
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.safelySetXOffset(41)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mthird line" + linebuffer.RST,
		"\x1b[38;2;0;0;255m that is f" + linebuffer.RST,
		"\x1b[38;2;0;0;255mairly long" + linebuffer.RST,
		"\x1b[38;2;0;0;255m as well" + linebuffer.RST,
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"airly long",
		" as well",
		"\x1b[38;2;0;0;255mfourth kin" + linebuffer.RST,
		"\x1b[38;2;0;0;255mda long" + linebuffer.RST,
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"fourth kin",
		"da long",
		"\x1b[38;2;0;0;255mfifth kind" + linebuffer.RST,
		"\x1b[38;2;0;0;255ma long too" + linebuffer.RST,
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"da long",
		"fifth kind",
		"a long too",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"da long",
		"\x1b[38;2;0;0;255mfifth kind" + linebuffer.RST,
		"\x1b[38;2;0;0;255ma long too" + linebuffer.RST,
		"sixth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfourth kin" + linebuffer.RST,
		"\x1b[38;2;0;0;255mda long" + linebuffer.RST,
		"fifth kind",
		"a long too",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mthird line" + linebuffer.RST,
		"\x1b[38;2;0;0;255m that is f" + linebuffer.RST,
		"\x1b[38;2;0;0;255mairly long" + linebuffer.RST,
		"\x1b[38;2;0;0;255m as well" + linebuffer.RST,
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255msecond lin" + linebuffer.RST,
		"\x1b[38;2;0;0;255me that is " + linebuffer.RST,
		"\x1b[38;2;0;0;255meven much " + linebuffer.RST,
		"\x1b[38;2;0;0;255mlonger tha" + linebuffer.RST,
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"\x1b[38;2;0;0;255mfirst line" + linebuffer.RST,
		"\x1b[38;2;0;0;255m that is f" + linebuffer.RST,
		"\x1b[38;2;0;0;255mairly long" + linebuffer.RST,
		"second lin",
		"16% (1/6)",
	})
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOn_MaintainSelection(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetSelectionComparator(ItemCompareFn)
	setContent(vp, []string{
		"sixth item",
		"seventh item",
		"eighth item",
		"ninth item",
		"tenth item",
		"eleventh item",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255msixth item" + linebuffer.RST,
		"seventh it",
		"em",
		"eighth ite",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it" + linebuffer.RST,
		"\x1b[38;2;0;0;255mem" + linebuffer.RST,
		"eighth ite",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer above
	setContent(vp, []string{
		"first item",
		"second item",
		"third item",
		"fourth item",
		"fifth item",
		"sixth item",
		"seventh item",
		"eighth item",
		"ninth item",
		"tenth item",
		"eleventh item",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it" + linebuffer.RST,
		"\x1b[38;2;0;0;255mem" + linebuffer.RST,
		"eighth ite",
		"63% (7/11)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer below
	setContent(vp, []string{
		"first item",
		"second item",
		"third item",
		"fourth item",
		"fifth item",
		"sixth item",
		"seventh item",
		"eighth item",
		"ninth item",
		"tenth item",
		"eleventh item",
		"twelfth item",
		"thirteenth item",
		"fourteenth item",
		"fifteenth item",
		"sixteenth item",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth item",
		"\x1b[38;2;0;0;255mseventh it" + linebuffer.RST,
		"\x1b[38;2;0;0;255mem" + linebuffer.RST,
		"eighth ite",
		"43% (7/16)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyTop(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetTopSticky(true)
	setContent(vp, []string{
		"the first line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe second" + linebuffer.RST,
		"\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyBottom(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"the first line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add longer LineBuffer at bottom
	setContent(vp, []string{
		"the second line",
		"the first line",
		"a very long line that wraps a lot",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255ma very lon" + linebuffer.RST,
		"\x1b[38;2;0;0;255mg line tha" + linebuffer.RST,
		"\x1b[38;2;0;0;255mt wraps a " + linebuffer.RST,
		"\x1b[38;2;0;0;255mlot" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"a very lon",
		"g line tha",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
		"a very long line that wraps a lot",
		"the third line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"a very lon",
		"g line tha",
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyBottomOverflowHeight(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetBottomSticky(true)

	// test covers case where first set LineBuffer to empty, then overflow height
	setContent(vp, []string{})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyTopBottom(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetTopSticky(true)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"the first line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer, top sticky wins out arbitrarily when both set
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe second" + linebuffer.RST,
		"\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
		"the fourth line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StickyBottomLongLine(t *testing.T) {
	w, h := 10, 10
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	// stickyness should override maintain selection
	vp.SetSelectionComparator(ItemCompareFn)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"first line",
		"next line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"\x1b[38;2;0;0;255mnext line" + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"first line",
		"next line",
		"a very long line at the bottom that wraps many times",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"next line",
		"\x1b[38;2;0;0;255ma very lon" + linebuffer.RST,
		"\x1b[38;2;0;0;255mg line at " + linebuffer.RST,
		"\x1b[38;2;0;0;255mthe bottom" + linebuffer.RST,
		"\x1b[38;2;0;0;255m that wrap" + linebuffer.RST,
		"\x1b[38;2;0;0;255ms many tim" + linebuffer.RST,
		"\x1b[38;2;0;0;255mes" + linebuffer.RST,
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_RemoveLogsWhenSelectionBottom(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)

	// add LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
		"the fourth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe second" + linebuffer.RST,
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe fourth" + linebuffer.RST,
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove LineBuffer
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ChangeHeight(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(6)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"the second",
		" line",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"the fourth",
		" line",
		"the fifth ",
		"line",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the fourth",
		" line",
		"the fifth ",
		"line",
		"\x1b[38;2;0;0;255mthe sixth " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe sixth " + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ChangeContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"the second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		"\x1b[38;2;0;0;255mthe sixth " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove LineBuffer
	setContent(vp, []string{
		"the second line",
		"the third line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		" line",
		"\x1b[38;2;0;0;255mthe third " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove all LineBuffer
	setContent(vp, []string{})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add LineBuffer
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe first " + linebuffer.RST,
		"\x1b[38;2;0;0;255mline" + linebuffer.RST,
		"the second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StringToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetStringToHighlight("second")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(green),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"\x1b[38;2;0;255;0msecond" + linebuffer.RST,
		"\x1b[38;2;0;255;0msecond" + linebuffer.RST,
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetStringToHighlight("first")
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;255;0;0mfirst" + linebuffer.RST,
		"second",
		"second",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"averylongwordthatwrapsover",
	})
	vp.SetStringToHighlight("wraps")
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255maverylongw" + linebuffer.RST,
		"\x1b[38;2;0;0;255mordthat" + linebuffer.RST + "\x1b[38;2;255;0;0mwra" + linebuffer.RST,
		"\x1b[38;2;255;0;0mps" + linebuffer.RST + "\x1b[38;2;0;0;255mover" + linebuffer.RST,
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line a super long line ",
	})
	vp.SetStringToHighlight("l")
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255ma super " + linebuffer.RST + "\x1b[38;2;255;0;0ml" + linebuffer.RST + "\x1b[38;2;0;0;255mo" + linebuffer.RST,
		"\x1b[38;2;0;0;255mng " + linebuffer.RST + "\x1b[38;2;255;0;0ml" + linebuffer.RST + "\x1b[38;2;0;0;255mine a " + linebuffer.RST,
		"\x1b[38;2;0;0;255msuper " + linebuffer.RST + "\x1b[38;2;255;0;0ml" + linebuffer.RST + "\x1b[38;2;0;0;255mong" + linebuffer.RST,
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_RegexToHighlight(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	vp.SetRegexToHighlight(regexp.MustCompile("seco?nd"))
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(green),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
		SelectedItemStyle:        selectionStyle,
	})
	setContent(vp, []string{
		"first",
		"second",
		"second",
		"third",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst" + linebuffer.RST,
		"\x1b[38;2;0;255;0msecond" + linebuffer.RST,
		"\x1b[38;2;0;255;0msecond" + linebuffer.RST,
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetRegexToHighlight(regexp.MustCompile("fi?rst"))
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;255;0;0mfirst" + linebuffer.RST,
		"second",
		"second",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_StringToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		vp.SetWrapText(true)
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetStringToHighlight("r")
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 10),
			"100% (1/1)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOn_RegexToHighlightManyMatches(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		vp.SetWrapText(true)
		setContent(vp, []string{
			strings.Repeat("r", 100000),
		})
		vp.SetRegexToHighlight(regexp.MustCompile("r"))
		vp.SetStyles(Styles{
			FooterStyle:              lipgloss.NewStyle(),
			HighlightStyle:           lipgloss.NewStyle().Foreground(green),
			HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
			SelectedItemStyle:        selectionStyle,
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 10),
			strings.Repeat("\x1b[38;2;255;0;0mr"+linebuffer.RST+"", 10),
			"100% (1/1)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 10*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOn_AnsiOnSelection(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"line with some \x1b[38;2;255;0;0mred" + linebuffer.RST + " text",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mline with " + linebuffer.RST,
		"\x1b[38;2;0;0;255msome " + linebuffer.RST + "\x1b[38;2;255;0;0mred" + linebuffer.RST + "\x1b[38;2;0;0;255m t" + linebuffer.RST,
		"\x1b[38;2;0;0;255mext" + linebuffer.RST,
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SelectionEmpty(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m " + linebuffer.RST,
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_ExtraSlash(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"|2024|\x1b[38;2;0mfl..lq" + linebuffer.RST + "/\x1b[38;2;0mflask-3" + linebuffer.RST + "|",
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255m|2024|" + linebuffer.RST + "\x1b[38;2;0mfl.." + linebuffer.RST,
		"\x1b[38;2;0mlq" + linebuffer.RST + "\x1b[38;2;0;0;255m/" + linebuffer.RST + "\x1b[38;2;0mflask-3" + linebuffer.RST,
		"\x1b[38;2;0;0;255m|" + linebuffer.RST,
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SuperLongWrappedLine(t *testing.T) {
	runTest := func(t *testing.T) {
		w, h := 10, 5
		vp := newViewport(w, h)
		vp.SetHeader([]string{"header"})
		vp.SetSelectionEnabled(true)
		vp.SetWrapText(true)
		setContent(vp, []string{
			"smol",
			strings.Repeat("12345678", 1000000),
			"smol",
		})
		expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"\x1b[38;2;0;0;255msmol" + linebuffer.RST,
			"1234567812",
			"3456781234",
			"33% (1/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"\x1b[38;2;0;0;255m1234567812" + linebuffer.RST,
			"\x1b[38;2;0;0;255m3456781234" + linebuffer.RST,
			"\x1b[38;2;0;0;255m5678123456" + linebuffer.RST,
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"5678123456",
			"7812345678",
			"\x1b[38;2;0;0;255msmol" + linebuffer.RST,
			"100% (3/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 500*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOn_StringToHighlightAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"A💖中é",
		"A💖中éA💖中é",
	})
	vp.SetStringToHighlight("中é")
	vp.SetStyles(Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle().Foreground(green),
		HighlightStyleIfSelected: lipgloss.NewStyle().Foreground(red),
		SelectedItemStyle:        selectionStyle,
	})
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		"\x1b[38;2;0;0;255mA💖" + linebuffer.RST + "\x1b[38;2;255;0;0m中é" + linebuffer.RST,
		"A💖\x1b[38;2;0;255;0m中é" + linebuffer.RST + "A💖",
		"\x1b[38;2;0;255;0m中é" + linebuffer.RST,
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # OTHER

func TestViewport_SelectionOn_ToggleWrap_PreserveSelection(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"first line that is fairly long",
		"second line that is even much longer than the first",
		"third line that is fairly long",
		"fourth",
		"fifth line that is fairly long",
		"sixth",
	})

	// wrap off, selection on first line
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mfirst line t..." + linebuffer.RST,
		"second line ...",
		"third line t...",
		"fourth",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t..." + linebuffer.RST,
		"fourth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"longer than the",
		" first",
		"\x1b[38;2;0;0;255mthird line that" + linebuffer.RST,
		"\x1b[38;2;0;0;255m is fairly long" + linebuffer.RST,
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t..." + linebuffer.RST,
		"fourth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth line that",
		" is fairly long",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		"\x1b[38;2;0;0;255msixth" + linebuffer.RST,
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_ToggleWrap_PreserveSelectionInView(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"a really really really really really really really really really really really really long preamble",
		"first line that is fairly long",
		"second line that is even much longer than the first",
		"third line that is fairly long",
	})
	vp.SetSelectedItemIdx(3)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t..." + linebuffer.RST,
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap, full wrapped selection should remain in view
	vp.SetWrapText(true)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"longer than the",
		" first",
		"\x1b[38;2;0;0;255mthird line that" + linebuffer.RST,
		"\x1b[38;2;0;0;255m is fairly long" + linebuffer.RST,
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		"\x1b[38;2;0;0;255mthird line t..." + linebuffer.RST,
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_ToggleWrap_ScrollInBounds(t *testing.T) {
	w, h := 10, 7
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})

	// scroll to bottom with selection at top of that view
	vp.SetSelectedItemIdx(5)
	vp, _ = vp.Update(upKeyMsg)
	vp, _ = vp.Update(upKeyMsg)
	expectedView := pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"\x1b[38;2;0;0;255mthe fourth" + linebuffer.RST,
		"\x1b[38;2;0;0;255m line" + linebuffer.RST,
		"the fifth ",
		"line",
		"the sixth ",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the sec...",
		"the thi...",
		"\x1b[38;2;0;0;255mthe fou..." + linebuffer.RST,
		"the fif...",
		"the six...",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}
