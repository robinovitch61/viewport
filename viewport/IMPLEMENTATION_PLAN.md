
1. Clean up package structure/naming
2. Remove the concept of ItemGetter's - just pass in viewport.Item[] to content directly
3. Hide away everything that shouldn't be public about viewport, item, and filterableviewport
4. Consider batching sequential highlights together so \x1b0...abc\x1bm[0\x1b0...abc\x1bm[0 becomes \x1b0...abcabc\x1bm[0
