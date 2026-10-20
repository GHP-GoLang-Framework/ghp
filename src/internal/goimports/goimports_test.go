package goimports

import (
	"reflect"
	"strings"
	"testing"
)

// TestBlock covers the codegen shape: always a parenthesized block, sorted
// by path, with the alias between the tab and the quoted path, followed by
// a blank line. The one-liner is never used here so the bytes a page's
// handler has always produced stay the same.
func TestBlock(t *testing.T) {
	tests := []struct {
		name  string
		specs []Import
		want  string
	}{
		{
			name:  "single spec is still a block",
			specs: []Import{{Path: "net/http"}},
			want:  "import (\n\t\"net/http\"\n)\n\n",
		},
		{
			name:  "two specs",
			specs: []Import{{Path: "net/http"}, {Path: "a/b"}},
			want:  "import (\n\t\"a/b\"\n\t\"net/http\"\n)\n\n",
		},
		{
			name:  "specs are sorted by path regardless of input order",
			specs: []Import{{Path: "z/z"}, {Path: "a/a"}, {Path: "m/m"}},
			want:  "import (\n\t\"a/a\"\n\t\"m/m\"\n\t\"z/z\"\n)\n\n",
		},
		{
			name:  "alias is written before the path",
			specs: []Import{{Alias: "s", Path: "strings"}, {Path: "net/http"}},
			want:  "import (\n\t\"net/http\"\n\ts \"strings\"\n)\n\n",
		},
		{
			name:  "single aliased spec keeps the alias",
			specs: []Import{{Alias: "s", Path: "strings"}},
			want:  "import (\n\ts \"strings\"\n)\n\n",
		},
		{
			name:  "empty input is an empty block",
			specs: nil,
			want:  "import (\n)\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Block(tt.specs); got != tt.want {
				t.Errorf("Block() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCompact covers the ghproutes.go shape: a lone spec collapses to the
// one-line form and there is no trailing blank line, because the result is
// spliced into a template rather than handed to go/format.
func TestCompact(t *testing.T) {
	tests := []struct {
		name  string
		specs []Import
		want  string
	}{
		{
			name:  "single spec uses the one-line form",
			specs: []Import{{Path: "net/http"}},
			want:  `import "net/http"`,
		},
		{
			name:  "single aliased spec keeps the alias in the one-line form",
			specs: []Import{{Alias: "s", Path: "strings"}},
			want:  `import s "strings"`,
		},
		{
			name:  "two specs use the block form with no trailing blank line",
			specs: []Import{{Path: "net/http"}, {Path: "a/b"}},
			want:  "import (\n\t\"a/b\"\n\t\"net/http\"\n)",
		},
		{
			name:  "specs are sorted by path",
			specs: []Import{{Path: "z/z"}, {Path: "a/a"}, {Path: "m/m"}},
			want:  "import (\n\t\"a/a\"\n\t\"m/m\"\n\t\"z/z\"\n)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compact(tt.specs); got != tt.want {
				t.Errorf("Compact() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBlockAndCompactAgreeOnMultiSpecs pins the shared core: for two or
// more specs the two shapes differ only by the trailing newline, so the
// dedup that moved them into this package really is one implementation.
func TestBlockAndCompactAgreeOnMultiSpecs(t *testing.T) {
	specs := []Import{
		{Path: "net/http"},
		{Alias: "f", Path: "fmt"},
		{Path: "io"},
	}

	block := Block(specs)
	compact := Compact(specs)

	if block != compact+"\n\n" {
		t.Errorf("Block() = %q, Compact() = %q; want Block to be Compact plus the trailing blank line", block, compact)
	}
}

// TestSpecsAreNotMutated is the guarantee that lets both generators hand
// over their own slice: rendering must not reorder what the caller still
// owns.
func TestSpecsAreNotMutated(t *testing.T) {
	specs := []Import{{Path: "z/z"}, {Path: "a/a"}}
	before := make([]Import, len(specs))
	copy(before, specs)

	Block(specs)
	Compact(specs)

	if !reflect.DeepEqual(specs, before) {
		t.Errorf("input mutated: got %v, want %v", specs, before)
	}
}

// TestRenderingIsDeterministic pins the property the generators rely on:
// the same set always renders the same bytes, so a rebuild never produces a
// spurious diff.
func TestRenderingIsDeterministic(t *testing.T) {
	specs := []Import{
		{Path: "net/http"},
		{Alias: "f", Path: "fmt"},
		{Path: "io"},
	}

	firstBlock, firstCompact := Block(specs), Compact(specs)
	for i := 0; i < 50; i++ {
		if got := Block(specs); got != firstBlock {
			t.Fatalf("Block call %d = %q, want %q", i, got, firstBlock)
		}
		if got := Compact(specs); got != firstCompact {
			t.Fatalf("Compact call %d = %q, want %q", i, got, firstCompact)
		}
	}
}

// TestUniquePathsAppearOnce guards the contract that the block never
// invents a duplicate line for a path that appears once, which is what the
// generators promise after their own dedup.
func TestUniquePathsAppearOnce(t *testing.T) {
	got := Block([]Import{
		{Path: "net/http"},
		{Path: "fmt"},
		{Path: "io"},
	})

	if n := strings.Count(got, `"net/http"`); n != 1 {
		t.Errorf("net/http appears %d times, want 1:\n%s", n, got)
	}
}

// TestCompactSinglePathQuotesExactly checks the quoting of the one-line
// form, which is the one thing the block form and the compact form do not
// share and so the easiest to get subtly wrong.
func TestCompactSinglePathQuotesExactly(t *testing.T) {
	got := Compact([]Import{{Path: "github.com/acme/site/internal/db"}})

	want := `import "github.com/acme/site/internal/db"`
	if got != want {
		t.Errorf("Compact() = %q, want %q", got, want)
	}
}
