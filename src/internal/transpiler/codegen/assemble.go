package codegen

import (
	"fmt"
	"go/format"
	"sort"
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
// still parses fine).
func Assemble(pkg, funcName string, prog *ast.Program) (string, error) {
	body, err := Generate(prog.Nodes)
	if err != nil {
		return "", err
	}

	imports, err := importsOf(prog)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	writeImports(&b, imports)
	fmt.Fprintf(&b, "func %s(w http.ResponseWriter, r *http.Request) {\n", funcName)
	b.WriteString(body)
	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", fmt.Errorf("codegen: assembled file is not valid Go: %w", err)
	}
	return string(formatted), nil
}

// importsOf returns the package set Assemble has to write out. Programs
// that came through parser.Parse carry it ready (ast.Program.Imports) -
// the parser is the only one walking the whole tree anyway, so it is also
// the one place that knows which packages the page ended up needing.
// A hand-built program may not have it, so fall back to deriving it here.
func importsOf(prog *ast.Program) (map[string]ast.ImportPath, error) {
	if prog.Imports != nil {
		return prog.Imports, nil
	}
	return ast.CollectImports(prog.Nodes)
}

// writeImports writes the import(...) block, one entry per path in
// imports, with the alias when the page declared one. The keys are walked
// in sorted order so the same tree always yields the same intermediate
// source; go/format sorts the block again anyway before Assemble returns.
func writeImports(b *strings.Builder, imports map[string]ast.ImportPath) {
	paths := make([]string, 0, len(imports))
	for path := range imports {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	b.WriteString("import (\n")
	for _, path := range paths {
		if alias := imports[path].Alias; alias != "" {
			fmt.Fprintf(b, "\t%s %q\n", alias, path)
		} else {
			fmt.Fprintf(b, "\t%q\n", path)
		}
	}
	b.WriteString(")\n\n")
}
