package filterable_viewport

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/robinovitch61/bubbleo/viewport"
)

type Model[T viewport.Renderable] struct {
	Viewport *viewport.Model[T]
}

// New creates a new filterable viewport model
func New[T viewport.Renderable](width, height int, km KeyMap, styles viewport.Styles) *Model[T] {
	return &Model[T]{
		Viewport: viewport.New[T](width, height, km.ViewportKeyMap, styles),
	}
}

func (m *Model[T]) Init() tea.Cmd {
	return nil
}

func (m *Model[T]) Update(msg tea.Msg) (*Model[T], tea.Cmd) {
	var cmd tea.Cmd
	m.Viewport, cmd = m.Viewport.Update(msg)
	return m, cmd
}

func (m *Model[T]) View() string {
	return m.Viewport.View()
}
