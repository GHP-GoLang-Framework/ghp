package transpiler

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ghp/src/internal/colors"
	"ghp/src/internal/pages"
)

func writePage(t *testing.T, root, rel, content string) string {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
	return path
}

func TestTranspileSuccess(t *testing.T) {
	src := t.TempDir()
	writePage(t, src, "blog/about.ghp", `<go:import ("strings")/>
<h1>Hello, <go= strings.ToUpper("a") /></h1>
`)
	out := t.TempDir()

	page := pages.NewPage("blog/about.ghp")
	var buf bytes.Buffer
	if err := transpileTo(page, src, out, &buf); err != nil {
		t.Fatalf("transpileTo: %v", err)
	}

	got, err := os.ReadFile(page.Go(out))
	if err != nil {
		t.Fatalf("generated file: %v", err)
	}
	if !strings.Contains(string(got), "func About(w http.ResponseWriter, r *http.Request)") {
		t.Errorf("generated handler missing:\n%s", got)
	}

	if !strings.Contains(buf.String(), "["+colors.Green("OK")+"]") {
		t.Errorf("status line missing OK: %q", buf.String())
	}
}

func TestTranspileReadError(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()

	page := pages.NewPage("missing.ghp")
	var buf bytes.Buffer
	err := transpileTo(page, src, out, &buf)
	if err == nil {
		t.Fatal("expected error for missing page")
	}
	if !strings.Contains(buf.String(), "["+colors.Red("FAIL")+"]") {
		t.Errorf("status line missing FAIL: %q", buf.String())
	}
	if _, statErr := os.Stat(page.Go(out)); !os.IsNotExist(statErr) {
		t.Errorf("no .go file should be written on failure")
	}
}

func TestTranspileParseError(t *testing.T) {
	src := t.TempDir()
	writePage(t, src, "broken.ghp", "<go:if n == 1/>\n")
	out := t.TempDir()

	page := pages.NewPage("broken.ghp")
	err := transpileTo(page, src, out, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestTranspileCodegenError(t *testing.T) {
	src := t.TempDir()
	writePage(t, src, "/dup.ghp", `<go:import ("strings")/>
<go:import ("strings" as s)/>
`)
	out := t.TempDir()

	page := pages.NewPage("dup.ghp")
	err := transpileTo(page, src, out, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected alias conflict error")
	}
}

func TestTranspilePublic(t *testing.T) {
	src := t.TempDir()
	writePage(t, src, "index.ghp", "<h1>Home</h1>\n")
	out := t.TempDir()

	page := pages.NewPage("index.ghp")
	if err := Transpile(page, src, out); err != nil {
		t.Fatalf("Transpile: %v", err)
	}
	if _, err := os.Stat(page.Go(out)); err != nil {
		t.Errorf("generated file missing: %v", err)
	}
}

func TestStatusCol(t *testing.T) {
	if got := statusCol(100); got != 80 {
		t.Errorf("statusCol(100) = %d, want 80", got)
	}
	if got := statusCol(40); got != 32 {
		t.Errorf("statusCol(40) = %d, want 32", got)
	}
}

func TestTranspileTopLevelNoRelDir(t *testing.T) {
	src := t.TempDir()
	writePage(t, src, "index.ghp", "<h1>Home</h1>\n")
	out := t.TempDir()

	page := pages.NewPage("index.ghp")
	if err := transpileTo(page, src, out, &bytes.Buffer{}); err != nil {
		t.Fatalf("transpileTo: %v", err)
	}
}
