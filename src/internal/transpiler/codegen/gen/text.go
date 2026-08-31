package gen

import (
	"fmt"
	"regexp"
	"strings"

	"ghp/src/internal/ast"
)

// lineBreakTab matches the layout-only whitespace (line breaks and tabs)
// that Text strips from content. Spaces are deliberately left alone: they
// may be meaningful user content.
var lineBreakTab = regexp.MustCompile(`[\n\r\t]`)

// Text emits a raw string write. The value's line breaks and tabs are
// removed outright (lineBreakTab) - they're just layout for the .ghp
// source, not content - while spaces are preserved verbatim (they can be
// meaningful, e.g. the " " in "Slug: <?= ... ?>"). A value left with only
// whitespace after stripping is dropped entirely.
//
// Output: io.WriteString(w, "<value>")
//
//	b   – destination buffer
//	n   – Text node whose Value is written to the response
func Text(b *strings.Builder, n *ast.Text) {
	value := lineBreakTab.ReplaceAllString(n.Value, "")
	if strings.TrimSpace(value) == "" {
		return
	}
	fmt.Fprintf(b, "io.WriteString(w, %q)\n", value)
}
