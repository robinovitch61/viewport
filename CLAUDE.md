# Viewport

A Go library for terminal-based viewports built on Bubble Tea.

## Reading Code

Don't eagerly read files that aren't immediately relevant to the task. Use grep to find what you need first.

These files are large and should be read with `offset` and `limit`, not in full:
- `viewport/viewport.go`
- `viewport/item/single.go`
- `viewport/item/concat.go`
- `viewport/item/ansi.go`
- `filterableviewport/filterableviewport.go`
- `filterableviewport/filterableviewport_test.go`
- `filterableviewport/filterableviewport_filterlineprefix_test.go`
- `viewport/viewport_selection_wrap_test.go`
- `viewport/viewport_selection_no_wrap_test.go`
- `viewport/viewport_no_selection_wrap_test.go`
- `viewport/viewport_no_selection_no_wrap_test.go`
- `viewport/viewport_prefooter_test.go`
- `viewport/item/single_test.go`
- `viewport/item/concat_test.go`
- `viewport/item/ansi_test.go`
- `viewport/item/string_test.go`

## Validation

Run the base `make` command to execute tests, formatting, linting, etc. This should be run to validate any change.

