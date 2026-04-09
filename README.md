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

- Customizable filter modes (exact, regex, case-insensitive built in; custom modes supported)
- Match highlighting with focused/unfocused styles
- Next/previous match navigation
- Matches-only view (hide non-matching items)
- Configurable match limit for large content
- Search history (up/down arrow while editing)

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

#### Custom Filter Keys

Use the built-in filter mode constructors with your own key bindings:

```go
import "charm.land/bubbles/v2/key"

fvp := filterableviewport.New[myObject](
    vp,
    filterableviewport.WithFilterModes[myObject]([]filterableviewport.FilterMode{
        filterableviewport.ExactFilterMode(
            key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
        ),
        filterableviewport.RegexFilterMode(
            key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "regex")),
        ),
        filterableviewport.CaseInsensitiveFilterMode(
            key.NewBinding(key.WithKeys("ctrl+i"), key.WithHelp("ctrl+i", "case insensitive")),
        ),
    }),
)
```

#### Custom Filter Modes

Define entirely custom filter logic. A `FilterMode` provides a name, key binding, label
shown in the filter line, and a `GetMatchFunc` that returns a `MatchFunc` for scanning items:

```go
import (
    "strings"

    "charm.land/bubbles/v2/key"
    "github.com/robinovitch61/viewport/filterableviewport"
    "github.com/robinovitch61/viewport/viewport/item"
)

// Define a name for your custom filter mode.
const FilterPrefix filterableviewport.FilterModeName = "prefix"

// A filter that only matches lines starting with the filter text.
prefixMode := filterableviewport.FilterMode{
    Name:  FilterPrefix,
    Key:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "prefix filter")),
    Label: "[prefix]",
    GetMatchFunc: func(filterText string) (filterableviewport.MatchFunc, error) {
        return func(content string) []item.ByteRange {
            if strings.HasPrefix(content, filterText) {
                return []item.ByteRange{{Start: 0, End: len(filterText)}}
            }
            return nil
        }, nil
    },
}

fvp := filterableviewport.New[myObject](
    vp,
    filterableviewport.WithFilterModes[myObject]([]filterableviewport.FilterMode{
        prefixMode,
        filterableviewport.RegexFilterMode(
            key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "regex")),
        ),
    }),
)
```

#### Programmatic Filter Control

Set and query the active filter using typed constants:

```go
// Set an exact filter programmatically
fvp.SetFilter("error", filterableviewport.FilterExact)

// Set a regex filter
fvp.SetFilter("err(or|ing)", filterableviewport.FilterRegex)

// Clear the filter
fvp.SetFilter("", "")

// Check what mode is active
if mode := fvp.GetActiveFilterMode(); mode != nil {
    fmt.Printf("Filtering with %s mode\n", mode.Name)
}
```

Built-in filter mode names: `FilterExact`, `FilterRegex`, `FilterCaseInsensitive`.

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
| `up` / `down` | Browse search history (while editing) |

Filter mode keys (`/`, `r`, `i`) are defined on each `FilterMode`, not in the `KeyMap`.
All other key bindings are configurable via `WithKeyMap`.

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
