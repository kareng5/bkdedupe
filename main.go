package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	var (
		inPath  string
		outPath string
		report  bool
	)
	flag.StringVar(&inPath, "in", "", "input bookmarks file in Netscape Bookmark File Format; defaults to stdin")
	flag.StringVar(&outPath, "out", "", "output file for deduplicated bookmarks; defaults to stdout")
	flag.BoolVar(&report, "report", false, "list duplicate URLs on stderr instead of writing deduplicated output")
	flag.Parse()

	in := os.Stdin
	if inPath != "" {
		f, err := os.Open(inPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bkdedupe:", err)
			os.Exit(1)
		}
		defer f.Close()
		in = f
	}

	var out io.Writer = os.Stdout
	if outPath != "" && !report {
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bkdedupe:", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}
	if report {
		out = io.Discard
	}

	result, err := dedupe(in, out, os.Stderr, report)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bkdedupe:", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "bkdedupe: %d bookmarks seen, %d duplicates %s\n",
		result.total, result.duplicates, verb(report))
}

func verb(report bool) string {
	if report {
		return "found"
	}
	return "removed"
}
