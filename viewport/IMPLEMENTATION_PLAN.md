# TODOs

* Clean up package structure/naming
* Remove the concept of ItemGetter's - just pass in viewport.Item[] to content directly
* Hide away everything that shouldn't be public about viewport, item, and filterableviewport
* Consider batching sequential highlights together so \x1b0...abc\x1bm[0\x1b0...abc\x1bm[0 becomes \x1b0...abcabc\x1bm[0
* Handle carriage returns naturally within item's (i.e. "this\nline" produces 2 lines, even when not wrapped)
