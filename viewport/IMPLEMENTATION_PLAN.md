# TODOs

* Complete filterable viewport functionality + tests
* Hide away everything that shouldn't be public about viewport, item, and filterableviewport (remaining TODOs in Item)
* Consider batching sequential highlights together so \x1b0...abc\x1bm[0\x1b0...abc\x1bm[0 becomes \x1b0...abcabc\x1bm[0
* Handle carriage returns naturally within item's (i.e. "this\nline" produces 2 lines, even when not wrapped)
