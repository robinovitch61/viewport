package filterableviewport

import (
	"github.com/robinovitch61/bubbleo/internal"
	"github.com/robinovitch61/bubbleo/viewport"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
	"testing"
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

func 

func stringsToItems(vals []string) []viewport.Item {
	items := make([]viewport.Item, len(vals))
	for i, s := range vals {
		items[i] = viewport.Item{LineBuffer: linebuffer.New(s)}
	}
	return items
}
