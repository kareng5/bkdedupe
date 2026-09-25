package main

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
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
			if href, ok := extractHref(line); ok {
				s.total++
				key := normalizeURL(href)
				if _, dup := seen[key]; dup {
					s.duplicates++
					keep = false
					if reportOnly {
						fmt.Fprintln(log, href)
					}
				} else {
					seen[key] = struct{}{}
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

// normalizeURL builds the key used for duplicate comparison. It folds
// together URLs that a browser treats as the same page but that come out
// as different strings depending on which device or import produced them:
// scheme case, host case, an explicit default port, a trailing slash on the
// path, and query parameter order. If the href doesn't parse as a URL at
// all (a bookmarklet's javascript: href, for example) it's compared as-is,
// since there's nothing safe to normalize.
func normalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	if (u.Scheme == "http" && strings.HasSuffix(u.Host, ":80")) ||
		(u.Scheme == "https" && strings.HasSuffix(u.Host, ":443")) {
		u.Host = u.Host[:strings.LastIndex(u.Host, ":")]
	}

	if u.Path == "" {
		u.Path = "/"
	} else if u.Path != "/" {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	// url.Values.Encode sorts by key, which is enough to make
	// ?a=1&b=2 and ?b=2&a=1 compare equal.
	u.RawQuery = u.Query().Encode()

	return u.String()
}
