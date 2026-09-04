package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genFor emits a Go for loop, recursing into Body via generateNodes.
//
// Output:
//
//	for <n.Expr> {
//	  <generateNodes(n.Body)>
//	}
//
//	b – destination buffer
//	n – For node; Expr is the loop header, Body is the loop body
func genFor(b *strings.Builder, n *ast.For) error {
	fmt.Fprintf(b, "for %s {\n", n.Expr)

	if err := generateNodes(b, n.Body); err != nil {
		return err
	}

	b.WriteString("}\n")
	return nil
}
