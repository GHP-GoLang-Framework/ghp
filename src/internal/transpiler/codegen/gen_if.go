package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genIf emits a Go if/elif/else chain, recursing into Then, each ElseIf
// body, and (when present) Else via generateNodes.
//
// Output:
//
//	if <n.Cond> {
//	  <generateNodes(n.Then)>
//	}
//	// for each elif:
//	} else if <elif.Cond> {
//	  <generateNodes(elif.Body)>
//	}
//	// optional:
//	} else {
//	  <generateNodes(n.Else)>
//	}
//
//	b – destination buffer
//	n – If node; Cond is the condition, Then/Elifs/Else are branch bodies
func genIf(b *strings.Builder, file string, n *ast.If) error {
	writeLine(b, file, n.Line())
	fmt.Fprintf(b, "if %s {\n", n.Cond)

	if err := generateNodes(b, file, n.Then); err != nil {
		return err
	}

	for _, elif := range n.Elifs {
		writeLine(b, file, elif.Line)
		fmt.Fprintf(b, "} else if %s {\n", elif.Cond)
		if err := generateNodes(b, file, elif.Body); err != nil {
			return err
		}
	}

	if n.Else != nil {
		b.WriteString("} else {\n")
		if err := generateNodes(b, file, n.Else); err != nil {
			return err
		}
	}

	b.WriteString("}\n")
	return nil
}
