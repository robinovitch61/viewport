package viewport

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/robinovitch61/bubbleo/viewport/linebuffer"
)

// Test utility functions
func pad(width, height int, lines []string) string {
	var res []string
	for _, line := range lines {
		resLine := line
		numSpaces := width - lipgloss.Width(line)
		if numSpaces > 0 {
			resLine += strings.Repeat(" ", numSpaces)
		}
		res = append(res, resLine)
	}
	numEmptyLines := height - len(lines)
	for i := 0; i < numEmptyLines; i++ {
		res = append(res, strings.Repeat(" ", width))
	}
	return strings.Join(res, "\n")
}

func setContent(vp *Model[Item], content []string) {
	renderableStrings := make([]Item, len(content))
	for i := range content {
		renderableStrings[i] = Item{LineBuffer: linebuffer.New(content[i])}
	}
	vp.SetContent(renderableStrings)
}
