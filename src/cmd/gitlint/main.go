// Command gitlint validates a Conventional Commit message, the local
// replacement for the commitlint step of the commit-msg git hook.
//
// Usage: gitlint --edit <file>   # read the commit message file
//
// It prints nothing and exits 0 on success; on a violation it prints the
// rule that failed and exits 1, so a git hook can block the commit.
package main

import (
	"flag"
	"fmt"
	"os"

	"ghp/src/internal/lint"
)

func main() {
	edit := flag.String("edit", "", "path to the commit message file")
	flag.Parse()

	if *edit == "" {
		fmt.Fprintln(os.Stderr, "usage: gitlint --edit <file>")
		os.Exit(2)
	}

	data, err := os.ReadFile(*edit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gitlint: %v\n", err)
		os.Exit(1)
	}

	if msg := (lint.Message{Header: string(data)}).Validate(); msg != "" {
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}
}
