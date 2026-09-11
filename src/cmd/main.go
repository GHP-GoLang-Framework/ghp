package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"ghp/src/internal/build"
)

// version is the build version, injected at release time via
// -ldflags "-X main.version=<tag>"; plain `go build` runs report "dev".
var version = "dev"

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
	case "version", "--version":
		fmt.Fprintln(stdout, version)
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
	  ghp build [flags] [path]
	  ghp build -p <path>      equivalent to the loose [path] argument
	  ghp dev   [flags] [path]

	Commands:
	  build [flags] [path]  Build the project (go.mod) into a binary
	  dev   [flags] [path]  Dev server (same single-shot pipeline for now)
	  version               Print the version (release tag, or "dev" on source builds)
	  help                  Show this help message

	Flags:
	  -o, --output <file>       Binary file path (default: <project>/app)
	  -e, --entry-point <dir>   Dir of the main package (default: dir of main.go)
	  -p, --project <path>      Project dir (go.mod); same as a bare path arg`)
}

// Build transpiles every page, generates ghproutes.go in the main package
// dir and compiles the project into the requested binary path. The staged
// dir is created inside Build and kept — Enter must be pressed to delete it,
// even when the build fails, so the generated files can be inspected.
func Build(args []string) int {
	p, err := resolveParameters(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 2
	}

	src, entry, output, err := resolveBuildPaths(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 1
	}

	tmpDir, err := build.Stage(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 1
	}

	perr := build.Compile(src, tmpDir, entry, output)
	build.RemoveStaged(tmpDir)
	if perr != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", perr)
		return 1
	}
	return 0
}

// Dev runs the same single-shot pipeline as Build for now, but owns the
// staged dir itself: it is created inside Dev, kept alive across reloads for
// the upcoming watcher (no re-staging per change) and removed when Dev ends.
func Dev(args []string) int {
	p, err := resolveParameters(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 2
	}

	src, entry, output, err := resolveBuildPaths(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 1
	}

	tmpDir, err := build.Stage(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	if err := build.Compile(src, tmpDir, entry, output); err != nil {
		fmt.Fprintf(os.Stderr, "ghp: %v\n", err)
		return 1
	}
	return 0
}

// ghpParameters holds the build-time overrides parsed from the CLI flags.
type ghpParameters struct {
	projectPath   string
	entryPointDir string
	output        string
}

// resolveParameters parses the build flags into ghpParameters. A bare path
// argument sets the project; "-p/--project" is the flag form of the same
// value, and "--name=value" and "--name value" are both accepted.
// Ex: ["-o", "bin/app", "--entry-point=cli", "site"] -> {site, cli, bin/app}.
func resolveParameters(args []string) (ghpParameters, error) {
	var p ghpParameters
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			if p.projectPath != "" {
				return p, fmt.Errorf("project path set twice: %q and %q", p.projectPath, arg)
			}
			p.projectPath = arg
			continue
		}

		name, value, hasValue := strings.Cut(arg, "=")
		if !hasValue {
			if i+1 >= len(args) {
				return p, fmt.Errorf("missing value for %s", name)
			}
			i++
			value = args[i]
		}

		switch name {
		case "-o", "--output":
			p.output = value
		case "-e", "--entry-point":
			p.entryPointDir = value
		case "-p", "--project":
			if p.projectPath != "" {
				return p, fmt.Errorf("project path set twice: %q and %q", p.projectPath, value)
			}
			p.projectPath = value
		default:
			return p, fmt.Errorf("unknown flag %s", name)
		}
	}
	return p, nil
}

// resolveBuildPaths completes the defaults (empty project -> ".", empty
// entry -> the dir holding the main package, empty output -> <project>/app),
// absolutizes them and checks the project has a go.mod and the entry point a
// main.go.
func resolveBuildPaths(p ghpParameters) (src, entry, output string, err error) {
	src = p.projectPath
	if src == "" {
		src = "."
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to resolve project path: %w", err)
	}

	entry = p.entryPointDir
	if entry == "" {
		entry = findMainDir(src)
		if entry == "" {
			entry = src
		}
	}
	entry, err = filepath.Abs(entry)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to resolve entry point: %w", err)
	}

	if _, err := os.Stat(filepath.Join(src, "go.mod")); err != nil {
		return "", "", "", fmt.Errorf("go.mod not found in %s: %w", src, err)
	}

	if _, err := os.Stat(filepath.Join(entry, "main.go")); err != nil {
		return "", "", "", fmt.Errorf("main.go not found in %s: %w", entry, err)
	}

	output = p.output
	if output == "" {
		output = filepath.Join(src, "app")
		return src, entry, output, nil
	}
	output, err = filepath.Abs(output)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to resolve output path: %w", err)
	}
	return src, entry, output, nil
}

// findMainDir locates the dir holding the main package under src, preferring
// a main.go at the project root; it returns "" when none is found. Ex: src
// with only src/main.go -> <src>/src.
func findMainDir(src string) string {
	if isMainFile(filepath.Join(src, "main.go")) {
		return src
	}
	var found string
	filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found != "" || d.IsDir() {
			return nil
		}
		if d.Name() == "main.go" && isMainFile(path) {
			found = filepath.Dir(path)
		}
		return nil
	})
	return found
}

// isMainFile reports whether path is a Go file declaring the main package.
// Ex: "/srv/pages/src/main.go" with "package main" -> true.
func isMainFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte("package main"))
}
