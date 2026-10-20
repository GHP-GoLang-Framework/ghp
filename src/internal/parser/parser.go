// Package parser turns .ghp source text into the tree defined by
// internal/ast.
//
// Parser only understands the shape of the template - the <go...> tags and
// how they nest inside one another. It knows nothing about Go syntax:
// whatever text sits inside a tag is carried byte-for-byte, and
// only becomes a real compile error later, when the generated file is
// built with `go build`.
package parser

import (
	"fmt"

	"ghp/src/internal/ast"
)

// Parse reads a .ghp source and returns its tree in source order, plus the
// set of packages the page needs (Program.Imports): the ones GHP generates
// code for (net/http, and fmt/html/io when the page echoes or prints) and
// every <go:import (...)/> the page declared, already deduplicated.
func Parse(source string) (*ast.Program, error) {
	nodes, _, err := parseNodes(newScanner(source), nil)
	if err != nil {
		return nil, err
	}

	imports, err := ast.CollectImports(nodes)
	if err != nil {
		return nil, err
	}

	return &ast.Program{Nodes: nodes, Imports: imports}, nil
}

// SyntaxError reports a problem found while scanning the source, together
// with the 1-based line it occurred on - so callers can point back at the
// exact spot in the original .ghp file, not just print a generic message.
type SyntaxError struct {
	Line    int
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("line %d: %s", e.Line, e.Message)
}
