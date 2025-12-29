package viewport

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/robinovitch61/bubbleo/internal"
	"github.com/robinovitch61/bubbleo/viewport/item"

	"github.com/muesli/termenv"

	"github.com/charmbracelet/lipgloss"
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

func objectsEqual(a, b object) bool {
	if a.item == nil || b.item == nil {
		return a.item == b.item
	}
	return a.item.Content() == b.item.Content()
}

var _ Object = object{}

var (
	downKeyMsg       = internal.MakeKeyMsg('j')
	halfPgDownKeyMsg = internal.MakeKeyMsg('d')
	fullPgDownKeyMsg = internal.MakeKeyMsg('f')
	upKeyMsg         = internal.MakeKeyMsg('k')
	halfPgUpKeyMsg   = internal.MakeKeyMsg('u')
	fullPgUpKeyMsg   = internal.MakeKeyMsg('b')
	goToTopKeyMsg    = internal.MakeKeyMsg('g')
	goToBottomKeyMsg = internal.MakeKeyMsg('G')
	selectionStyle   = internal.BlueFg
)

func newViewport(width, height int, options ...Option[object]) *Model[object] {
	styles := Styles{
		FooterStyle:              lipgloss.NewStyle(),
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	}

	options = append([]Option[object]{
		WithKeyMap[object](DefaultKeyMap()),
		WithStyles[object](styles),
	}, options...)

	return New[object](width, height, options...)
}

// # SELECTION DISABLED, WRAP OFF

func TestViewport_SelectionOff_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"."})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"..", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"...", "hi", "..."})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		internal.RedFg.Render("a") + " really rea...",
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
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		internal.RedFg.Render("a") + " really rea...",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		internal.RedFg.Render("a") + " really rea...",
		"",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h, WithStyles[object](Styles{
		FooterStyle:              internal.RedFg,
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	}))
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"1",
		"2",
		"3",
		internal.RedFg.Render("75% (3/4)"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"",
		"100% (2/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"long header ...",
		"123456789012345",
		"123456789012...",
		"",
		"100% (2/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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

func TestViewport_SelectionOff_WrapOff_EnsureItemInView(t *testing.T) {
	w, h := 10, 4
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth line that is really long",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(5, 0, 0, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth l...",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(5, len("sixth line"), len("sixth line "), 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"...h",
		"...h li...", // 's|ixth line '
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(5, len("sixth line that is really lon"), len("sixth line that is really long"), 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"...",
		"...ly long",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(1, 0, 0, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(4, 0, 0, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// ensure idempotence
	vp.EnsureItemInView(4, 0, 0, 0, 0)
	internal.CmpStr(t, expectedView, vp.View())

	// invalid values truncated
	vp.EnsureItemInView(4, -1, 1e9, 0, 0)
	internal.CmpStr(t, expectedView, vp.View())

	// full width ok
	vp.EnsureItemInView(4, 0, len("fifth"), 0, 0)
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_EnsureItemInViewVerticalPad(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	numItems := 100
	nums := make([]string, 0, numItems)
	for i := 0; i < numItems; i++ {
		nums = append(nums, strconv.Itoa(i+1))
	}
	setContent(vp, nums)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"1",
		"2",
		"3",
		"4",
		"4% (4/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "5" with verticalPad=1
	// should leave 1 line of context below
	vp.EnsureItemInView(4, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"3",
		"4",
		"5",
		"6",
		"6% (6/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll up to "3" with verticalPad=1
	// should leave 1 line of context above
	vp.EnsureItemInView(2, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"2",
		"3",
		"4",
		"5",
		"5% (5/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll up to visible "8" with verticalPad=2
	// should leave 2 lines of context above
	vp.EnsureItemInView(9, 0, 0, 0, 0) // reset to bottom
	vp.EnsureItemInView(7, 0, 0, 2, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"6",
		"7",
		"8",
		"9",
		"9% (9/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "99", not enough content below for verticalPad=3
	// pad below as much as possible
	vp.EnsureItemInView(0, 0, 0, 0, 0) // reset to top
	vp.EnsureItemInView(98, 0, 0, 3, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"97",
		"98",
		"99",
		"100",
		"100% (1...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "50", request more padding than is available given viewport height -> center item
	vp.EnsureItemInView(0, 0, 0, 0, 0) // reset to top
	vp.EnsureItemInView(49, 0, 0, 3, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"49",
		"50",
		"51",
		"52",
		"52% (52...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_EnsureItemInViewHorizontalPad(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"some line that is really long",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"some li...",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: pan right to space after "line" with horizontalPad=2
	// should leave 2 columns of padding to the right
	vp.EnsureItemInView(0, 0, 0, 0, 0) // reset to top
	vp.EnsureItemInView(0, len("some line"), len("some line "), 0, 2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"...line...", // 'so|me line_th|at is really long'
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: pan to the visible "me" of "some" with horizontalPad=1
	// should leave 1 column of context to the left
	vp.EnsureItemInView(0, len("so"), len("some"), 0, 1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"... lin...", // 's|o__ line t|hat is really long'
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: pan right to the " r" of "is really" with huge horizontalPad
	// should center the target portion horizontally
	vp.EnsureItemInView(0, len("some line that is"), len("some line that is r"), 0, 100)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"...s re...", // 'some line tha|t is__eall|y long'
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SetXOffset(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"the first line",
		"the second line",
	})
	initialExpectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the fir...",
		"the sec...",
		"",
		"100% (2/2)",
	})
	internal.CmpStr(t, initialExpectedView, vp.View())

	vp.SetXOffset(-1)
	internal.CmpStr(t, initialExpectedView, vp.View())

	vp.SetXOffset(0)
	internal.CmpStr(t, initialExpectedView, vp.View())

	vp.SetXOffset(4)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"...st line",
		"...ond ...",
		"",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(1000)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"...t line ",
		"...nd line",
		"",
		"100% (2/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"first l...",
		"second ...",
		"third l...",
		"fourth",
		"66% (4/6)",
	})
	validate(expectedView)

	// pan right
	vp.SetXOffset(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ine ...",
		"...ne t...",
		".",
		"...ne t...",
		"83% (5/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.SetXOffset(41)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		"...",
		"...ly long",
		"...",
		"100% (6/6)",
	})
	validate(expectedView)

	// set shorter Item
	setContent(vp, []string{
		"the first one",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...rst one",
		"",
		"",
		"",
		"100% (1/1)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove Item
	setContent(vp, []string{})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// re-add Item
	setContent(vp, []string{
		"first",
		"second",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"",
		"100% (2/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		internal.BlueFg.Render("third"),
		"fourth",
		"fifth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SetHighlights(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
	})
	highlights := []Highlight{
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   10,
				},
				Style: internal.RedFg,
			},
		},
		{
			ItemIndex: 2,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   9,
				},
				Style: internal.GreenFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first line",
		"the " + internal.RedFg.Render("second") + " line",
		"the " + internal.GreenFg.Render("third") + " line",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SetHighlightsStyledContent(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{
		internal.RedFg.Render("the first line"),
		internal.GreenFg.Render("the second line"),
		internal.BlueFg.Render("the third line"),
		internal.RedFg.Render("the fourth line"),
	})
	highlights := []Highlight{
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   10,
				},
				Style: internal.BlueFg,
			},
		},
		{
			ItemIndex: 2,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   9,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.RedFg.Render("the first line"),
		internal.GreenFg.Render("the ") + internal.BlueFg.Render("second") + internal.GreenFg.Render(" line"),
		internal.BlueFg.Render("the ") + internal.RedFg.Render("third") + internal.BlueFg.Render(" line"),
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOff_SetHighlightsAnsiUnicode(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	// A (1w, 1b), 💖 (2w, 4b), 中 (2w, 3b), é (1w, 3b) = 6w, 11b
	vp.SetHeader([]string{"A💖中é"})
	setContent(vp, []string{
		"A💖中é line",
		"another line",
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 1,
					End:   8,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		"A" + internal.RedFg.Render("💖中") + "é line",
		"another line",
		"",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP OFF

func TestViewport_SelectionOn_WrapOff_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"."})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"..", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"...",
		internal.BlueFg.Render("hi"),
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
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		internal.RedFg.Render("a") + " really rea...",
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
	if selectedItem := vp.GetSelectedItem(); selectedItem != nil && selectedItem.GetItem().Content() != "second" {
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
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		internal.RedFg.Render("a") + " really rea...",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(7)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really rea..."),
		internal.RedFg.Render("a") + " really rea...",
		"",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h, WithStyles[object](Styles{
		FooterStyle:              internal.RedFg,
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	}))
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("1"),
		"2",
		"3",
		internal.RedFg.Render("25% (1/4)"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		"second line",
		"third line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("    first li..."),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		internal.BlueFg.Render("line1"),
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		internal.BlueFg.Render("line2"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		internal.BlueFg.Render("line2"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		internal.BlueFg.Render("line2"),
		"",
		"100% (2/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"long header ...",
		internal.BlueFg.Render("123456789012345"),
		"123456789012...",
		"",
		"50% (1/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("123456789012345"),
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		internal.BlueFg.Render("second"),
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOff_EnsureItemInView(t *testing.T) {
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so last item in view
	vp.EnsureItemInView(5, 0, 0, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll so second item in view
	vp.EnsureItemInView(1, 0, 0, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("second"),
		"third",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// ensure idempotence
	vp.EnsureItemInView(1, 0, 0, 0, 0)
	internal.CmpStr(t, expectedView, vp.View())

	// invalid values truncated
	vp.EnsureItemInView(1, -1, 1e9, 0, 0)
	internal.CmpStr(t, expectedView, vp.View())

	// full width ok
	vp.EnsureItemInView(1, 0, len("second"), 0, 0)
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_EnsureItemInViewVerticalPad(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	numItems := 100
	nums := make([]string, 0, numItems)
	for i := 0; i < numItems; i++ {
		nums = append(nums, strconv.Itoa(i+1))
	}
	setContent(vp, nums)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("1"),
		"2",
		"3",
		"4",
		"1% (1/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "5" with verticalPad=1
	// should leave 1 line of context below
	vp.SetSelectedItemIdx(4)
	vp.EnsureItemInView(4, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"3",
		"4",
		selectionStyle.Render("5"),
		"6",
		"5% (5/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll up to "3" with verticalPad=1
	// should leave 1 line of context above
	vp.SetSelectedItemIdx(2)
	vp.EnsureItemInView(2, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"2",
		selectionStyle.Render("3"),
		"4",
		"5",
		"3% (3/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "8" with verticalPad=2
	// should leave 2 lines of context above
	vp.SetSelectedItemIdx(99) // reset to bottom
	vp.EnsureItemInView(99, 0, 0, 0, 0)
	vp.SetSelectedItemIdx(7)
	vp.EnsureItemInView(7, 0, 0, 2, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"6",
		"7",
		selectionStyle.Render("8"),
		"9",
		"8% (8/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "99", not enough content below for verticalPad=3
	// pad below as much as possible
	vp.SetSelectedItemIdx(0) // reset to top
	vp.EnsureItemInView(0, 0, 0, 0, 0)
	vp.SetSelectedItemIdx(98)
	vp.EnsureItemInView(98, 0, 0, 3, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"97",
		"98",
		selectionStyle.Render("99"),
		"100",
		"99% (99...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "50", request more padding than is available given viewport height -> center item
	vp.SetSelectedItemIdx(0) // reset to top
	vp.EnsureItemInView(0, 0, 0, 0, 0)
	vp.SetSelectedItemIdx(49)
	vp.EnsureItemInView(49, 0, 0, 3, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"49",
		selectionStyle.Render("50"),
		"51",
		"52",
		"50% (50...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_EnsureItemInViewHorizontalPad(t *testing.T) {
	w, h := 10, 3
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"some line that is really long",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("some li..."),
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: pan right to space after "line" with horizontalPad=2
	// should leave 2 columns of padding to the right
	vp.SetSelectedItemIdx(0) // reset to top
	vp.EnsureItemInView(0, 0, 0, 0, 0)
	vp.EnsureItemInView(0, len("some line"), len("some line "), 0, 2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("...line..."), // 'so|me line_th|at is really long'
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: pan to the visible "me" of "some" with horizontalPad=1
	// should leave 1 column of context to the left
	vp.EnsureItemInView(0, len("so"), len("some"), 0, 1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("... lin..."), // 's|o__ line t|hat is really long'
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: pan right to the " r" of "is really" with huge horizontalPad
	// should center the target portion horizontally
	vp.EnsureItemInView(0, len("some line that is"), len("some line that is r"), 0, 100)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("...s re..."), // 'some line tha|t is__eall|y long'
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SetXOffset(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
	})
	initialExpectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("the fir..."),
		"the sec...",
		"",
		"50% (1/2)",
	})
	internal.CmpStr(t, initialExpectedView, vp.View())

	vp.SetXOffset(-1)
	internal.CmpStr(t, initialExpectedView, vp.View())

	vp.SetXOffset(0)
	internal.CmpStr(t, initialExpectedView, vp.View())

	vp.SetXOffset(4)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("...st line"),
		"...ond ...",
		"",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(1000)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("...t line"),
		"...nd line",
		"",
		"50% (1/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("third"),
		"fourth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("fourth"),
		"fifth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		internal.BlueFg.Render("fourth"),
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		internal.BlueFg.Render("third"),
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		internal.BlueFg.Render("second"),
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		internal.BlueFg.Render("first l..."),
		"second ...",
		"third l...",
		"fourth",
		"16% (1/6)",
	})
	validate(expectedView)

	// pan right
	vp.SetXOffset(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		internal.BlueFg.Render("...ne t..."),
		"...ine ...",
		"...ne t...",
		".",
		"16% (1/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ne t...",
		internal.BlueFg.Render("...ine ..."),
		"...ne t...",
		".",
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.SetXOffset(41)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...",
		internal.BlueFg.Render("...e first"),
		"...",
		"...",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...",
		"...e first",
		internal.BlueFg.Render("..."),
		"...",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...",
		"...e first",
		"...",
		internal.BlueFg.Render("..."),
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...e first",
		"...",
		"...",
		internal.BlueFg.Render("..."),
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		"...",
		"...ly long",
		internal.BlueFg.Render("..."),
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		"...",
		internal.BlueFg.Render("...ly long"),
		"...",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ly long",
		internal.BlueFg.Render("..."),
		"...ly long",
		"...",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		internal.BlueFg.Render("...ly long"),
		"...",
		"...ly long",
		"...",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		internal.BlueFg.Render("...n mu..."),
		"...ly long",
		"...",
		"...ly long",
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		internal.BlueFg.Render("...ly long"),
		"...n mu...",
		"...ly long",
		"...",
		"16% (1/6)",
	})
	validate(expectedView)

	// set shorter Item
	setContent(vp, []string{
		"the first one",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		internal.BlueFg.Render("...rst one"),
		"",
		"",
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_MaintainSelection(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetSelectionComparator(objectsEqual)
	setContent(vp, []string{
		"sixth",
		"seventh",
		"eighth",
		"ninth",
		"tenth",
		"eleventh",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("sixth"),
		"seventh",
		"eighth",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth",
		internal.BlueFg.Render("seventh"),
		"eighth",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item above
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth",
		internal.BlueFg.Render("seventh"),
		"eighth",
		"63% (7/11)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item below
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth",
		internal.BlueFg.Render("seventh"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetTopSticky(true)
	setContent(vp, []string{
		"first",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("second"),
		"first",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		internal.BlueFg.Render("first"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		internal.BlueFg.Render("first"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"first",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		internal.BlueFg.Render("first"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("second"),
		"first",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("second"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetBottomSticky(true)

	// test covers case where first set Item to empty, then overflow height
	setContent(vp, []string{})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		internal.BlueFg.Render("third"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetTopSticky(true)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"first",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item, top sticky wins out arbitrarily when both set
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("second"),
		"first",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		internal.BlueFg.Render("first"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"second",
		"first",
		"third",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		internal.BlueFg.Render("third"),
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"third",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"second",
		"first",
		"third",
		"fourth",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
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

	// add Item
	setContent(vp, []string{
		"second",
		"first",
		"third",
		"fourth",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("second"),
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("fourth"),
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove Item
	setContent(vp, []string{
		"second",
		"first",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		internal.BlueFg.Render("third"),
		"fourth",
		"fifth",
		"sixth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("third"),
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		internal.BlueFg.Render("third"),
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		internal.BlueFg.Render("sixth"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"second",
		"third",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove Item
	setContent(vp, []string{
		"second",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("second"),
		"",
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove all Item
	setContent(vp, []string{})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item (maintain selection off)
	setContent(vp, []string{
		"first",
		"second",
		"third",
		"fourth",
		"fifth",
		"sixth",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"second",
		"third",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_AnsiOnSelection(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"line with " + internal.RedFg.Render("red") + " text",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("line with ") + internal.RedFg.Render("red") + internal.BlueFg.Render(" text"),
		"",
		"",
		"100% (1/1)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render(" "),
		"",
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_ExtraSlash(t *testing.T) {
	w, h := 25, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"|2024|" + internal.RedFg.Render("fl..lq") + "/" + internal.RedFg.Render("flask-3") + "|",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("|2024|") + internal.RedFg.Render("fl..lq") + internal.BlueFg.Render("/") + internal.RedFg.Render("flask-3") + internal.BlueFg.Render("|"),
		"",
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SetHighlights(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   9,
				},
				Style: internal.GreenFg,
			},
		},
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   10,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the ") + internal.GreenFg.Render("first") + internal.BlueFg.Render(" line"),
		"the " + internal.RedFg.Render("second") + " line",
		"the third line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SetHighlightsStyledContent(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		internal.RedFg.Render("the first line"),
		internal.GreenFg.Render("the second line"),
		internal.BlueFg.Render("the third line"),
		internal.RedFg.Render("the fourth line"),
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   9,
				},
				Style: internal.GreenFg,
			},
		},
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 4,
					End:   10,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.RedFg.Render("the ") + internal.GreenFg.Render("first") + internal.RedFg.Render(" line"),
		internal.GreenFg.Render("the ") + internal.RedFg.Render("second") + internal.GreenFg.Render(" line"),
		internal.BlueFg.Render("the third line"),
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOff_SetHighlightsAnsiUnicode(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"A💖中é"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"A💖中é line",
		"another line",
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 1,
					End:   8,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		internal.BlueFg.Render("A") + internal.RedFg.Render("💖中") + internal.BlueFg.Render("é line"),
		"another line",
		"",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # SELECTION DISABLED, WRAP ON

func TestViewport_SelectionOff_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"h"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"he", "ad"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"hea", "der", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"head", "er", "hi", "1..."})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_Basic(t *testing.T) {
	w, h := 15, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
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
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
		internal.RedFg.Render("a") + " really really",
		"99% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
		internal.RedFg.Render("a") + " really really",
		" long line",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
		internal.RedFg.Render("a") + " really really",
		" long line",
		"",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h, WithStyles[object](Styles{
		FooterStyle:              internal.RedFg,
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	}))
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"1",
		"2",
		"3",
		internal.RedFg.Render("75% (3/4)"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"second line",
		"third line",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		"line2",
		"",
		"100% (2/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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

func TestViewport_SelectionOff_WrapOn_EnsureItemInView(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line that is super long",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first ",
		"line",
		"the second",
		" line",
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(2, 0, 9, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		"the second",
		" line",
		"the third",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the fourth",
		" line that",
		" is super ",
		"long",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(1, len("the second"), len("the second line"), 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		" line",
		"the third ",
		"line",
		"the fourth",
		"99% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(0, 0, 0, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first ",
		"line",
		"the second",
		" line",
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(3, 0, len("the fourth line that is super "), 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		"the fourth",
		" line that",
		" is super ",
		"99% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_EnsureItemInViewVerticalPad(t *testing.T) {
	w, h := 10, 10
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	numItems := 100
	nums := make([]string, 0, numItems)
	for i := 0; i < numItems; i++ {
		nums = append(nums, strconv.Itoa(i+1))
	}
	setContent(vp, nums)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"8% (8/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "10" with verticalPad=1
	// should leave 1 line of context below
	vp.EnsureItemInView(9, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
		"11% (11...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll up to "5" with verticalPad=1
	// should leave 1 line of context above
	vp.EnsureItemInView(4, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
		"11% (11...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "15" with verticalPad=2
	// should leave 2 lines of context above
	vp.EnsureItemInView(99, 0, 0, 0, 0) // reset to bottom
	vp.EnsureItemInView(14, 0, 0, 2, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"13",
		"14",
		"15",
		"16",
		"17",
		"18",
		"19",
		"20",
		"20% (20...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "99", not enough content below for verticalPad=3
	// pad below as much as possible
	vp.EnsureItemInView(0, 0, 0, 0, 0) // reset to top
	vp.EnsureItemInView(98, 0, 0, 3, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"93",
		"94",
		"95",
		"96",
		"97",
		"98",
		"99",
		"100",
		"100% (1...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "50", request more padding than is available given viewport height -> center item
	vp.EnsureItemInView(0, 0, 0, 0, 0) // reset to top
	vp.EnsureItemInView(49, 0, 0, 5, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"47",
		"48",
		"49",
		"50",
		"51",
		"52",
		"53",
		"54",
		"54% (54...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_EnsureItemInViewHorizontalPad(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"some line that is really long",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"some line ",
		"that is re",
		"ally long",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: ensure "line " is visible with horizontalPad=2
	// in wrap mode, horizontal padding ensures character ranges are visible
	vp.EnsureItemInView(0, 0, 0, 0, 0) // reset
	vp.EnsureItemInView(0, len("some line"), len("some line "), 0, 2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"some line ",
		"that is re",
		"ally long",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: ensure "really" is visible with horizontalPad=1
	vp.EnsureItemInView(0, len("some line that is "), len("some line that is really"), 0, 1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"some line ",
		"that is re",
		"ally long",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: ensure end of string is visible with large horizontalPad
	vp.EnsureItemInView(0, len("some line that is really lon"), len("some line that is really long"), 0, 100)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"some line ",
		"that is re",
		"ally long",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SetXOffset(t *testing.T) {
	w, h := 10, 8
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first ",
		"line",
		"the second",
		" line",
		"",
		"",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(-1)
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(0)
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(4)
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(1000)
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		" line",
		"the third ",
		"99% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		"the second",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third ",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	vp.SetXOffset(5)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	vp.SetXOffset(41)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third",
		"99% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first",
		"line",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to bottom
	vp, _ = vp.Update(fullPgDownKeyMsg)
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the third",
		"line",
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove Item
	setContent(vp, []string{
		"the first line",
		"the second line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove all Item
	setContent(vp, []string{})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
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
		expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"smol",
			"1234567812",
			"3456781234",
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"1234567812",
			"3456781234",
			"5678123456",
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"3456781234",
			"5678123456",
			"7812345678",
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(goToBottomKeyMsg)
		expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"very long ",
		"line",
		"another sh",
		"75% (3/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetSelectionEnabled(true)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		internal.BlueFg.Render("this is a "),
		internal.BlueFg.Render("very long "),
		internal.BlueFg.Render("line"),
		"50% (2/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SetHighlights(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first",
		"second line that wraps",
		"third",
	})
	highlights := []Highlight{
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 0,
					End:   6,
				},
				Style: internal.RedFg,
			},
		},
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 12,
					End:   16,
				},
				Style: internal.GreenFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		internal.RedFg.Render("second") + " lin",
		"e " + internal.GreenFg.Render("that") + " wra",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SetHighlightsStyledContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		internal.GreenFg.Render("first"),
		internal.BlueFg.Render("second line that wraps"),
		internal.RedFg.Render("third"),
	})
	highlights := []Highlight{
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 0,
					End:   6,
				},
				Style: internal.RedFg,
			},
		},
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 12,
					End:   16,
				},
				Style: internal.GreenFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.GreenFg.Render("first"),
		internal.RedFg.Render("second") + internal.BlueFg.Render(" lin"),
		internal.BlueFg.Render("e ") + internal.GreenFg.Render("that") + internal.BlueFg.Render(" wra"),
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOff_WrapOn_SetHighlightsAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"A💖中é"})
	vp.SetWrapText(true)
	setContent(vp, []string{
		"A💖中é text that wraps",
		"another line",
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 1,
					End:   8,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		"A" + internal.RedFg.Render("💖中") + "é tex",
		"t that wra",
		"ps",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # SELECTION ENABLED, WRAP ON

func TestViewport_SelectionOn_WrapOn_Empty(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{})
	internal.CmpStr(t, expectedView, vp.View())
	vp.SetHeader([]string{"header"})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"header"})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SmolDimensions(t *testing.T) {
	w, h := 0, 0
	vp := newViewport(w, h)
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	vp.SetHeader([]string{"header"})
	setContent(vp, []string{"hi"})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(1)
	vp.SetHeight(1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"h"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(2)
	vp.SetHeight(2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"he", "ad"})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(3)
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{"hea", "der", ""})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetWidth(4)
	vp.SetHeight(4)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"head",
		"er",
		internal.BlueFg.Render("hi"),
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
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
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
	if selectedItem := vp.GetSelectedItem(); selectedItem != nil && selectedItem.GetItem().Content() != "second" {
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
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really long line"),
		internal.RedFg.Render("a") + " really really long line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
		internal.RedFg.Render("a") + " really really",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
		internal.RedFg.Render("a") + " really really",
		" long line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(9)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		internal.RedFg.Render("second") + " line",
		internal.RedFg.Render("a really really"),
		internal.RedFg.Render(" long line"),
		internal.RedFg.Render("a") + " really really",
		" long line",
		"",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_FooterStyle(t *testing.T) {
	w, h := 15, 5
	vp := newViewport(w, h, WithStyles[object](Styles{
		FooterStyle:              internal.RedFg,
		HighlightStyle:           lipgloss.NewStyle(),
		HighlightStyleIfSelected: lipgloss.NewStyle(),
		SelectedItemStyle:        selectionStyle,
	}))
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"1",
		"2",
		"3",
		"4",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("1"),
		"2",
		"3",
		internal.RedFg.Render("25% (1/4)"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
		"second line",
		"third line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetFooterEnabled(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("    first line "),
		internal.BlueFg.Render("    "),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(4)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		internal.BlueFg.Render("line1"),
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		internal.BlueFg.Render("line2"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		internal.BlueFg.Render("line2"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header1",
		"header2",
		"line1",
		internal.BlueFg.Render("line2"),
		"",
		"100% (2/2)",
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"long header ove",
		"rflows",
		internal.BlueFg.Render("123456789012345"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"123456789012345",
		internal.BlueFg.Render("123456789012345"),
		internal.BlueFg.Render("6"),
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		internal.BlueFg.Render("second"),
		"third",
		"fourth",
		"33% (2/6)",
	})
	validate(expectedView)

	// scrolling down by one again
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		internal.BlueFg.Render("third"),
		"fourth",
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll to bottom
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	validate(expectedView)

	// scrolling down past bottom when at bottom is no-op
	vp, _ = vp.Update(downKeyMsg)
	validate(expectedView)
}

func TestViewport_SelectionOn_WrapOn_EnsureItemInView(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line that is super long",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"the second",
		" line",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(2, 0, 9, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("line"),
		"the second",
		" line",
		"the third",
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the fourth"),
		internal.BlueFg.Render(" line that"),
		internal.BlueFg.Render(" is super "),
		internal.BlueFg.Render("long"),
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(1, len("the second"), len("the second line"), 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		" line",
		"the third ",
		"line",
		internal.BlueFg.Render("the fourth"),
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(0, 0, 0, 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the first ",
		"line",
		"the second",
		" line",
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.EnsureItemInView(3, 0, len("the fourth line that is super "), 0, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		internal.BlueFg.Render("the fourth"),
		internal.BlueFg.Render(" line that"),
		internal.BlueFg.Render(" is super "),
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_EnsureItemInViewVerticalPad(t *testing.T) {
	w, h := 10, 10
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	numItems := 100
	nums := make([]string, 0, numItems)
	for i := 0; i < numItems; i++ {
		nums = append(nums, strconv.Itoa(i+1))
	}
	setContent(vp, nums)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("1"),
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"1% (1/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "10" with verticalPad=1
	// should leave 1 line of context below
	vp.SetSelectedItemIdx(9)
	vp.EnsureItemInView(9, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
		selectionStyle.Render("10"),
		"11",
		"10% (10...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll up to "5" with verticalPad=1
	// should leave 1 line of context above
	vp.SetSelectedItemIdx(4)
	vp.EnsureItemInView(4, 0, 0, 1, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"4",
		selectionStyle.Render("5"),
		"6",
		"7",
		"8",
		"9",
		"10",
		"11",
		"5% (5/100)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "15" with verticalPad=2
	// should leave 2 lines of context above
	vp.SetSelectedItemIdx(99) // reset to bottom
	vp.EnsureItemInView(99, 0, 0, 0, 0)
	vp.SetSelectedItemIdx(14)
	vp.EnsureItemInView(14, 0, 0, 2, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"13",
		"14",
		selectionStyle.Render("15"),
		"16",
		"17",
		"18",
		"19",
		"20",
		"15% (15...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "99", not enough content below for verticalPad=3
	// pad below as much as possible
	vp.SetSelectedItemIdx(0) // reset to top
	vp.EnsureItemInView(0, 0, 0, 0, 0)
	vp.SetSelectedItemIdx(98)
	vp.EnsureItemInView(98, 0, 0, 3, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"93",
		"94",
		"95",
		"96",
		"97",
		"98",
		selectionStyle.Render("99"),
		"100",
		"99% (99...",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scroll down to "50", request more padding than is available given viewport height -> center item
	vp.SetSelectedItemIdx(0) // reset to top
	vp.EnsureItemInView(0, 0, 0, 0, 0)
	vp.SetSelectedItemIdx(49)
	vp.EnsureItemInView(49, 0, 0, 5, 0)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"47",
		"48",
		"49",
		selectionStyle.Render("50"),
		"51",
		"52",
		"53",
		"54",
		"50% (50...",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_EnsureItemInViewHorizontalPad(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"some line that is really long",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("some line "),
		selectionStyle.Render("that is re"),
		selectionStyle.Render("ally long"),
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: ensure "line " is visible with horizontalPad=2
	// in wrap mode, horizontal padding ensures character ranges are visible
	vp.SetSelectedItemIdx(0) // reset
	vp.EnsureItemInView(0, 0, 0, 0, 0)
	vp.EnsureItemInView(0, len("some line"), len("some line "), 0, 2)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("some line "),
		selectionStyle.Render("that is re"),
		selectionStyle.Render("ally long"),
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: ensure "really" is visible with horizontalPad=1
	vp.EnsureItemInView(0, len("some line that is "), len("some line that is really"), 0, 1)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("some line "),
		selectionStyle.Render("that is re"),
		selectionStyle.Render("ally long"),
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// horizontalPad: ensure end of string is visible with large horizontalPad
	vp.EnsureItemInView(0, len("some line that is really lon"), len("some line that is really long"), 0, 100)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		selectionStyle.Render("some line "),
		selectionStyle.Render("that is re"),
		selectionStyle.Render("ally long"),
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SetXOffset(t *testing.T) {
	w, h := 10, 8
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetWrapText(true)
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"the first line",
		"the second line",
	})

	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"the second",
		" line",
		"",
		"",
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(-1)
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(0)
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(4)
	internal.CmpStr(t, expectedView, vp.View())

	vp.SetXOffset(1000)
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// full page down
	vp, _ = vp.Update(fullPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the second"),
		internal.BlueFg.Render(" line"),
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the third "),
		internal.BlueFg.Render("line"),
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page down
	vp, _ = vp.Update(halfPgDownKeyMsg)
	internal.CmpStr(t, expectedView, vp.View())

	// full page up
	vp, _ = vp.Update(fullPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the second"),
		internal.BlueFg.Render(" line"),
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// half page up
	vp, _ = vp.Update(halfPgUpKeyMsg)
	internal.CmpStr(t, expectedView, vp.View())

	// go to bottom
	vp, _ = vp.Update(goToBottomKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the third "),
		internal.BlueFg.Render("line"),
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// go to top
	vp, _ = vp.Update(goToTopKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
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
		// set Item multiple times to confirm no side effects of doing it
		internal.CmpStr(t, expectedView, vp.View())
		doSetContent()
		internal.CmpStr(t, expectedView, vp.View())
	}
	doSetContent()
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		internal.BlueFg.Render("first line"),
		internal.BlueFg.Render(" that is f"),
		internal.BlueFg.Render("airly long"),
		"second lin",
		"16% (1/6)",
	})
	validate(expectedView)

	// pan right
	vp.SetXOffset(5)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		internal.BlueFg.Render("second lin"),
		internal.BlueFg.Render("e that is "),
		internal.BlueFg.Render("even much "),
		internal.BlueFg.Render("longer tha"),
		"33% (2/6)",
	})
	validate(expectedView)

	// pan all the way right
	vp.SetXOffset(41)
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		internal.BlueFg.Render("third line"),
		internal.BlueFg.Render(" that is f"),
		internal.BlueFg.Render("airly long"),
		internal.BlueFg.Render(" as well"),
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"airly long",
		" as well",
		internal.BlueFg.Render("fourth kin"),
		internal.BlueFg.Render("da long"),
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"fourth kin",
		"da long",
		internal.BlueFg.Render("fifth kind"),
		internal.BlueFg.Render("a long too"),
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"da long",
		"fifth kind",
		"a long too",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		"da long",
		internal.BlueFg.Render("fifth kind"),
		internal.BlueFg.Render("a long too"),
		"sixth",
		"83% (5/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		internal.BlueFg.Render("fourth kin"),
		internal.BlueFg.Render("da long"),
		"fifth kind",
		"a long too",
		"66% (4/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		internal.BlueFg.Render("third line"),
		internal.BlueFg.Render(" that is f"),
		internal.BlueFg.Render("airly long"),
		internal.BlueFg.Render(" as well"),
		"50% (3/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		internal.BlueFg.Render("second lin"),
		internal.BlueFg.Render("e that is "),
		internal.BlueFg.Render("even much "),
		internal.BlueFg.Render("longer tha"),
		"33% (2/6)",
	})
	validate(expectedView)

	// scroll up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header lon",
		"g",
		internal.BlueFg.Render("first line"),
		internal.BlueFg.Render(" that is f"),
		internal.BlueFg.Render("airly long"),
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
	vp.SetSelectionComparator(objectsEqual)
	setContent(vp, []string{
		"sixth item",
		"seventh item",
		"eighth item",
		"ninth item",
		"tenth item",
		"eleventh item",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("sixth item"),
		"seventh it",
		"em",
		"eighth ite",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth item",
		internal.BlueFg.Render("seventh it"),
		internal.BlueFg.Render("em"),
		"eighth ite",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item above
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth item",
		internal.BlueFg.Render("seventh it"),
		internal.BlueFg.Render("em"),
		"eighth ite",
		"63% (7/11)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item below
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"sixth item",
		internal.BlueFg.Render("seventh it"),
		internal.BlueFg.Render("em"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetTopSticky(true)
	setContent(vp, []string{
		"the first line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the second"),
		internal.BlueFg.Render(" line"),
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection down
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"the first line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"",
		"",
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add longer Item at bottom
	setContent(vp, []string{
		"the second line",
		"the first line",
		"a very long line that wraps a lot",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("a very lon"),
		internal.BlueFg.Render("g line tha"),
		internal.BlueFg.Render("t wraps a "),
		internal.BlueFg.Render("lot"),
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"a very lon",
		"g line tha",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the second line",
		"the first line",
		"a very long line that wraps a lot",
		"the third line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetBottomSticky(true)

	// test covers case where first set Item to empty, then overflow height
	setContent(vp, []string{})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the third "),
		internal.BlueFg.Render("line"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetTopSticky(true)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"the first line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"100% (1/1)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item, top sticky wins out arbitrarily when both set
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the second"),
		internal.BlueFg.Render(" line"),
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the third "),
		internal.BlueFg.Render("line"),
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// de-activate by moving selection up
	vp, _ = vp.Update(upKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
		"the fourth line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
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
	vp.SetSelectionComparator(objectsEqual)
	vp.SetBottomSticky(true)
	setContent(vp, []string{
		"first line",
		"next line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		internal.BlueFg.Render("next line"),
		"",
		"",
		"",
		"",
		"",
		"",
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	setContent(vp, []string{
		"first line",
		"next line",
		"a very long line at the bottom that wraps many times",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line",
		"next line",
		internal.BlueFg.Render("a very lon"),
		internal.BlueFg.Render("g line at "),
		internal.BlueFg.Render("the bottom"),
		internal.BlueFg.Render(" that wrap"),
		internal.BlueFg.Render("s many tim"),
		internal.BlueFg.Render("es"),
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

	setContent(vp, []string{
		"the second line",
		"the first line",
		"the third line",
		"the fourth line",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the second"),
		"25% (1/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection to bottom
	vp.SetSelectedItemIdx(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the fourth"),
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove bottom items
	setContent(vp, []string{
		"the second line",
		"the first line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(6)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"the second",
		" line",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the second",
		" line",
		internal.BlueFg.Render("the third "),
		internal.BlueFg.Render("line"),
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the third "),
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// increase height
	vp.SetHeight(8)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the third "),
		internal.BlueFg.Render("line"),
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
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the fourth",
		" line",
		"the fifth ",
		"line",
		internal.BlueFg.Render("the sixth "),
		internal.BlueFg.Render("line"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// reduce height
	vp.SetHeight(3)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the sixth "),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"the second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to bottom
	vp.SetSelectedItemIdx(5)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"line",
		internal.BlueFg.Render("the sixth "),
		internal.BlueFg.Render("line"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove Item
	setContent(vp, []string{
		"the second line",
		"the third line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		" line",
		internal.BlueFg.Render("the third "),
		internal.BlueFg.Render("line"),
		"100% (2/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// remove all Item
	setContent(vp, []string{})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// add Item
	setContent(vp, []string{
		"the first line",
		"the second line",
		"the third line",
		"the fourth line",
		"the fifth line",
		"the sixth line",
	})
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the first "),
		internal.BlueFg.Render("line"),
		"the second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_AnsiOnSelection(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"line with some " + internal.RedFg.Render("red") + " text",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("line with "),
		internal.BlueFg.Render("some ") + internal.RedFg.Render("red") + internal.BlueFg.Render(" t"),
		internal.BlueFg.Render("ext"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render(" "),
		"",
		"",
		"100% (1/1)",
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
		"|2024|" + internal.RedFg.Render("fl..lq") + "/" + internal.RedFg.Render("flask-3") + "|",
	})
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("|2024|") + internal.RedFg.Render("fl.."),
		internal.RedFg.Render("lq") + internal.BlueFg.Render("/") + internal.RedFg.Render("flask-3"),
		internal.BlueFg.Render("|"),
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
		expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			internal.BlueFg.Render("smol"),
			"1234567812",
			"3456781234",
			"33% (1/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			internal.BlueFg.Render("1234567812"),
			internal.BlueFg.Render("3456781234"),
			internal.BlueFg.Render("5678123456"),
			"66% (2/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())

		vp, _ = vp.Update(downKeyMsg)
		expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
			"header",
			"5678123456",
			"7812345678",
			internal.BlueFg.Render("smol"),
			"100% (3/3)",
		})
		internal.CmpStr(t, expectedView, vp.View())
	}
	internal.RunWithTimeout(t, runTest, 500*time.Millisecond)
}

func TestViewport_SelectionOn_WrapOn_SetHighlights(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"first line that wraps",
		"second",
		"third",
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 0,
					End:   5,
				},
				Style: internal.GreenFg,
			},
		},
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 11,
					End:   15,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.GreenFg.Render("first") + internal.BlueFg.Render(" line"),
		internal.BlueFg.Render(" ") + internal.RedFg.Render("that") + internal.BlueFg.Render(" wrap"),
		internal.BlueFg.Render("s"),
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SetHighlightsStyledContent(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		internal.BlueFg.Render("first line that wraps"),
		internal.GreenFg.Render("second"),
		internal.RedFg.Render("third"),
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 0,
					End:   5,
				},
				Style: internal.GreenFg,
			},
		},
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 11,
					End:   15,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.GreenFg.Render("first") + internal.BlueFg.Render(" line"),
		internal.BlueFg.Render(" ") + internal.RedFg.Render("that") + internal.BlueFg.Render(" wrap"),
		internal.BlueFg.Render("s"),
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_SelectionOn_WrapOn_SetHighlightsAnsiUnicode(t *testing.T) {
	w, h := 10, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"A💖中é"})
	vp.SetSelectionEnabled(true)
	vp.SetWrapText(true)
	setContent(vp, []string{
		"A💖中é text that wraps",
		"another line",
	})
	highlights := []Highlight{
		{
			ItemIndex: 0,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 1,
					End:   8,
				},
				Style: internal.RedFg,
			},
		},
	}
	vp.SetHighlights(highlights)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"A💖中é",
		internal.BlueFg.Render("A") + internal.RedFg.Render("💖中") + internal.BlueFg.Render("é tex"),
		internal.BlueFg.Render("t that wra"),
		internal.BlueFg.Render("ps"),
		"50% (1/2)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

// # OTHER

func TestViewport_StyleOverlay(t *testing.T) {
	w, h := 20, 5
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header"})
	vp.SetSelectionEnabled(true)
	setContent(vp, []string{
		"plain text",
		internal.RedFg.Render("red text"),
		"more plain",
	})

	// add highlight to the second item which already has red styling
	highlights := []Highlight{
		{
			ItemIndex: 1,
			ItemHighlight: item.Highlight{
				ByteRangeUnstyledContent: item.ByteRange{
					Start: 0,
					End:   3,
				},
				Style: internal.GreenFg,
			},
		},
	}
	vp.SetHighlights(highlights)

	// first item is selected, highlight should show on second item
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("plain text"),
		internal.GreenFg.Render("red") + internal.RedFg.Render(" text"),
		"more plain",
		"33% (1/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// selection style on second item should override both the red styling and green highlight
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"plain text",
		internal.GreenFg.Render("red") + internal.RedFg.Render(" text"),
		"more plain",
		"66% (2/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to third item, highlight should show again on second item
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"plain text",
		internal.GreenFg.Render("red") + internal.RedFg.Render(" text"),
		internal.BlueFg.Render("more plain"),
		"100% (3/3)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first line t..."),
		"second line ...",
		"third line t...",
		"fourth",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to third line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line t...",
		"second line ...",
		internal.BlueFg.Render("third line t..."),
		"fourth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"longer than the",
		" first",
		internal.BlueFg.Render("third line that"),
		internal.BlueFg.Render(" is fairly long"),
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first line t...",
		"second line ...",
		internal.BlueFg.Render("third line t..."),
		"fourth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// move selection to last line
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	vp, _ = vp.Update(downKeyMsg)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap on
	vp.SetWrapText(true)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth line that",
		" is fairly long",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap off
	vp.SetWrapText(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third line t...",
		"fourth",
		"fifth line t...",
		internal.BlueFg.Render("sixth"),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		internal.BlueFg.Render("third line t..."),
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap, full wrapped selection should remain in view
	vp.SetWrapText(true)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"longer than the",
		" first",
		internal.BlueFg.Render("third line that"),
		internal.BlueFg.Render(" is fairly long"),
		"100% (4/4)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"a really rea...",
		"first line t...",
		"second line ...",
		internal.BlueFg.Render("third line t..."),
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
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("the fourth"),
		internal.BlueFg.Render(" line"),
		"the fifth ",
		"line",
		"the sixth ",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// toggle wrap
	vp.SetWrapText(false)
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"the sec...",
		"the thi...",
		internal.BlueFg.Render("the fou..."),
		"the fif...",
		"the six...",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func setContent(vp *Model[object], content []string) {
	renderableStrings := make([]object, len(content))
	for i := range content {
		renderableStrings[i] = object{item: item.NewItem(content[i])}
	}
	vp.SetObjects(renderableStrings)
}
