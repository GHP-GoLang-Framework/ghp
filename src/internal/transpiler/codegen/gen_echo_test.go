package codegen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenEcho(t *testing.T) {
	var b strings.Builder
	genEcho(&b, ast.NewEcho("user.Name", 1))

	want := "io.WriteString(w, html.EscapeString(fmt.Sprint(user.Name)))\n"
	if got := b.String(); got != want {
		t.Errorf("genEcho() = %q, want %q", got, want)
	}
}
