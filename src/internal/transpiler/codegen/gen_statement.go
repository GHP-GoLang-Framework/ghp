package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// genStatement emits raw Go code as-is. The parser already validated
// syntax; codegen just appends it followed by a newline.
//
// Output: <n.Code>\n
//
//	b – destination buffer
//	n – Statement node whose Code is emitted verbatim
func genStatement(b *strings.Builder, n *ast.Statement) {
	fmt.Fprintf(b, "%s\n", n.Code)
}
