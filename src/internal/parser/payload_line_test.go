package parser

import (
	"testing"

	"ghp/src/internal/ast"
)

// TestPayloadStartLine pins the line a node reports to the line where its
// payload begins, not the line of the tag head: //line directives built from
// it must match the .ghp line of the Go code, or debugger breakpoints inside
// multi-line blocks never match.
func TestPayloadStartLine(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want int
	}{
		{"single-line statement", "<go x := 1 />", 1},
		{"statement starting on next line", "<go\n\tx := 1\n/>", 2},
		{"blank lines before the code", "<go\n\n\tx := 1\n/>", 3},
		{"CRLF before the code", "<go\r\n\tx := 1\r\n/>", 2},
		{"after preceding text", "a\nb\n<go\n\tx := 1\n/>", 4},
		{"statement on the same line as the head", "<go x := 1\n\ty := 2\n/>", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, err := Parse(tt.src)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			var stmt *ast.Statement
			for _, n := range prog.Nodes {
				if s, ok := n.(*ast.Statement); ok {
					stmt = s
				}
			}
			if stmt == nil {
				t.Fatalf("no Statement node in %v", prog.Nodes)
			}
			if got := stmt.Line(); got != tt.want {
				t.Errorf("Line() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPayloadStartLineEmptyPayload(t *testing.T) {
	prog, err := Parse("<go:if x/>\n<go:else\n/>\n<go:endif/>")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ifNode, ok := prog.Nodes[0].(*ast.If)
	if !ok {
		t.Fatalf("node type = %T, want *ast.If", prog.Nodes[0])
	}
	if got := ifNode.Line(); got != 1 {
		t.Errorf("If.Line() = %d, want 1", got)
	}
}
