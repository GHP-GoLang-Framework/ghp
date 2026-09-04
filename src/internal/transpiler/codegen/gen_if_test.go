package codegen

import (
	"strings"
	"testing"

	"ghp/src/internal/ast"
)

func TestGenIfWithoutElse(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("a == b", []ast.Node{ast.NewText("yes", 2)}, nil, nil, 1))
	if err != nil {
		t.Fatalf("genIf(): %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("genIf() = %q, want %q", got, want)
	}
}

func TestGenIfWithElse(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("a == b",
		[]ast.Node{ast.NewText("yes", 2)},
		nil,
		[]ast.Node{ast.NewText("no", 3)},
		1))
	if err != nil {
		t.Fatalf("genIf(): %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"} else {\n" +
		"io.WriteString(w, \"no\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("genIf() = %q, want %q", got, want)
	}
}

func TestGenIfWithElif(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("a == b",
		[]ast.Node{ast.NewText("yes", 2)},
		[]ast.ElseIf{{Cond: "c == d", Body: []ast.Node{ast.NewText("maybe", 3)}, Line: 3}},
		nil, 1))
	if err != nil {
		t.Fatalf("genIf(): %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"} else if c == d {\n" +
		"io.WriteString(w, \"maybe\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("genIf() = %q, want %q", got, want)
	}
}

func TestGenIfWithElifAndElse(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("a == b",
		[]ast.Node{ast.NewText("yes", 2)},
		[]ast.ElseIf{{Cond: "c == d", Body: []ast.Node{ast.NewText("maybe", 3)}, Line: 3}},
		[]ast.Node{ast.NewText("no", 4)},
		1))
	if err != nil {
		t.Fatalf("genIf(): %v", err)
	}

	want := "if a == b {\n" +
		"io.WriteString(w, \"yes\")\n" +
		"} else if c == d {\n" +
		"io.WriteString(w, \"maybe\")\n" +
		"} else {\n" +
		"io.WriteString(w, \"no\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("genIf() = %q, want %q", got, want)
	}
}

func TestGenIfWithMultipleElif(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("a",
		[]ast.Node{ast.NewText("x", 2)},
		[]ast.ElseIf{
			{Cond: "b", Body: []ast.Node{ast.NewText("y", 3)}, Line: 3},
			{Cond: "c", Body: []ast.Node{ast.NewText("z", 4)}, Line: 4},
		},
		nil, 1))
	if err != nil {
		t.Fatalf("genIf(): %v", err)
	}

	want := "if a {\n" +
		"io.WriteString(w, \"x\")\n" +
		"} else if b {\n" +
		"io.WriteString(w, \"y\")\n" +
		"} else if c {\n" +
		"io.WriteString(w, \"z\")\n" +
		"}\n"
	if got := b.String(); got != want {
		t.Errorf("genIf() = %q, want %q", got, want)
	}
}

func TestGenIfThenBodyError(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("true", []ast.Node{nil}, nil, nil, 1))
	if err == nil {
		t.Fatal("genIf() = nil error, want error from generateNodes")
	}
}

func TestGenIfElseBodyError(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("true", nil, nil, []ast.Node{nil}, 1))
	if err == nil {
		t.Fatal("genIf() = nil error, want error from generateNodes")
	}
}

func TestGenIfElifBodyError(t *testing.T) {
	var b strings.Builder
	err := genIf(&b, ast.NewIf("true", nil, []ast.ElseIf{{Cond: "false", Body: []ast.Node{nil}, Line: 1}}, nil, 1))
	if err == nil {
		t.Fatal("genIf() = nil error, want error from generateNodes")
	}
}
