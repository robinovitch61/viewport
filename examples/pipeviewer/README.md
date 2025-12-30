# Pipe Viewer Example

A simple example application that reads from stdin and displays the content in a filterable viewport.

## Usage

Pipe any command output into the viewer:

```bash
# view a file
cat myfile.txt | go run main.go

# view command output
ls -la | go run main.go

# view git log
git log --oneline | go run main.go

# view process list
ps aux | go run main.go
```

## Features

- Displays piped input in real-time as it arrives (incremental loading)
- Auto-scrolls to bottom as new data arrives (sticky bottom enabled)
- Save current contents to timestamped file with `ctrl+s`
- Scrollable, filterable viewport with full regex support
- Navigate between matches with visual highlighting
- Toggle text wrapping and selection mode
- Vim-style and arrow key navigation
- Start filtering and navigating immediately, even while data is still loading

## Key Bindings

### Navigation
- Arrow keys or `j/k` - scroll up/down
- `g/G` - go to top/bottom
- `ctrl+d/ctrl+u` - half page down/up
- `ctrl+f/ctrl+b` - full page down/up
- `h/l` - scroll left/right

### Filtering
- `/` - start exact match filter
- `ctrl+r` - start regex filter
- `enter` - apply filter
- `esc` - cancel/clear filter
- `n` - next match
- `N` - previous match
- `ctrl+m` - toggle showing matching items only

### Display Options
- `w` - toggle text wrapping
- `s` - toggle selection mode

### File Operations
- `ctrl+s` - save current contents to timestamped file (e.g., `pipeviewer-20241229-143052.txt`)

### Other
- `ctrl+c` or `ctrl+d` - quit
