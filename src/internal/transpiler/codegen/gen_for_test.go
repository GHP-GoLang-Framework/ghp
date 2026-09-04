package codegen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenFor(t *testing.T) {
	var b strings.Builder
	err := genFor(&b, ast.NewFor("i := 0; i < n; i++", nil, 1))
	if err != nil {
		t.Fatalf("genFor(): %v", err)
	}

	want := "for i := 0; i < n; i++ {\n}\n"
	if got := b.String(); got != want {
		t.Errorf("genFor() = %q, want %q", got, want)
	}
}

func TestGenForWithBody(t *testing.T) {
	var b strings.Builder
	body := []ast.Node{ast.NewText("hi", 2)}
	err := genFor(&b, ast.NewFor("_, item := range items", body, 1))
	if err != nil {
		t.Fatalf("genFor(): %v", err)
	}

	want := "for _, item := range items {\n" +
		"io.WriteString(w, \"hi\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("genFor() = %q, want %q", got, want)
	}
}

func TestGenForBodyError(t *testing.T) {
	var b strings.Builder
	err := genFor(&b, ast.NewFor("i := 0; i < 1; i++", []ast.Node{nil}, 1))
	if err == nil {
		t.Fatal("genFor() = nil error, want error from generateNodes")
	}
}
