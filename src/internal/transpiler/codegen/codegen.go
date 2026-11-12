// Package codegen converts the tree defined by internal/ast into the Go
// source that makes up the body of a page's handler function.
//
// Generate does not produce a compilable file by itself: it only emits the
// statements that go inside the function body. The generated statements
// assume:
//
//   - a variable named w, implementing io.Writer, is already in scope
//   - the packages fmt, html and io are imported in the file that embeds
//     this output
//
// Wiring the function signature, those imports and the final import(...)
// block together (the package set ast.CollectImports resolved while the
// page was parsed) is the job of whoever assembles the full .go file,
// not this package.
package codegen

import (
	"fmt"
	"strings"

	"ghp/src/internal/ast"
)

// Generate walks nodes in source order and returns the Go statements that
// render them, one node at a time.
func Generate(nodes []ast.Node) (string, error) {
	return GenerateFile(nodes, "")
}

// GenerateFile is Generate that also emits a `//line file:N` directive
// before every statement that originates from template line N, so the Go
// compiler, stack traces and debuggers report positions in the .ghp file.
// An empty file disables the directives.
func GenerateFile(nodes []ast.Node, file string) (string, error) {
	var b strings.Builder
	if err := generateNodes(&b, file, nodes); err != nil {
		return "", err
	}
	return b.String(), nil
}

// generateNodes writes each node in nodes to b, in order. Besides backing
// Generate itself, this is what a tag with a nested body (<go:if>,
// <go:switch>, <go:for>) calls to render its own Then/Else/Cases/Body -
// see gen_if.go, gen_switch.go and gen_for.go for examples.
func generateNodes(b *strings.Builder, file string, nodes []ast.Node) error {
	for _, n := range nodes {
		if err := generateNode(b, file, n); err != nil {
			return err
		}
	}
	return nil
}

// generateNode dispatches a single node to its generator by concrete type.
// This is the extension point future tags plug into: each adds its own
// case here and implements the generator in its own gen_*.go file, so two
// tags being built at the same time only ever touch this one shared line
// each, not each other's code.
func generateNode(b *strings.Builder, file string, n ast.Node) error {
	switch node := n.(type) {
	case *ast.Import:
		// Imports don't render anything in the function body - GHP-11
		// collects every *ast.Import in the file separately, to build
		// the import(...) block once, deduplicated.
		return nil
	case *ast.Text:
		genText(b, file, node)
	case *ast.Echo:
		genEcho(b, file, node)
	case *ast.Statement:
		genStatement(b, file, node)
	case *ast.If:
		return genIf(b, file, node)
	case *ast.Switch:
		return genSwitch(b, file, node)
	case *ast.For:
		return genFor(b, file, node)
	default:
		// Every concrete type ast.Node currently seals is handled above,
		// so this is unreachable with real data today - Node can only
		// be implemented from inside package ast (see its node()
		// method), so nothing outside this switch can construct a value
		// that lands here; the one value that does is nil, which is why
		// %T is printed without touching any method on n. It stays as a
		// guardrail: if ast ever grows an 8th node type, this is what
		// stops it from silently vanishing from generated pages instead
		// of failing loudly.
		return fmt.Errorf("codegen: no generator registered for %T", n)
	}
	return nil
}
