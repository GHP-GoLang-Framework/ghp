package router

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestSearchGhpFiles builds a temporary tree with .ghp files, nested dirs,
// and non-.ghp files, then checks only the recursive .ghp pages come back
// keyed by their path relative to the walk root.
func TestSearchGhpFiles(t *testing.T) {
	root := t.TempDir()

	write(t, root, "about.ghp")
	write(t, root, "blog/index.ghp")
	write(t, root, "blog/tech/go.ghp")
	write(t, root, "blog/notes.md")
	write(t, root, "assets/img.png")

	got := SearchGhpFiles(root)

	want := []string{
		"about.ghp",
		"blog/index.ghp",
		"blog/tech/go.ghp",
	}

	if len(got) != len(want) {
		t.Fatalf("got %d files, want %d: %+v", len(got), len(want), got)
	}

	gotPaths := make([]string, 0, len(got))
	for _, f := range got {
		rel := f.RelDir
		if rel != "" {
			rel += "/"
		}
		gotPaths = append(gotPaths, rel+f.FileName+".ghp")
	}

	if !reflect.DeepEqual(gotPaths, want) {
		t.Errorf("paths = %v, want %v", gotPaths, want)
	}
}

// TestSearchGhpFilesMissingDir verifies an unreadable root does not panic
// and simply yields no files.
func TestSearchGhpFilesMissingDir(t *testing.T) {
	got := SearchGhpFiles(filepath.Join(t.TempDir(), "does-not-exist"))

	if len(got) != 0 {
		t.Fatalf("got %d files from a missing dir, want 0", len(got))
	}
}

// TestSearchGhpFilesRelativeKeys checks the returned pages carry the
// expected package/function identifiers for their relative location.
func TestSearchGhpFilesRelativeKeys(t *testing.T) {
	root := t.TempDir()
	write(t, root, "blog/about.ghp")

	got := SearchGhpFiles(root)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}

	f := got[0]
	if f.FuncName != "About" || f.PkgName != "blog" {
		t.Errorf("got func=%q pkg=%q, want func=About pkg=blog", f.FuncName, f.PkgName)
	}
}

// write creates path under root with the given content, so tests can build
// fixtures without spawning external tools.
func write(t *testing.T, root, path string) {
	t.Helper()

	full := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(""), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
