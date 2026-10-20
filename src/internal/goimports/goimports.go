// Package goimports renders the import declaration of a generated Go file.
//
// Two different generators need the same block: codegen writes the imports
// of one page's handler, and pages writes the imports of ghproutes.go.
// They live in sibling packages that cannot import each other, so the
// shared type and renderer live here, in a leaf package both may use.
//
// Two shapes exist because the two callers disagree on how a lone import
// should look: ghproutes.go is written verbatim and reads better as
// `import "net/http"`, while a page's handler is always passed through
// go/format, which keeps whichever shape it is given. Neither caller may
// silently change the bytes the other produces, so both shapes are
// explicit rather than one being inferred from the spec count.
package goimports

import (
	"fmt"
	"sort"
	"strings"
)

// Import is one package inside a <go:import (...)/> tag: the path and the
// optional alias the page declared (e.g. `f "fmt"`).
type Import struct {
	Alias string
	Path  string
}

// Block renders the parenthesized import declaration for specs, one path
// per line in sorted order, followed by a blank line. A single spec is
// still written as a block, so a caller that hands its output to go/format
// gets the same bytes it always has. Ex:
// [{Path: "a/b"}, {Path: "net/http"}] ->
// "import (\n\t\"a/b\"\n\t\"net/http\"\n)\n\n".
func Block(specs []Import) string {
	var b strings.Builder
	writeBlock(&b, sortedSpecs(specs))
	b.WriteString("\n")
	return b.String()
}

// Compact renders the import declaration for specs in the form a
// hand-maintained file would use: a single spec becomes the one-line
// `import "net/http"`, several become the parenthesized block with no
// trailing blank line. Ex: [{Path: "net/http"}] -> `import "net/http"`.
func Compact(specs []Import) string {
	sorted := sortedSpecs(specs)

	if len(sorted) == 1 {
		return oneLine(sorted[0])
	}

	var b strings.Builder
	writeBlock(&b, sorted)
	return strings.TrimSuffix(b.String(), "\n")
}

// oneLine renders a single import without parentheses. The alias, when
// there is one, still comes out: dropping it would leave page code that
// refers to the alias compiled against the package's default name.
func oneLine(spec Import) string {
	if spec.Alias != "" {
		return fmt.Sprintf("import %s %q", spec.Alias, spec.Path)
	}
	return fmt.Sprintf("import %q", spec.Path)
}

// writeBlock writes the parenthesized form of sorted into b, ending in a
// newline. Callers own whatever spacing follows.
func writeBlock(b *strings.Builder, sorted []Import) {
	b.WriteString("import (\n")
	for _, spec := range sorted {
		if spec.Alias != "" {
			fmt.Fprintf(b, "\t%s %q\n", spec.Alias, spec.Path)
		} else {
			fmt.Fprintf(b, "\t%q\n", spec.Path)
		}
	}
	b.WriteString(")\n")
}

// sortedSpecs copies specs and orders them by path, without mutating the
// caller's slice. Ordering by path is what makes the same input always
// produce the same block; duplicates are the caller's responsibility to
// resolve, since only it knows whether two entries are a clash or a repeat.
func sortedSpecs(specs []Import) []Import {
	sorted := make([]Import, len(specs))
	copy(sorted, specs)

	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	return sorted
}
