package codegen

import (
	"strconv"
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestWriteLine(t *testing.T) {
	tests := []struct {
		name string
		file string
		line int
		want string
	}{
		{"directive", "/p/a.ghp", 7, "//line /p/a.ghp:7\n"},
		{"no file", "", 7, ""},
		{"no line", "/p/a.ghp", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			writeLine(&b, tt.file, tt.line)
			if got := b.String(); got != tt.want {
				t.Errorf("writeLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAssembleFileDirectives(t *testing.T) {
	prog := &ast.Program{Nodes: []ast.Node{
		ast.NewStatement("x := 1", 2),
		ast.NewText("hi", 3),
		ast.NewEcho("x", 4),
		ast.NewIf("x > 0",
			[]ast.Node{ast.NewText("pos", 6)},
			[]ast.ElseIf{{Cond: "x < 0", Body: []ast.Node{ast.NewText("neg", 8)}, Line: 7}},
			[]ast.Node{ast.NewText("zero", 10)}, 5),
		ast.NewSwitch("x",
			[]ast.Case{{Value: "1", Body: []ast.Node{ast.NewText("one", 13)}, Line: 12}}, nil, 11),
		ast.NewFor("i := 0; i < 1; i++", []ast.Node{ast.NewText("loop", 15)}, 14),
	}}

	out, err := AssembleFile("pages", "Index", "/p/a.ghp", prog)
	if err != nil {
		t.Fatalf("AssembleFile: %v", err)
	}

	for _, line := range []int{2, 3, 4, 5, 6, 7, 8, 11, 12, 13, 14, 15} {
		want := "\n//line /p/a.ghp:" + strconv.Itoa(line) + "\n"
		if !strings.Contains(out, want) {
			t.Errorf("missing directive for line %d (at column 1) in:\n%s", line, out)
		}
	}

	plain, err := Assemble("pages", "Index", prog)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	if strings.Contains(plain, "//line") {
		t.Errorf("Assemble must not emit directives:\n%s", plain)
	}
}
