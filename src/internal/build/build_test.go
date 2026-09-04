package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ghp/src/internal/pages"
)

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// minProject returns a temp dir with the minimum scaffold a build needs:
// a go.mod declaring the module and a main.go that mounts the generated
// routes. Pages are added by each test.
func minProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/site\n\ngo 1.26\n")
	writeFile(t, root, "main.go", `package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	AddRoutes(mux)
}
`)
	return root
}

func TestProjectSuccess(t *testing.T) {
	src := minProject(t)
	writeFile(t, src, "about.ghp", "<h1>About</h1>\n")
	writeFile(t, src, "blog/index.ghp", "<h1>Blog</h1>\n")

	if err := Project(src); err != nil {
		t.Fatalf("Project: %v", err)
	}

	for _, p := range []string{
		filepath.Join(src, "ghproutes.go"),
		filepath.Join(src, "app"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	routes, err := os.ReadFile(filepath.Join(src, "ghproutes.go"))
	if err != nil {
		t.Fatalf("read ghproutes.go: %v", err)
	}
	if !strings.Contains(string(routes), `mux.HandleFunc("GET /about", About)`) ||
		!strings.Contains(string(routes), `mux.HandleFunc("GET /blog", blog.Index)`) ||
		!strings.Contains(string(routes), `"example.com/site/blog"`) {
		t.Errorf("unexpected ghproutes.go:\n%s", routes)
	}
}

func TestProjectFailedPage(t *testing.T) {
	src := minProject(t)
	writeFile(t, src, "about.ghp", "<h1>About</h1>\n")
	writeFile(t, src, "broken.ghp", "<go:if n == 1/>\n")

	err := Project(src)
	if err == nil {
		t.Fatal("expected failure when a page does not transpile")
	}
	if !strings.Contains(err.Error(), "skipping binary") {
		t.Errorf("error = %q, want mention of skipping binary", err)
	}

	// Routes for the good pages must still be written, but no binary.
	if _, statErr := os.Stat(filepath.Join(src, "ghproutes.go")); statErr != nil {
		t.Errorf("ghproutes.go missing: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(src, "app")); !os.IsNotExist(statErr) {
		t.Errorf("app should not exist on partial failure")
	}
}

func TestProjectRoutesWriteError(t *testing.T) {
	src := minProject(t)
	writeFile(t, src, "about.ghp", "<h1>About</h1>\n")
	// A directory squatting on the output name makes WriteFile fail.
	if err := os.MkdirAll(filepath.Join(src, "ghproutes.go"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := Project(src); err == nil {
		t.Fatal("expected error when routes cannot be written")
	}
}

func TestProjectCompileError(t *testing.T) {
	src := minProject(t)
	writeFile(t, src, "main.go", "package main\nfunc\n")
	writeFile(t, src, "about.ghp", "<h1>About</h1>\n")

	err := Project(src)
	if err == nil {
		t.Fatal("expected compile error")
	}
	if !strings.Contains(err.Error(), "go build") {
		t.Errorf("error = %q, want mention of go build", err)
	}
}

func TestStageWalkError(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "unreadable/about.ghp", "<h1>About</h1>\n")
	locked := filepath.Join(src, "unreadable")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })

	if _, err := stage(src); err == nil {
		t.Fatal("expected error walking an unreadable dir")
	}
}

func TestReadModule(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")

	if err := os.WriteFile(gomod, []byte("module example.com/site\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if got := readModule(gomod); got != "example.com/site" {
		t.Errorf("readModule = %q, want %q", got, "example.com/site")
	}

	if got := readModule(filepath.Join(dir, "missing.mod")); got != "" {
		t.Errorf("readModule(missing) = %q, want empty", got)
	}

	if err := os.WriteFile(gomod, []byte("go 1.26\n"), 0o644); err != nil {
		t.Fatalf("rewrite go.mod: %v", err)
	}
	if got := readModule(gomod); got != "" {
		t.Errorf("readModule(no module) = %q, want empty", got)
	}
}

func TestWriteRoutes(t *testing.T) {
	dir := t.TempDir()
	files := []*pages.Page{
		{RelDir: "", FileName: "about", FuncName: "About", PkgName: "main", Route: "/about"},
	}
	tmpRoutes := filepath.Join(dir, "tmp-ghproutes.go")
	routesPath := filepath.Join(dir, "ghproutes.go")

	if err := writeRoutes(files, "example.com/site", tmpRoutes, routesPath); err != nil {
		t.Fatalf("writeRoutes: %v", err)
	}

	for _, p := range []string{tmpRoutes, routesPath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	got, err := os.ReadFile(tmpRoutes)
	if err != nil {
		t.Fatalf("read temp routes: %v", err)
	}
	if !strings.Contains(string(got), "GET /about") {
		t.Errorf("temp routes missing handler:\n%s", got)
	}
}
