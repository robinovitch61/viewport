# viewport

[![Go Reference](https://pkg.go.dev/badge/github.com/robinovitch61/viewport.svg)](https://pkg.go.dev/github.com/robinovitch61/viewport)

An advanced terminal viewport component for [Bubble Tea](https://github.com/charmbracelet/bubbletea) terminal UI (TUI) applications.

## Install

```sh
go get github.com/robinovitch61/viewport
```

## Features

Core `viewport`:

- Toggleable text wrapping
- Horizontal panning for unwrapped lines
- ANSI escape code and Unicode support
- Individual item selection
- Customizable styling
- Sticky top/bottom scrolling (auto-follow new content)
- Configurable sticky header
- Highlight ranges with custom styles
- Save viewport content to file
- Efficient item concatenation (e.g. prefixing line numbers via `MultiItem`)

The `filterableviewport` package wraps the core viewport and adds:

- Exact match, regex, and case-insensitive filtering
- Match highlighting with focused/unfocused styles
- Next/previous match navigation
- Matches-only view (hide non-matching items)
- Configurable match limit for large content

## Usage

Implement the `Object` interface on your type:

```go
import (
    "github.com/robinovitch61/viewport/viewport"
    "github.com/robinovitch61/viewport/viewport/item"
)

type myObject struct {
    item item.Item
}

func (o myObject) GetItem() item.Item {
    return o.item
}
```

Create a viewport and set content:

```go
vp := viewport.New[myObject](
    width, height,
    viewport.WithSelectionEnabled[myObject](true),
    viewport.WithWrapText[myObject](true),
)

objects := []myObject{
    {item: item.NewItem("first line")},
    {item: item.NewItem("second line")},
}

vp.SetObjects(objects)
```

Wire it into your Bubble Tea model's `Update` and `View`:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m model) View() string {
    return m.viewport.View()
}
```

### Filterable Viewport

Wrap an existing viewport to add filtering:

```go
import "github.com/robinovitch61/viewport/filterableviewport"

fvp := filterableviewport.New[myObject](
    vp,
    filterableviewport.WithPrefixText[myObject]("Filter:"),
    filterableviewport.WithEmptyText[myObject]("No Current Filter"),
    filterableviewport.WithMatchingItemsOnly[myObject](false),
    filterableviewport.WithCanToggleMatchingItemsOnly[myObject](true),
)

fvp.SetObjects(objects)
```

## Default Key Bindings

### Viewport Navigation

| Key | Action |
|---|---|
| `j` / `down` | Scroll down |
| `k` / `up` | Scroll up |
| `f` / `pgdown` / `ctrl+f` | Page down |
| `b` / `pgup` / `ctrl+b` | Page up |
| `d` / `ctrl+d` | Half page down |
| `u` / `ctrl+u` | Half page up |
| `g` / `ctrl+g` | Jump to top |
| `G` | Jump to bottom |
| `left` / `right` | Horizontal pan |

### Filterable Viewport

| Key | Action |
|---|---|
| `/` | Start exact filter |
| `r` | Start regex filter |
| `i` | Start case-insensitive filter |
| `enter` | Apply filter |
| `esc` | Cancel/clear filter |
| `n` | Next match |
| `N` (shift+n) | Previous match |
| `o` | Toggle matches-only view |

All key bindings are configurable via `WithKeyMap`.

## Examples

See the [`examples`](examples/) directory for runnable programs:

- **[viewport](examples/viewport/main.go)** -- core viewport with wrapping and selection toggles
- **[filterableviewport](examples/filterableviewport/main.go)** -- viewport with filtering, match navigation, and matches-only mode

```sh
go run ./examples/viewport
go run ./examples/filterableviewport
```

## Used By

- [lore](https://github.com/robinovitch61/lore) -- a `less`-like terminal pager
- [kl](https://github.com/robinovitch61/kl) -- a Kubernetes log viewer
