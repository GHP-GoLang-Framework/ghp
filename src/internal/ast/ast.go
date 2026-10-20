// Package ast defines the tree produced by parsing a .ghp source file.
//
// Every node knows the source line it started on, so the parser and the
// codegen that consumes this tree can both report errors that point back
// at the .ghp file the developer actually wrote.
package ast

// Node is implemented by every element of a parsed .ghp file
type Node interface {
	// Line reports the 1-based source line where this node begins.
	Line() int

	// node is unexported so Node can only be implemented by types in this
	// package - callers switch over the concrete type below instead of
	// inventing new ones
	node()
}

// Text is a run of raw output copied to the response verbatim
type Text struct {
	Value string
	line  int
}

// Import declares one or more Go packages needed by the page, from a
// <go:import (...)/> tag.
type Import struct {
	Paths []ImportPath
	line  int
}

// ImportPath is a single package path inside a <go:import (...)/> tag,
// with an optional alias (e.g. `f "fmt"`).
type ImportPath struct {
	Alias string
	Path  string
}

// Statement is a raw Go statement block from a <go ...> tag.
type Statement struct {
	Code string
	line int
}

// Echo renders a Go expression's value into the output, from a
// <go= ...> tag
type Echo struct {
	Expr string
	line int
}

// If is a conditional: <go:if cond/> Then [<go:elif .../>]* [<go:else/> Else] <go:endif/>
type If struct {
	Cond  string
	Then  []Node
	Elifs []ElseIf // zero or more <go:elif/> branches, in source order
	Else  []Node   // nil when there is no <go:else/>
	line  int
}

// ElseIf is a single <go:elif cond/> branch inside an If.
type ElseIf struct {
	Cond string
	Body []Node
	Line int
}

// Switch is a <go:switch expr/>...<go:endswitch/>
type Switch struct {
	Expr    string
	Cases   []Case
	Default []Node // nil when there is no <go:default/>
	line    int
}

// Case is a single <go:case value/> branch inside a Switch
type Case struct {
	Value string
	Body  []Node
	Line  int
}

// For is a loop: <go:for expr/> Body <go:endfor/>
type For struct {
	Expr string
	Body []Node
	line int
}

func (n *Text) Line() int      { return n.line }
func (n *Import) Line() int    { return n.line }
func (n *Statement) Line() int { return n.line }
func (n *Echo) Line() int      { return n.line }
func (n *If) Line() int        { return n.line }
func (n *Switch) Line() int    { return n.line }
func (n *For) Line() int       { return n.line }

func (*Text) node()      {}
func (*Import) node()    {}
func (*Statement) node() {}
func (*Echo) node()      {}
func (*If) node()        {}
func (*Switch) node()    {}
func (*For) node()       {}

var _ Node = (*Text)(nil)
var _ Node = (*Import)(nil)
var _ Node = (*Statement)(nil)
var _ Node = (*Echo)(nil)
var _ Node = (*If)(nil)
var _ Node = (*Switch)(nil)
var _ Node = (*For)(nil)

// Program is the root of a parsed .ghp file: the full sequence of
// top-level nodes, in source order.
type Program struct {
	Nodes []Node

	// Imports is the deduplicated set of packages the generated handler
	// needs, keyed by import path - the ones GHP brings on its own
	// (net/http for the signature, fmt/html/io for what the echo and
	// text emitters generate) plus every <go:import (...)/> the page
	// declared. Ex: {"fmt": {Path: "fmt"}, "strings": {Alias: "s",
	// Path: "strings"}}. See CollectImports.
	Imports map[string]ImportPath
}

func NewText(value string, line int) *Text {
	return &Text{Value: value, line: line}
}

func NewImport(paths []ImportPath, line int) *Import {
	return &Import{Paths: paths, line: line}
}

func NewStatement(code string, line int) *Statement {
	return &Statement{Code: code, line: line}
}

func NewEcho(expr string, line int) *Echo {
	return &Echo{Expr: expr, line: line}
}

func NewIf(cond string, then []Node, elifs []ElseIf, els []Node, line int) *If {
	return &If{Cond: cond, Then: then, Elifs: elifs, Else: els, line: line}
}

func NewSwitch(expr string, cases []Case, def []Node, line int) *Switch {
	return &Switch{Expr: expr, Cases: cases, Default: def, line: line}
}

func NewFor(expr string, body []Node, line int) *For {
	return &For{Expr: expr, Body: body, line: line}
}
