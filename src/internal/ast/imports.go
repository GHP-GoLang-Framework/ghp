package ast

import (
	"fmt"
	"strings"
)

// CollectImports returns every package the generated handler needs, keyed
// by import path: the managed ones plus every <go:import (...)/> the page
// declared. Registering the same path twice is fine as long as the alias
// matches (it is simply deduplicated); the same path under two different
// aliases is an error, since there is no way to tell which name the page's
// own code expects.
//
// A <go:import> nested inside a <go:if>/<go:switch>/<go:for> body is also
// an error: Go has no conditional imports, and dropping it silently (which
// is what rendering an ast.Import amounts to) would only break the page
// later, at `go build`.
func CollectImports(nodes []Node) (map[string]ImportPath, error) {
	needs := scanManagedNeeds(nodes)

	// net/http is what the handler signature is made of, so it is never
	// optional; fmt/html/io only show up when the tree actually emits
	// code calling them - importing them unconditionally would fail to
	// compile on a page that echoes and prints nothing.
	imports := map[string]ImportPath{"net/http": {Path: "net/http"}}
	managed := map[string]bool{"net/http": true}
	for _, path := range needs.paths() {
		imports[path] = ImportPath{Path: path}
		managed[path] = true
	}

	declared := declaredImports(nodes)
	if imp := nestedImport(nodes); imp != nil {
		return nil, fmt.Errorf("line %d: <go:import> is not at the top level of the file - conditional imports do not exist in Go", imp.Line())
	}

	first := make(map[string]aliasOf, len(declared))
	for _, d := range declared {
		// genText/genEcho always call the managed packages by their
		// default name, so an alias on top of one of them would leave
		// the alias with nothing to refer to.
		if managed[d.Path] && d.Alias != "" {
			return nil, fmt.Errorf("line %d: %q is managed automatically by the page and cannot be imported with an alias (%q)", d.line, d.Path, d.Alias)
		}
		if prev, ok := first[d.Path]; ok && prev.alias != d.Alias {
			return nil, fmt.Errorf("line %d: %q is imported with alias %q, but line %d already imports it with %q", d.line, d.Path, d.Alias, prev.line, prev.alias)
		}
		if _, ok := first[d.Path]; !ok {
			first[d.Path] = aliasOf{alias: d.Alias, line: d.line}
		}
		imports[d.Path] = d.ImportPath
	}

	return imports, nil
}

// aliasOf remembers the first <go:import> seen for a given path, so a
// later one with a different alias can be reported as the clash it is,
// pointing at both lines.
type aliasOf struct {
	alias string
	line  int
}

// declaredImport is one package from a <go:import (...)/> tag, together
// with the line that tag started on.
type declaredImport struct {
	ImportPath
	line int
}

// declaredImports collects the <go:import (...)/> packages sitting at the
// top level of nodes, in source order.
func declaredImports(nodes []Node) []declaredImport {
	var out []declaredImport
	for _, n := range nodes {
		imp, ok := n.(*Import)
		if !ok {
			continue
		}
		for _, p := range imp.Paths {
			out = append(out, declaredImport{ImportPath: p, line: imp.Line()})
		}
	}
	return out
}

// managedNeeds reports which of fmt/html/io the tree actually exercises.
// A Text that is only whitespace emits no io.WriteString (see genText), so
// it doesn't count; an Echo always uses all three.
func scanManagedNeeds(nodes []Node) managedNeeds {
	var needs managedNeeds
	Walk(nodes, func(n Node) {
		switch node := n.(type) {
		case *Text:
			if strings.TrimSpace(node.Value) != "" {
				needs.io = true
			}
		case *Echo:
			needs.fmt, needs.html, needs.io = true, true, true
		}
	})
	return needs
}

// managedNeeds is the set of automatic packages a tree pulls in, beyond
// the always-present net/http.
type managedNeeds struct {
	fmt, html, io bool
}

// paths lists the managed packages this tree actually uses, in a stable
// order.
func (m managedNeeds) paths() []string {
	var out []string
	if m.fmt {
		out = append(out, "fmt")
	}
	if m.html {
		out = append(out, "html")
	}
	if m.io {
		out = append(out, "io")
	}
	return out
}

// nestedImport returns the first *ast.Import sitting below the top level
// of nodes, if any - only looking inside tag bodies, since a top-level
// *ast.Import is exactly where one belongs.
func nestedImport(nodes []Node) *Import {
	for _, n := range nodes {
		var bodies [][]Node
		switch node := n.(type) {
		case *If:
			bodies = [][]Node{node.Then}
			for i := range node.Elifs {
				bodies = append(bodies, node.Elifs[i].Body)
			}
			bodies = append(bodies, node.Else)
		case *Switch:
			for i := range node.Cases {
				bodies = append(bodies, node.Cases[i].Body)
			}
			bodies = append(bodies, node.Default)
		case *For:
			bodies = [][]Node{node.Body}
		default:
			continue
		}
		for _, body := range bodies {
			for _, bn := range body {
				if imp, ok := bn.(*Import); ok {
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
