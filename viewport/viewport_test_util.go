package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/viewport/internal"
	"github.com/robinovitch61/viewport/viewport/item"
)

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
