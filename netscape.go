package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

// Netscape Bookmark File Format (what Chrome, Firefox and Safari all export
// to) puts exactly one bookmark per line, as a <DT><A HREF="...">Title</A>
// tag. Everything else - the DOCTYPE, <H3> folder headers, <DL>/</DL>
// nesting - is structural and has no HREF, so it passes through untouched.
var hrefPattern = regexp.MustCompile(`(?i)<A\s[^>]*HREF="([^"]*)"`)

type stats struct {
	total      int
	duplicates int
}

// dedupe reads r one line at a time and writes each line to w, skipping
// every line whose HREF has already been seen. It only holds the set of
// URLs seen so far in memory, not the file itself, so it scales to bookmark
// exports too large to load in one go. log receives one URL per line when
// reportOnly is set.
func dedupe(r io.Reader, w io.Writer, log io.Writer, reportOnly bool) (stats, error) {
	var s stats
	seen := make(map[string]struct{})
	reader := bufio.NewReader(r)

	for {
		line, readErr := reader.ReadString('\n')
		if len(line) > 0 {
			keep := true
			if url, ok := extractHref(line); ok {
				s.total++
				if _, dup := seen[url]; dup {
					s.duplicates++
					keep = false
					if reportOnly {
						fmt.Fprintln(log, url)
					}
				} else {
					seen[url] = struct{}{}
				}
			}
			if keep && !reportOnly {
				if _, err := io.WriteString(w, line); err != nil {
					return s, err
				}
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return s, readErr
		}
	}
	return s, nil
}

func extractHref(line string) (string, bool) {
	m := hrefPattern.FindStringSubmatch(line)
	if m == nil {
		return "", false
	}
	return m[1], true
}
