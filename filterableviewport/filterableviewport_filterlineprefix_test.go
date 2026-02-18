package filterableviewport

import (
	"testing"

	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/viewport/internal"
	"github.com/robinovitch61/viewport/viewport"
)

func TestFilterLinePrefixNoFilter(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object]("Prefix"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Prefix should be prepended to the empty text
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"line 1",
		"line 2",
		"line 3",
		"Prefix No Filter",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterLinePrefixWithActiveFilter(t *testing.T) {
	fv := makeFilterableViewport(
		60,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object]("Prefix"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Apply a filter
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('l'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	// Prefix should be prepended to the filter content
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		focusedStyle.Render("l") + "ine 1",
		unfocusedStyle.Render("l") + "ine 2",
		unfocusedStyle.Render("l") + "ine 3",
		"Prefix [exact] Filter: l  (1/3 matches on 3 items)",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterLinePrefixDuringEditing(t *testing.T) {
	fv := makeFilterableViewport(
		60,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithFilterLinePrefix[object]("Prefix"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Enter filter editing mode
	fv, _ = fv.Update(filterKeyMsg)

	// Prefix should be prepended even during editing
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"line 1",
		"line 2",
		"line 3",
		"Prefix [exact] Filter: " + cursorStyle.Render(" ") + " type to filter",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterLinePrefixWithPositionTop(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object]("Prefix"),
			WithFilterLinePosition[object](FilterLineTop),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Prefix at top position
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Prefix No Filter",
		"line 1",
		"line 2",
		"line 3",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterLinePrefixEmpty(t *testing.T) {
	fv := makeFilterableViewport(
		50,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object](""),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Empty prefix should behave the same as no prefix
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"line 1",
		"line 2",
		"line 3",
		"No Filter",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterLinePrefixWithFilterCancelRestore(t *testing.T) {
	fv := makeFilterableViewport(
		60,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object]("Prefix"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Initially shows prefix with empty text
	expectedInitial := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"line 1",
		"line 2",
		"line 3",
		"Prefix No Filter",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedInitial, fv.View())

	// Apply filter
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('l'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	// Cancel filter - should go back to prefix + empty text
	fv, _ = fv.Update(cancelFilterKeyMsg)
	internal.CmpStr(t, expectedInitial, fv.View())
}

func TestFilterLinePrefixTruncation(t *testing.T) {
	fv := makeFilterableViewport(
		20,
		4,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object]("VeryLongLabelText"),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
	}))

	// Prefix + empty text exceeds width, should truncate
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"line 1",
		"line 2",
		"VeryLongLabelText...",
		footerStyle.Render("100% (2/2)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterLinePrefixAndPositionTopWithActiveFilter(t *testing.T) {
	fv := makeFilterableViewport(
		60,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithPrefixText[object]("Filter:"),
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object]("Prefix"),
			WithFilterLinePosition[object](FilterLineTop),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Apply a filter
	fv, _ = fv.Update(filterKeyMsg)
	fv, _ = fv.Update(internal.MakeKeyMsg('l'))
	fv, _ = fv.Update(applyFilterKeyMsg)

	// Prefix at top with active filter
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"Prefix [exact] Filter: l  (1/3 matches on 3 items)",
		focusedStyle.Render("l") + "ine 1",
		unfocusedStyle.Render("l") + "ine 2",
		unfocusedStyle.Render("l") + "ine 3",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}

func TestFilterLinePrefixStyled(t *testing.T) {
	prefixStyle := lipgloss.NewStyle().Bold(true)
	fv := makeFilterableViewport(
		50,
		5,
		[]viewport.Option[object]{},
		[]Option[object]{
			WithEmptyText[object]("No Filter"),
			WithFilterLinePrefix[object](prefixStyle.Render("Prefix:")),
		},
	)
	fv.SetObjects(stringsToItems([]string{
		"line 1",
		"line 2",
		"line 3",
	}))

	// Styled prefix should render correctly
	expectedView := internal.Pad(fv.GetWidth(), fv.GetHeight(), []string{
		"line 1",
		"line 2",
		"line 3",
		prefixStyle.Render("Prefix:") + " No Filter",
		footerStyle.Render("100% (3/3)"),
	})
	internal.CmpStr(t, expectedView, fv.View())
}
