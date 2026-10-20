package pages

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// TestGenRoutesIsValidGo parses the generated file with the real Go parser
// so a regression in the import block (duplicate path, bad quoting, a
// missing net/http) fails here instead of at a later `go build`.
func TestGenRoutesIsValidGo(t *testing.T) {
	files := []*Page{
		{RelDir: "", FileName: "index", FuncName: "Index", PkgName: "main", Route: "/"},
		{RelDir: "blog", FileName: "index", FuncName: "Index", PkgName: "blog", Route: "/blog"},
		{RelDir: "blog", FileName: "[slug]", FuncName: "Slug", PkgName: "blog", Route: "/blog/{slug}"},
		{RelDir: "docs", FileName: "guide", FuncName: "Guide", PkgName: "docs", Route: "/docs/guide"},
		{RelDir: "docs/api", FileName: "v1", FuncName: "V1", PkgName: "api", Route: "/docs/api/v1"},
	}

	got := GenRoutes(files, "example.com/site")

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ghproutes.go", got, 0); err != nil {
		t.Fatalf("GenRoutes produced invalid Go: %v\n---\n%s", err, got)
	}
}

// TestGenRoutesOnlyMainHasNoSubdirImports pins the minimal case: a site
// whose pages all live in the main package imports net/http and nothing
// else, in the one-line form.
func TestGenRoutesOnlyMainHasNoSubdirImports(t *testing.T) {
	files := []*Page{
		{RelDir: "", FileName: "index", FuncName: "Index", PkgName: "main", Route: "/"},
		{RelDir: "", FileName: "about", FuncName: "About", PkgName: "main", Route: "/about"},
	}

	got := GenRoutes(files, "example.com/site")

	if !strings.Contains(got, `import "net/http"`) {
		t.Errorf("expected the one-line net/http import:\n%s", got)
	}
	if strings.Contains(got, "import (") {
		t.Errorf("no subdirectory page should mean no import block:\n%s", got)
	}
}

// TestGenRoutesDedupesSharedSubdirImport checks two pages in the same
// subdirectory yield a single import line for it - a duplicate would not
// compile.
func TestGenRoutesDedupesSharedSubdirImport(t *testing.T) {
	files := []*Page{
		{RelDir: "blog", FileName: "index", FuncName: "Index", PkgName: "blog", Route: "/blog"},
		{RelDir: "blog", FileName: "about", FuncName: "About", PkgName: "blog", Route: "/blog/about"},
	}

	got := GenRoutes(files, "example.com/site")

	if n := strings.Count(got, `"example.com/site/blog"`); n != 1 {
		t.Errorf("blog import appears %d times, want 1:\n%s", n, got)
	}
}

// TestGenRoutesEmptySiteStillImportsHTTP guards the degenerate case: no
// pages at all must still produce a file that compiles, which requires
// net/http for the ServeMux signatures.
func TestGenRoutesEmptySiteStillImportsHTTP(t *testing.T) {
	got := GenRoutes(nil, "example.com/site")

	if !strings.Contains(got, `import "net/http"`) {
		t.Errorf("empty site must still import net/http:\n%s", got)
	}

	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "ghproutes.go", got, 0); err != nil {
		t.Fatalf("empty-site routes are not valid Go: %v\n---\n%s", err, got)
	}
}

func TestGenRoutesTopLevel(t *testing.T) {
	files := []*Page{
		{RelDir: "", FileName: "about", FuncName: "About", PkgName: "main", Route: "/about"},
	}
	got := GenRoutes(files, "example.com/site")

	for _, want := range []string{
		"package main",
		`import "net/http"`,
		"mux.HandleFunc(\"GET /about\", About)",
		"func NewServer() *http.ServeMux",
		"func AddRoutes(mux *http.ServeMux)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}
}

func TestGenRoutesSubdirSortsAndImports(t *testing.T) {
	files := []*Page{
		{RelDir: "blog", FileName: "[slug]", FuncName: "Slug", PkgName: "blog", Route: "/blog/{slug}"},
		{RelDir: "", FileName: "index", FuncName: "Index", PkgName: "main", Route: "/"},
		{RelDir: "blog", FileName: "index", FuncName: "Index", PkgName: "blog", Route: "/blog"},
		{RelDir: "docs", FileName: "guide", FuncName: "Guide", PkgName: "docs", Route: "/docs/guide"},
	}
	got := GenRoutes(files, "example.com/site")

	for _, want := range []string{
		`import (`,
		`"example.com/site/blog"`,
		`"example.com/site/docs"`,
		"mux.HandleFunc(\"GET /\", Index)",
		"mux.HandleFunc(\"GET /blog\", blog.Index)",
		"mux.HandleFunc(\"GET /blog/{slug}\", blog.Slug)",
		"mux.HandleFunc(\"GET /docs/guide\", docs.Guide)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q:\n%s", want, got)
		}
	}

	// Routes must be emitted sorted by pattern.
	routes := []string{"GET /", "GET /blog", "GET /blog/{slug}", "GET /docs/guide"}
	pos := make([]int, len(routes))
	for i, r := range routes {
		idx := strings.Index(got, r)
		if idx < 0 {
			t.Fatalf("route %q not found", r)
		}
		pos[i] = idx
	}
	for i := 1; i < len(pos); i++ {
		if pos[i] < pos[i-1] {
			t.Errorf("routes out of order: %v", routes)
		}
	}
}
