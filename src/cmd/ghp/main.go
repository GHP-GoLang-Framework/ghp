package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

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
