// Package transpiler orchestrates the whole .ghp -> Go pipeline: it
// scans a directory (router), parses each file (parser) and emits both
// the handler functions (codegen) and the route wiring in main.go
// as the files of a runnable Go module.
package transpiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"ghp/src/internal/parser"
	"ghp/src/internal/transpiler/codegen"
	"ghp/src/internal/transpiler/router"
)

// File is one generated file of the output module: Name is the path
// relative to the module root, Content the full source.
type File struct {
	Name    string
	Content string
}

// Generate builds the Go module that serves the .ghp files under dir:
// one handler file per page (codegen.Assemble), a main.go wiring every
// route to its handler and mounting them on net/http.ServeMux, and a
// go.mod pinning the current toolchain.
//
// Files are returned, not written - the caller decides where they land
// (dev runs them from a temp dir, build from an output flag).
//
// Routes follow the router's file-based convention: blog/[slug].ghp
// becomes /blog/{slug}, which ServeMux matches natively, and the page
// reads the value with r.PathValue("slug").
func Generate(dir string) ([]File, error) {
	pages, err := router.Scan(dir)
	if err != nil {
		return nil, err
	}

	var files []File
	for _, p := range pages {
		src, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(p.GhpPath)))
		if err != nil {
			return nil, err
		}
		prog, err := parser.Parse(string(src))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.GhpPath, err)
		}
		out, err := codegen.Assemble("pages", p.FuncName, prog)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.GhpPath, err)
		}
		files = append(files, File{Name: filepath.Join("pages", p.GoFile), Content: out})
	}

	files = append(files,
		File{Name: "main.go", Content: mainSource(pages)},
		File{Name: "go.mod", Content: goModSource()},
	)
	return files, nil
}

// mainSource is the generated entrypoint: it mounts each page's handler
// on net/http.ServeMux and listens on GHP_PORT (default 8080). Handlers
// live in the pages subpackage and are referenced via pages.<FuncName>.
func mainSource(pages []router.Page) string {
	var b strings.Builder
	b.WriteString(`package main

import (
	"fmt"
	"net/http"
	"os"

	"ghpapp/pages"
)

func main() {
	port := os.Getenv("GHP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
`)
	for _, p := range pages {
		fmt.Fprintf(&b, "\tmux.HandleFunc(%q, pages.%s)\n", p.Route, p.FuncName)
	}
	b.WriteString(`
	fmt.Println("listening on http://localhost:" + port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`)
	return b.String()
}

// goModSource pins the module to the toolchain that generated it, so a
// built project always compiles with at least the Go that ran ghp.
// The module is named ghpapp (not pages) so that main.go's import of
// the generated "pages" subpackage stays unambiguous.
// runtime.Version may carry a toolchain build suffix (e.g. "go1.26.5-X"),
// which the go directive does not accept, so only the semver is kept.
func goModSource() string {
	return "module ghpapp\n\ngo " + goVersion() + "\n"
}

var goVersionRe = regexp.MustCompile(`^go(\d+\.\d+(?:\.\d+)?)`)

func goVersion() string {
	return goVersionFrom(runtime.Version())
}

// goVersionFrom extracts the semver from a Go toolchain version string.
// Ex: "go1.26.5-X:nodwarf5" -> "1.26.5"
func goVersionFrom(version string) string {
	if m := goVersionRe.FindStringSubmatch(version); m != nil {
		return m[1]
	}
	return "1.22"
}
