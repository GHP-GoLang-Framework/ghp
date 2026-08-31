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
	"io"
	"os"
	"strings"

	"ghp/src/internal/lint"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses args, validates the commit message in the --edit file and
// returns the exit code — separated from main() so it can be tested without
// killing the test process with os.Exit.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gitlint", flag.ContinueOnError)
	fs.SetOutput(stderr)
	edit := fs.String("edit", "", "path to the commit message file")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *edit == "" {
		fmt.Fprintln(stderr, "usage: gitlint --edit <file>")
		return 2
	}

	data, err := os.ReadFile(*edit)
	if err != nil {
		fmt.Fprintf(stderr, "gitlint: %v\n", err)
		return 1
	}

	// Only the header (first line) is validated; the body may span many
	// lines, so pass just the first line to Message.Validate.
	header, _, _ := strings.Cut(string(data), "\n")
	if msg := (lint.Message{Header: header}).Validate(); msg != "" {
		fmt.Fprintln(stderr, msg)
		return 1
	}
	return 0
}
