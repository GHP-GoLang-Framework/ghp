package pages

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestDiscover builds a temporary tree with .ghp files, nested dirs,
// and non-.ghp files, then checks only the recursive .ghp pages come back
// keyed by their path relative to the walk root.
func TestDiscover(t *testing.T) {
	root := t.TempDir()

	write(t, root, "about.ghp")
	write(t, root, "blog/index.ghp")
	write(t, root, "blog/tech/go.ghp")
	write(t, root, "blog/notes.md")
	write(t, root, "assets/img.png")

	got, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

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

// TestDiscoverMissingDir verifies an unreadable root is reported instead of
// being flattened into "no pages found" - a mistyped project path must fail
// the build with the path it could not read, not build an empty site.
func TestDiscoverMissingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	got, err := Discover(missing)
	if err == nil {
		t.Fatalf("Discover(%q) = nil error, want error for missing root", missing)
	}
	if got != nil {
		t.Errorf("pages = %+v, want nil on error", got)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("error = %q, want it to name %q", err, missing)
	}
}

// TestDiscoverRelativeKeys checks the returned pages carry the
// expected package/function identifiers for their relative location.
func TestDiscoverRelativeKeys(t *testing.T) {
	root := t.TempDir()
	write(t, root, "blog/about.ghp")

	got, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
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
