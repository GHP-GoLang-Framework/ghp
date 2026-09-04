package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ghp/src/internal/build"
)

// run executes the ghp command and returns the exit code — separated from
// main() so it can be tested without killing the test process with os.Exit.
func run(args []string, stdout io.Writer) int {
	if len(args) < 1 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "dev":
		return Dev(args[1:])
	case "build":
		return Build(args[1:])
	case "help":
		printUsage()
	default:
		printUsage()
		return 2
	}

	return 0
}

func printUsage() {
	fmt.Println(`
	GHP - Good Hygiene Practices

	Usage:
	  ghp <command> [dir]

	Commands:
	  dev [dir]    Start the dev server with live reload
	  build [dir]  Build the project into <dir>/build
	  help         Show this help message

	Run 'ghp help' for more information.`)
}

// Build transpiles every page, generates ghproutes.go and compiles the
// project in dir into a binary dropped on its root.
func Build(args []string) int {
	src, err := getPath(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 1
	}

	if err := build.Project(src); err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 1
	}
	return 0
}

// Dev is a build: live reload is not implemented yet, so it runs the same
// single-shot pipeline.
func Dev(args []string) int {
	return Build(args)
}

// getPath resolves the positional dir arg to an absolute path and checks
// the project has the go.mod and main.go a staged build needs.
func getPath(args []string) (string, error) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve build path: %w", err)
	}

	if _, err := os.Stat(filepath.Join(path, "go.mod")); err != nil {
		return "", fmt.Errorf("go.mod not found in %s: %w", path, err)
	}

	if _, err := os.Stat(filepath.Join(path, "main.go")); err != nil {
		return "", fmt.Errorf("main.go not found in %s: %w", path, err)
	}

	return path, nil
}
