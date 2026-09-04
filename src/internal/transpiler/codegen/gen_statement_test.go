package codegen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenStatement(t *testing.T) {
	var b strings.Builder
	genStatement(&b, ast.NewStatement("x := 1", 1))

	want := "x := 1\n"
	if got := b.String(); got != want {
		t.Errorf("genStatement() = %q, want %q", got, want)
	}
}
