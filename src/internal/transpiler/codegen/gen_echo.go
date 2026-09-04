package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genEcho emits an HTML-escaped expression write.
//
// Output: io.WriteString(w, html.EscapeString(fmt.Sprint(<expr>)))
//
//	b – destination buffer
//	n – Echo node containing the Go expression to render
func genEcho(b *strings.Builder, n *ast.Echo) {
	fmt.Fprintf(b, "io.WriteString(w, html.EscapeString(fmt.Sprint(%s)))\n", n.Expr)
}
