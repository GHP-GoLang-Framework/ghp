// Package integration exercises multiple packages together (parser+codegen+project)
// or calls real `go build`, without needing the compiled `ghp` binary.
package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"ghp/src/internal/parser"
	"ghp/src/internal/transpiler/codegen"
)

// TestParseAssembleBuild is the GHP-11 end-to-end acceptance criterion:
// a page combining all 7 tag kinds has to produce a .go file that
// actually compiles with `go build`, not just parse cleanly on its own.
func TestParseAssembleBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode (go test -short)")
	}

	src := `<go:import ("strings")/>
<go
	name := "World"
/>
<h1>Hello, <go= strings.ToUpper(name) />!</h1>
<go:if (len(name) > 3)/>
<p>Long name</p>
<go:else/>
<p>Short name</p>
<go:endif/>
<go:for i := 0; i < 3; i++/>
<go:switch i/>
<go:case 0/>
<p>zero</p>
<go:default/>
<p>other</p>
<go:endswitch/>
<go:endfor/>
`

	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parser.Parse: %v", err)
	}

	out, err := codegen.Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("codegen.Assemble: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.go"), []byte(out), 0o644); err != nil {
		t.Fatalf("write index.go: %v", err)
	}
	goMod := "module pages\n\ngo 1.26\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	if buildOut, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s\n---\ngenerated file:\n%s", err, buildOut, out)
	}
}

// TestParseAssembleBuildElif verifies that <go:elif/> compiles into
// valid Go through the full parser → codegen → go build pipeline.
func TestParseAssembleBuildElif(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode (go test -short)")
	}

	src := `<go:import ("fmt")/>
<go
	n := 2
/>
<go:if n == 1/>
<p>one</p>
<go:elif n == 2/>
<p>two</p>
<go:elif n == 3/>
<p>three</p>
<go:else/>
<p>other</p>
<go:endif/>
<go= fmt.Sprintf("done:%d", n) />
`

	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parser.Parse: %v", err)
	}

	out, err := codegen.Assemble("pages", "Elif", prog)
	if err != nil {
		t.Fatalf("codegen.Assemble: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "elif.go"), []byte(out), 0o644); err != nil {
		t.Fatalf("write elif.go: %v", err)
	}
	goMod := "module pages\n\ngo 1.26\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = dir
	if buildOut, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s\n---\ngenerated file:\n%s", err, buildOut, out)
	}
}

// TestLineDirectivesPointToGhp proves compiler errors in generated code
// resolve to the original .ghp line, the foundation for debugger support.
func TestLineDirectivesPointToGhp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipped in short mode (go test -short)")
	}

	src := "<h1>Title</h1>\n<p>\n<go= undefinedName />\n</p>\n"
	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parser.Parse: %v", err)
	}
	ghp := "/abs/pages/index.ghp"
	out, err := codegen.AssembleFile("pages", "Index", ghp, prog)
	if err != nil {
		t.Fatalf("AssembleFile: %v", err)
	}

	dir := t.TempDir()
	for name, body := range map[string]string{"index.go": out, "go.mod": "module pages\n\ngo 1.26\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"build", "./..."}, {"vet", "./..."}} {
		cmd := exec.Command("go", args...)
		cmd.Dir = dir
		got, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("go %s succeeded, want undefined-name error", args[0])
		}
		if want := ghp + ":3:"; !strings.Contains(string(got), want) {
			t.Errorf("go %s output lacks %q:\n%s", args[0], want, got)
		}
	}
}
