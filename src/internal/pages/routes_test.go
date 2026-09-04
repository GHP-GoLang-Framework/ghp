package pages

import (
	"strings"
	"testing"
)

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
