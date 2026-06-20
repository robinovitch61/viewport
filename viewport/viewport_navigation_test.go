package viewport

import (
	"testing"

	"github.com/robinovitch61/viewport/internal"
)

func TestViewport_Navigation_SelectionOff_GoToAndPaging(t *testing.T) {
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

	vp.PageDown()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.HalfPageDown()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.PageDown()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.PageUp()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.HalfPageUp()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.PageUp()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.GoToBottom()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.GoToTop()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"33% (2/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_Navigation_SelectionOn_GoToAndPaging(t *testing.T) {
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

	vp.PageDown()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("third"),
		"fourth",
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.HalfPageDown()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("fourth"),
		"fifth",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.PageDown()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.PageUp()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		internal.BlueFg.Render("fourth"),
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.HalfPageUp()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		internal.BlueFg.Render("third"),
		"50% (3/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.GoToBottom()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"fifth",
		internal.BlueFg.Render("sixth"),
		"100% (6/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.GoToTop()
	expectedView = internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		internal.BlueFg.Render("first"),
		"second",
		"16% (1/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())
}

func TestViewport_Navigation_SelectionOff_ScrollUpDown(t *testing.T) {
	w, h := 15, 6
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
	topView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"first",
		"second",
		"third",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, topView, vp.View())

	// scrolling up past top is a no-op
	vp.ScrollUp(1)
	internal.CmpStr(t, topView, vp.View())

	vp.ScrollDown(1)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"second",
		"third",
		"fourth",
		"fifth",
		"83% (5/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	// scrolling down by more than one line at a time
	vp.ScrollDown(1)
	bottomView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header",
		"third",
		"fourth",
		"fifth",
		"sixth",
		"100% (6/6)",
	})
	internal.CmpStr(t, bottomView, vp.View())

	// scrolling down past bottom is a no-op
	vp.ScrollDown(1)
	internal.CmpStr(t, bottomView, vp.View())

	// numLines argument scrolls multiple lines at once
	vp.ScrollUp(2)
	internal.CmpStr(t, topView, vp.View())

	vp.ScrollDown(2)
	internal.CmpStr(t, bottomView, vp.View())
}

func TestViewport_Navigation_SelectionOn_ScrollUpDown(t *testing.T) {
	w, h := 15, 6
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
	if got := vp.GetSelectedItemIdx(); got != 0 {
		t.Fatalf("expected selected item index 0, got %d", got)
	}

	// scrolling up past top is a no-op
	vp.ScrollUp(1)
	if got := vp.GetSelectedItemIdx(); got != 0 {
		t.Fatalf("expected selected item index 0, got %d", got)
	}

	vp.ScrollDown(1)
	if got := vp.GetSelectedItemIdx(); got != 1 {
		t.Fatalf("expected selected item index 1, got %d", got)
	}

	// numLines argument moves the selection by multiple items at once
	vp.ScrollDown(2)
	if got := vp.GetSelectedItemIdx(); got != 3 {
		t.Fatalf("expected selected item index 3, got %d", got)
	}

	vp.ScrollUp(2)
	if got := vp.GetSelectedItemIdx(); got != 1 {
		t.Fatalf("expected selected item index 1, got %d", got)
	}

	// scrolling down past bottom clamps to the last item
	vp.ScrollDown(100)
	if got := vp.GetSelectedItemIdx(); got != 5 {
		t.Fatalf("expected selected item index 5, got %d", got)
	}
}

func TestViewport_Navigation_ScrollLeftRight(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h)
	vp.SetHeader([]string{"header long"})
	setContent(vp, []string{
		"first line that is fairly long",
		"second line that is even much longer than the first",
		"third line that is fairly long",
		"fourth",
		"fifth line that is fairly long",
		"sixth",
	})
	leftView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"first l...",
		"second ...",
		"third l...",
		"fourth",
		"66% (4/6)",
	})
	internal.CmpStr(t, leftView, vp.View())

	// scrolling left past the start is a no-op
	vp.ScrollLeft(5)
	internal.CmpStr(t, leftView, vp.View())

	vp.ScrollRight(5)
	expectedView := internal.Pad(vp.GetWidth(), vp.GetHeight(), []string{
		"header ...",
		"...ne t...",
		"...ine ...",
		"...ne t...",
		".",
		"66% (4/6)",
	})
	internal.CmpStr(t, expectedView, vp.View())

	vp.ScrollLeft(5)
	internal.CmpStr(t, leftView, vp.View())
}

func TestViewport_Navigation_ScrollLeftRight_WrapOnIsNoOp(t *testing.T) {
	w, h := 10, 6
	vp := newViewport(w, h, WithWrapText[object](true))
	vp.SetHeader([]string{"header long"})
	setContent(vp, []string{
		"first line that is fairly long",
		"second line that is even much longer than the first",
	})
	wrappedView := vp.View()

	vp.ScrollRight(5)
	internal.CmpStr(t, wrappedView, vp.View())

	vp.ScrollLeft(5)
	internal.CmpStr(t, wrappedView, vp.View())
}
