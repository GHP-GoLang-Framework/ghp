package codegen

import (
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

// parseAssembled confirms out is valid Go on its own (independent of
// Assemble's own internal format.Source check) and returns the parsed
// file so tests can inspect its import list precisely, without being
// tripped up by gofmt reordering imports alphabetically inside a single
// block - a plain string comparison would be brittle here.
func parseAssembled(t *testing.T, out string) *goast.File {
	t.Helper()
	fset := token.NewFileSet()
	f, err := goparser.ParseFile(fset, "generated.go", out, 0)
	if err != nil {
		t.Fatalf("Assemble output is not valid Go: %v\n---\n%s", err, out)
	}
	return f
}

// assertImports checks that f imports exactly want, regardless of order
// - catches a missing import and a duplicate/extra one with the same
// helper.
func assertImports(t *testing.T, f *goast.File, want ...string) {
	t.Helper()
	var got []string
	for _, imp := range f.Imports {
		got = append(got, imp.Path.Value)
	}
	if len(got) != len(want) {
		t.Fatalf("imports = %v, want exactly %v", got, want)
	}
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("import %s missing; imports = %v", w, got)
		}
	}
}

func TestAssembleOnlyText(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{ast.NewText("hi", 1)}}

	out, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	if f.Name.Name != "pages" {
		t.Errorf("package = %q, want %q", f.Name.Name, "pages")
	}
	// Text only needs io - fmt/html stay out because there is no
	// <go=...> in the page.
	assertImports(t, f, `"net/http"`, `"io"`)

	if !strings.Contains(out, "func Index(w http.ResponseWriter, r *http.Request) {") {
		t.Errorf("function signature not found in:\n%s", out)
	}
}

func TestAssembleEcho(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{ast.NewEcho("name", 1)}}

	out, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"fmt"`, `"html"`, `"io"`)
}

func TestAssembleNoOutputNeedsOnlyHTTP(t *testing.T) {
	// A page with no <go=...> and no plain text (only a statement)
	// does not need fmt/html/io - only net/http, for the function
	// signature.
	prog := &ast.Program{Nodes: []ast.Node{ast.NewStatement("_ = 1", 1)}}

	out, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`)
}

func TestAssembleUserImport(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Path: "strings"}}, 1),
		ast.NewText("hi", 2),
	}}

	out, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"io"`, `"strings"`)
}

func TestAssembleUserImportWithAlias(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Alias: "s", Path: "strings"}}, 1),
		ast.NewText("hi", 2),
	}}

	out, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	var found bool
	for _, imp := range f.Imports {
		if imp.Path.Value == `"strings"` {
			found = true
			if imp.Name == nil || imp.Name.Name != "s" {
				t.Errorf("strings import alias = %v, want \"s\"", imp.Name)
			}
		}
	}
	if !found {
		t.Fatal("import \"strings\" missing")
	}
}

func TestAssembleDedupesUserImportAgainstAuto(t *testing.T) {
	// The page declares <go:import ("fmt")> explicitly, but the echo
	// also needs fmt automatically - it can only appear once in the
	// import block, or the generated file won't compile.
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Path: "fmt"}}, 1),
		ast.NewEcho("name", 2),
	}}

	out, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"fmt"`, `"html"`, `"io"`)
}

func TestAssembleAliasedImportCollidingWithAutoFails(t *testing.T) {
	// <go:import (myio "io")> cannot be honored: genText/genEcho
	// always call io.WriteString (the default name), so a different
	// alias would leave "myio" with no matching import in the final
	// file - go build would fail with "undefined: myio" and no clue
	// that the cause is an alias on top of a package the page already
	// manages on its own.
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Alias: "myio", Path: "io"}}, 1),
		ast.NewText("hi", 2),
	}}

	_, err := Assemble("pages", "Index", prog)
	if err == nil {
		t.Fatal("Assemble() = nil error, want alias error on automatic package")
	}
}

func TestAssembleNestedImportFails(t *testing.T) {
	// <go:import> inside a go:if body makes no sense (Go has no
	// conditional import), but the parser accepts the tag anywhere it
	// accepts other tags - without this check, the import would simply
	// disappear (generateNode treats *ast.Import as a no-op at any
	// depth) and any code depending on it would only break later, at
	// go build.
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewIf("true",
			[]ast.Node{ast.NewImport([]ast.ImportPath{{Path: "strings"}}, 2)},
			nil, nil, 1),
	}}

	_, err := Assemble("pages", "Index", prog)
	if err == nil {
		t.Fatal("Assemble() = nil error, want nested <go:import> error")
	}
}

func TestAssembleConflictingAliasForSamePathFails(t *testing.T) {
	// Two different <go:import> tags for the same package, each with a
	// different alias - there is no way to tell which one the rest of
	// the page expects to use, so this is an error instead of silently
	// keeping only the first one.
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Alias: "a", Path: "strings"}}, 1),
		ast.NewImport([]ast.ImportPath{{Alias: "b", Path: "strings"}}, 2),
		ast.NewText("hi", 3),
	}}

	_, err := Assemble("pages", "Index", prog)
	if err == nil {
		t.Fatal("Assemble() = nil error, want conflicting aliases error")
	}
}

func TestAssembleDedupesRepeatedUserImport(t *testing.T) {
	// Two different <go:import> tags declaring the same package - it
	// can only appear once in the import block.
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewImport([]ast.ImportPath{{Path: "strings"}}, 1),
		ast.NewImport([]ast.ImportPath{{Path: "strings"}}, 2),
		ast.NewText("hi", 3),
	}}

	out, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}

	f := parseAssembled(t, out)
	assertImports(t, f, `"net/http"`, `"io"`, `"strings"`)
}

func TestAssembleInvalidFuncNameFails(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{ast.NewText("hi", 1)}}

	_, err := Assemble("pages", "invalid name", prog)
	if err == nil {
		t.Fatal("Assemble() = nil error, want error for invalid funcName")
	}
}

func TestAssembleUnsupportedNodeFails(t *testing.T) {
	// A node Generate has no generator for has to fail Assemble too,
	// instead of the rest of the file being written around it - see
	// TestGenerateErrorsOnUnsupportedNode for the same guard at the
	// Generate level.
	prog := &ast.Program{Nodes: []ast.Node{nil}}

	_, err := Assemble("pages", "Index", prog)
	if err == nil {
		t.Fatal("Assemble() = nil error, want unsupported node type error")
	}
}
