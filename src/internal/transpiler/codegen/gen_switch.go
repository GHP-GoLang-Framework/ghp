package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genSwitch emits a Go switch with one case per ast.Case and an optional
// default, recursing into each branch body via generateNodes.
//
// Output:
//
//	switch <n.Expr> {
//	case <value>:
//	  <generateNodes(body)>
//	// ...
//	default:
//	  <generateNodes(n.Default)>  // when present
//	}
//
//	b – destination buffer
//	n – Switch node; Expr is the switched value, Cases/Default are branches
func genSwitch(b *strings.Builder, n *ast.Switch) error {
	fmt.Fprintf(b, "switch %s {\n", n.Expr)

	for _, c := range n.Cases {
		fmt.Fprintf(b, "case %s:\n", c.Value)
		if err := generateNodes(b, c.Body); err != nil {
			return err
		}
	}

	if n.Default != nil {
		b.WriteString("default:\n")
		if err := generateNodes(b, n.Default); err != nil {
			return err
		}
	}

	b.WriteString("}\n")
	return nil
}
