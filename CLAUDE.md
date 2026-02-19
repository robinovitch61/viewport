# Viewport

A Go library for terminal-based viewports built on Bubble Tea.

## Validation

Run `make` to execute tests, linting, etc. This should be run to validate any change.

## Manual Validation with tmux

After making changes to viewport rendering, manually validate by running examples in a tmux session and capturing the output.

### Basic workflow

```bash
# Start the example in a detached tmux session
tmux new-session -d -s viewport -x 120 -y 40 "go run ./examples/viewport/main.go"

# Wait for the app to start, then capture the pane
sleep 2
tmux capture-pane -t viewport -p       # plain text
tmux capture-pane -t viewport -p -e    # with ANSI color/style escape sequences

# Clean up
tmux kill-session -t viewport
```

### Sending key input

Use `tmux send-keys` to interact with the running example. Add a short sleep before capturing to let the UI update.

```bash
# Scroll down one line (j or Down)
tmux send-keys -t viewport j && sleep 0.5 && tmux capture-pane -t viewport -p

# Jump to bottom (shift+g, sent as literal G)
tmux send-keys -t viewport G && sleep 0.5 && tmux capture-pane -t viewport -p

# Jump to top
tmux send-keys -t viewport g && sleep 0.5 && tmux capture-pane -t viewport -p

# Page down
tmux send-keys -t viewport f && sleep 0.5 && tmux capture-pane -t viewport -p

# Toggle wrapping
tmux send-keys -t viewport w && sleep 0.5 && tmux capture-pane -t viewport -p

# Toggle selection
tmux send-keys -t viewport s && sleep 0.5 && tmux capture-pane -t viewport -p
```

### Checking scroll position

When selection is off, the footer shows `X% (Y/Z)` where Y is the bottom visible line out of Z total objects. You can extract it with:

```bash
tmux capture-pane -t viewport -p | grep -o '[0-9]*% ([0-9]*/[0-9]*)'
```

### Default key bindings

See `viewport/keymap.go` for the full key map:
- `j` / `down` - scroll down
- `k` / `up` - scroll up
- `f` / `pgdown` / `ctrl+f` - page down
- `b` / `pgup` / `ctrl+b` - page up
- `d` / `ctrl+d` - half page down
- `u` / `ctrl+u` - half page up
- `g` / `ctrl+g` - top
- `shift+g` (G) - bottom
- `left` / `right` - horizontal pan

### Filterable viewport example

```bash
# Start the filterable viewport example
tmux new-session -d -s fvp -x 120 -y 40 "go run ./examples/filterableviewport/main.go"
sleep 2
```

The filterable viewport adds a filter bar at the bottom of the content area. When no filter is active it shows "No Current Filter". The quit key is `ctrl+c` only (not `q` or `esc`).

#### Filter workflow

```bash
# Start exact filter mode with '/'
tmux send-keys -t fvp / && sleep 0.5 && tmux capture-pane -t fvp -p -e

# Type a search term (each character is a separate argument)
tmux send-keys -t fvp U n i x && sleep 0.5 && tmux capture-pane -t fvp -p -e

# Apply the filter with Enter
tmux send-keys -t fvp Enter && sleep 0.5 && tmux capture-pane -t fvp -p -e

# Cancel the filter with Escape (clears filter, back to "No Current Filter")
tmux send-keys -t fvp Escape && sleep 0.5 && tmux capture-pane -t fvp -p -e
```

#### Regex filter workflow

```bash
# Start regex filter mode with 'r'
tmux send-keys -t fvp r && sleep 0.5

# Type a regex pattern (pipe character needs no escaping in tmux send-keys)
tmux send-keys -t fvp 'file|disk' && sleep 0.5 && tmux capture-pane -t fvp -p -e

# Apply with Enter
tmux send-keys -t fvp Enter && sleep 0.5 && tmux capture-pane -t fvp -p -e
```

#### Navigating matches and toggling matches-only

```bash
# Jump to next match
tmux send-keys -t fvp n && sleep 0.5 && tmux capture-pane -t fvp -p -e

# Jump to previous match (shift+n, sent as literal N)
tmux send-keys -t fvp N && sleep 0.5 && tmux capture-pane -t fvp -p -e

# Toggle showing matching items only (hides non-matching lines)
tmux send-keys -t fvp o && sleep 0.5 && tmux capture-pane -t fvp -p -e
```

#### What to look for in captured output

- **Filter bar** (second-to-last line inside the border): shows `[exact]` or `[regex]` mode, the filter text, and match counts like `(2/4 matches on 3 items)`. When capturing input, a cursor `[7m [0m` and "type to filter" prompt appear.
- **Match highlights in ANSI output (`-e` flag)**:
  - Active match: `[30m[103m...[39m[49m` (black text on bright yellow background)
  - Other matches: `[30m[47m...[39m[49m` (black text on white background)
- **Matches-only mode**: footer shows `showing matches only` and the object count drops to only matching items.

#### Filterable viewport key bindings

See `filterableviewport/keymap.go` for the full key map:
- `/` - start exact filter input
- `r` - start regex filter input
- `i` - start case-insensitive filter input
- `enter` - apply filter
- `esc` - cancel filter (clears it)
- `o` - toggle showing matching items only
- `n` - next match
- `shift+n` (N) - previous match

All viewport navigation keys (`j`/`k`, `f`/`b`, `g`/`G`, etc.) also work when not capturing filter input.

### Pipe viewer example

```bash
# Pipe viewer with a file
tmux new-session -d -s pv -x 120 -y 40 "go run ./examples/pipeviewer/main.go somefile.txt"
```
