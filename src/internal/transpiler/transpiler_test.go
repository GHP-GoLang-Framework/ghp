package transpiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
	os.MkdirAll(filepath.Join(dir, "blog"), 0o755)
	os.WriteFile(filepath.Join(dir, "blog", "[slug].ghp"), []byte(`<p>Slug: <go= r.PathValue("slug") /></p>`), 0o644)

	files, err := Generate(dir)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got := make(map[string]string, len(files))
	for _, f := range files {
		got[f.Name] = f.Content
	}

	for _, name := range []string{"pages/index.go", "pages/blog_slug.go", "main.go", "go.mod"} {
		if _, ok := got[name]; !ok {
			t.Errorf("missing generated file %q\nfiles: %v", name, keys(files))
		}
	}

	if !strings.Contains(got["main.go"], `mux.HandleFunc("/blog/{slug}", pages.BlogSlug)`) {
		t.Errorf("main.go missing slug route:\n%s", got["main.go"])
	}
	if !strings.Contains(got["main.go"], `mux.HandleFunc("/", pages.Index)`) {
		t.Errorf("main.go missing index route:\n%s", got["main.go"])
	}
	if !strings.Contains(got["pages/blog_slug.go"], `r.PathValue("slug")`) {
		t.Errorf("blog_slug.go missing PathValue:\n%s", got["pages/blog_slug.go"])
	}
	if !strings.Contains(got["main.go"], `"ghpapp/pages"`) {
		t.Errorf("main.go missing pages import:\n%s", got["main.go"])
	}
	if !strings.HasPrefix(got["go.mod"], "module ghpapp") {
		t.Errorf("go.mod missing module line:\n%s", got["go.mod"])
	}
}

func TestGenerateSyntaxError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.ghp"), []byte(`<go:if x/>`), 0o644)

	if _, err := Generate(dir); err == nil {
		t.Fatal("Generate() = nil error, want syntax error")
	} else if !strings.Contains(err.Error(), "bad.ghp") {
		t.Errorf("error missing file reference: %v", err)
	}
}

func TestGenerateConflictError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "blog.ghp"), []byte("x"), 0o644)
	os.MkdirAll(filepath.Join(dir, "blog"), 0o755)
	os.WriteFile(filepath.Join(dir, "blog", "index.ghp"), []byte("x"), 0o644)

	if _, err := Generate(dir); err == nil {
		t.Fatal("Generate() = nil error, want conflict")
	}
}

func TestGenerateReadError(t *testing.T) {
	dir := t.TempDir()
	os.Symlink(filepath.Join(dir, "missing.ghp"), filepath.Join(dir, "index.ghp"))

	if _, err := Generate(dir); err == nil {
		t.Fatal("Generate() = nil error, want read error")
	}
}

func TestGenerateAssembleError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<go:if true/>\n<go:import (\"fmt\")/>\n<go:endif/>"), 0o644)

	if _, err := Generate(dir); err == nil {
		t.Fatal("Generate() = nil error, want assemble error")
	}
}

func keys(files []File) []string {
	var ks []string
	for _, f := range files {
		ks = append(ks, f.Name)
	}
	return ks
}

func TestGoVersion(t *testing.T) {
	tests := []struct {
		name    string
		runtime string
		want    string
	}{
		{name: "patch release", runtime: "go1.22.1", want: "1.22.1"},
		{name: "minor release", runtime: "go1.23", want: "1.23"},
		{name: "custom toolchain suffix", runtime: "go1.26.5-X:nodwarf5", want: "1.26.5"},
		{name: "unrecognized falls back", runtime: "dev-toolchain", want: "1.22"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := goVersionFrom(tt.runtime); got != tt.want {
				t.Errorf("goVersionFor(%q) = %q, want %q", tt.runtime, got, tt.want)
			}
		})
	}
}
