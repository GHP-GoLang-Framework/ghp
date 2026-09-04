package codegen

import (
	"fmt"
	"go/format"
	"strings"

	"ghp/src/internal/ast"
)

// Assemble builds a complete, compilable .go source file for one page:
// a package declaration, only the imports the page actually needs, and
// an http.HandlerFunc-shaped function (named funcName) whose body comes
// from Generate.
//
// The result is run through go/format before being returned - both for
// idiomatic formatting and as a cheap correctness check: if the assembled
// source isn't valid Go (e.g. funcName isn't a valid identifier), this
// reports that as an error here instead of handing back something that
// would only fail later, confusingly, at `go build`. That check only
// catches syntax problems, not semantic ones (an undefined identifier
// still parses fine) - the import-handling errors below exist because
// that gap is real.
func Assemble(pkg, funcName string, prog *ast.Program) (string, error) {
	body, err := Generate(prog.Nodes)
	if err != nil {
		return "", err
	}

	var need neededImports
	scanNeededImports(prog.Nodes, &need)

	userImports, err := collectImports(prog.Nodes)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	if err := writeImports(&b, need, userImports); err != nil {
		return "", err
	}
	fmt.Fprintf(&b, "func %s(w http.ResponseWriter, r *http.Request) {\n", funcName)
	b.WriteString(body)
	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", fmt.Errorf("codegen: assembled file is not valid Go: %w", err)
	}
	return string(formatted), nil
}

// neededImports tracks which of the packages genEcho/genText depend on
// are actually exercised somewhere in the page - importing fmt/html/io
// unconditionally would fail to compile on a page with no <go=...> or
// plain text anywhere in it (an unlikely but possible page made up only
// of <go ...>/<go:for>/etc).
type neededImports struct {
	io, html, fmt bool
}

// scanNeededImports walks nodes (recursing into every tag with a body)
// looking for *ast.Text and *ast.Echo, and sets the corresponding fields
// on need. It never returns early: a single pass has to see the whole
// tree, since any branch might be the only one using io/html/fmt.
func scanNeededImports(nodes []ast.Node, need *neededImports) {
	for _, n := range nodes {
		switch node := n.(type) {
		case *ast.Text:
			// A Text that is only whitespace emits no io.WriteString (see
			// genText), so it doesn't need the io import either.
			if strings.TrimSpace(node.Value) != "" {
				need.io = true
			}
		case *ast.Echo:
			need.io, need.html, need.fmt = true, true, true
		case *ast.If:
			scanNeededImports(node.Then, need)
			for i := range node.Elifs {
				scanNeededImports(node.Elifs[i].Body, need)
			}
			scanNeededImports(node.Else, need)
		case *ast.Switch:
			for _, c := range node.Cases {
				scanNeededImports(c.Body, need)
			}
			scanNeededImports(node.Default, need)
		case *ast.For:
			scanNeededImports(node.Body, need)
		}
	}
}

// collectImports gathers every ast.ImportPath declared via <go:import> at
// the top level of nodes, deduplicated by Path. Two <go:import> tags
// declaring the same path with different aliases is an error - there's
// no way to tell which alias the page's own code actually expects to
// use, so silently keeping one would risk leaving the other's references
// pointing at an identifier that was never imported.
//
// <go:import> only makes sense at the top level of a file - Go doesn't
// have conditional imports - but the parser doesn't reject one nested
// inside a <go:if>/<go:switch>/<go:for> body, so this also checks for
// that and reports it as an error instead of silently ignoring it (which
// is what happens to any *ast.Import generateNode is asked to render:
// see its own comment).
func collectImports(nodes []ast.Node) ([]ast.ImportPath, error) {
	byPath := make(map[string]ast.ImportPath)
	var paths []ast.ImportPath

	for _, n := range nodes {
		imp, ok := n.(*ast.Import)
		if !ok {
			continue
		}
		for _, p := range imp.Paths {
			if existing, ok := byPath[p.Path]; ok {
				if existing.Alias != p.Alias {
					return nil, fmt.Errorf("codegen: %q imported with different aliases (%q and %q)", p.Path, existing.Alias, p.Alias)
				}
				continue
			}
			byPath[p.Path] = p
			paths = append(paths, p)
		}
	}

	if nested := nestedImport(nodes); nested != nil {
		return nil, fmt.Errorf("codegen: <go:import> on line %d is not at the top level of the file - conditional imports do not exist in Go", nested.Line())
	}

	return paths, nil
}

// nestedImport looks for an *ast.Import inside any <go:if>/<go:switch>/
// <go:for> body reachable from nodes. It deliberately only looks inside
// those bodies, never at nodes' own top-level entries - collectImports
// already handles those separately, and a top-level *ast.Import is
// exactly where one belongs.
func nestedImport(nodes []ast.Node) *ast.Import {
	for _, n := range nodes {
		var bodies [][]ast.Node
		switch node := n.(type) {
		case *ast.If:
			bodies = [][]ast.Node{node.Then}
			for i := range node.Elifs {
				bodies = append(bodies, node.Elifs[i].Body)
			}
			bodies = append(bodies, node.Else)
		case *ast.Switch:
			for _, c := range node.Cases {
				bodies = append(bodies, c.Body)
			}
			bodies = append(bodies, node.Default)
		case *ast.For:
			bodies = [][]ast.Node{node.Body}
		default:
			continue
		}
		for _, body := range bodies {
			for _, bn := range body {
				if imp, ok := bn.(*ast.Import); ok {
					return imp
				}
			}
			if imp := nestedImport(body); imp != nil {
				return imp
			}
		}
	}
	return nil
}

// writeImports writes the import(...) block: net/http always (the
// handler signature needs it), fmt/html/io only when need says the page
// actually uses them, then every path collected from the page's own
// <go:import> tags.
//
// A user import whose path matches one of the automatic ones is skipped
// if it has no alias (genuinely redundant, e.g. the page also explicitly
// wrote <go:import ("io")/>). But it can't be honored if it does have an
// alias: genText/genEcho's generated calls always use the default name
// (io.WriteString, fmt.Sprint, html.EscapeString), so a page that aliases
// one of those paths would end up with code calling a name nothing
// imports - this is reported as an error instead of silently dropping
// the alias.
func writeImports(b *strings.Builder, need neededImports, userImports []ast.ImportPath) error {
	auto := []string{"net/http"}
	if need.fmt {
		auto = append(auto, "fmt")
	}
	if need.html {
		auto = append(auto, "html")
	}
	if need.io {
		auto = append(auto, "io")
	}

	b.WriteString("import (\n")
	autoSet := make(map[string]bool, len(auto))
	for _, path := range auto {
		autoSet[path] = true
		fmt.Fprintf(b, "\t%q\n", path)
	}

	for _, p := range userImports {
		if autoSet[p.Path] {
			if p.Alias != "" {
				return fmt.Errorf("codegen: %q is managed automatically by the page and cannot be imported with an alias (%q)", p.Path, p.Alias)
			}
			continue
		}
		if p.Alias != "" {
			fmt.Fprintf(b, "\t%s %q\n", p.Alias, p.Path)
		} else {
			fmt.Fprintf(b, "\t%q\n", p.Path)
		}
	}
	b.WriteString(")\n\n")
	return nil
}
