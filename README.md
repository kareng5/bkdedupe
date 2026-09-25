# bkdedupe

A command line tool that removes duplicate bookmarks from a browser
bookmarks export, without loading the whole file into memory.

## The problem

Chrome, Firefox and Safari all export bookmarks to the same format
(Netscape Bookmark File Format, an HTML file with one `<A HREF="...">` tag
per bookmark). After a few years of bookmarking the same article twice,
importing an old backup on top of a live profile, or syncing across
devices, that export ends up full of exact duplicate URLs scattered across
different folders. Browsers have no built-in way to clean this up.

`bkdedupe` reads that export and drops every bookmark whose URL it has
already seen, keeping the first occurrence and leaving folder structure and
everything else in the file untouched.

## Usage

Write a cleaned copy to a new file:

```
bkdedupe -in bookmarks.html -out cleaned.html
```

Read from stdin, write to stdout:

```
cat bookmarks.html | bkdedupe > cleaned.html
```

See which URLs are duplicated without changing anything:

```
bkdedupe -in bookmarks.html -report
```

Sample input line and what happens to it:

```
<DT><A HREF="https://example.com/article" ADD_DATE="1690000000">Example</A>
```

The first time this URL appears, the line is copied to the output as-is.
Any later line with the same HREF is dropped (or, with `-report`, printed
to stderr as `https://example.com/article`).

## How it works

The file is read one line at a time. Only the set of URLs seen so far is
kept in memory - not the file contents - so a multi-gigabyte export costs
about as much memory as it has unique bookmarks, not as much as it is
bytes on disk.

This relies on the real-world shape of the format: every browser writes
one bookmark per line. A line with no `<A HREF="...">` tag (the DOCTYPE,
`<H3>` folder headers, `<DL>`/`</DL>` nesting) is passed through unchanged.

## Limitations

- Before comparing, URLs are normalized: scheme and host are lowercased, a
  default port (`:80` on `http`, `:443` on `https`) is dropped, a trailing
  slash on the path is ignored, and query parameters are reordered. So
  `https://x.com` and `HTTPS://X.com/` are treated as the same bookmark, but
  `http://x.com` and `https://x.com` are not, since that's a real difference
  in what gets fetched. The output line itself is left exactly as it was
  read - normalization only affects what counts as a duplicate.
- The host case-insensitivity above is host-only; nothing in the path or
  query is case-folded, since some servers treat path case as significant.
- Assumes one bookmark tag per line, which is what every major browser
  produces but is not guaranteed by the format itself.
- Only understands `HREF="..."` with double quotes.

## Building

Requires Go 1.22 or later and nothing else.

```
go build -o bkdedupe .
```

## License

MIT, see LICENSE.
