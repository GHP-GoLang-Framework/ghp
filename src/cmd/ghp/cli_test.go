package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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

// minProject mirrors build.minProject: scaffold + one page.
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
	writeFile(t, root, "about.ghp", "<h1>About</h1>\n")
	return root
}

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"no args prints usage", nil, 2},
		{"help", []string{"help"}, 0},
		{"unknown command", []string{"bogus"}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := run(tt.args, &bytes.Buffer{}); got != tt.want {
				t.Errorf("run(%v) = %d, want %d", tt.args, got, tt.want)
			}
		})
	}
}

func TestRunBuildWithoutGoMod(t *testing.T) {
	dir := t.TempDir()
	if got := run([]string{"build", dir}, &bytes.Buffer{}); got != 1 {
		t.Errorf("run(build, empty dir) = %d, want 1", got)
	}
}

func TestRunDevWithoutGoMod(t *testing.T) {
	dir := t.TempDir()
	if got := run([]string{"dev", dir}, &bytes.Buffer{}); got != 1 {
		t.Errorf("run(dev, empty dir) = %d, want 1", got)
	}
}

func TestBuildSuccess(t *testing.T) {
	dir := minProject(t)
	if got := Build([]string{dir}); got != 0 {
		t.Fatalf("Build = %d, want 0", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "app")); err != nil {
		t.Errorf("app not built: %v", err)
	}
}

func TestBuildProjectError(t *testing.T) {
	dir := minProject(t)
	writeFile(t, dir, "broken.ghp", "<go:if n == 1/>\n")
	if got := Build([]string{dir}); got != 1 {
		t.Errorf("Build = %d, want 1", got)
	}
}

func TestDevSuccess(t *testing.T) {
	dir := minProject(t)
	if got := Dev([]string{dir}); got != 0 {
		t.Fatalf("Dev = %d, want 0", got)
	}
}

func TestGetPath(t *testing.T) {
	ok := minProject(t)
	got, err := getPath([]string{ok})
	if err != nil {
		t.Fatalf("getPath(ok dir) = %v", err)
	}
	if got != ok {
		t.Errorf("getPath = %q, want %q", got, ok)
	}

	if _, err := getPath([]string{t.TempDir()}); err == nil {
		t.Error("getPath(dir without go.mod) should fail")
	}

	noMain := t.TempDir()
	writeFile(t, noMain, "go.mod", "module example.com/site\n\ngo 1.26\n")
	if _, err := getPath([]string{noMain}); err == nil {
		t.Error("getPath(dir without main.go) should fail")
	}
}

func TestGetPathDefaultsToCwd(t *testing.T) {
	// The package dir has no go.mod, so the default "." must fail cleanly.
	if _, err := getPath(nil); err == nil {
		t.Error("getPath() with no args in the package dir should fail")
	}
}
