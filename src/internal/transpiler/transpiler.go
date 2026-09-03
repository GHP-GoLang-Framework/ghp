// Package transpiler converts one .ghp page into its generated Go file:
// it reads the source (router), parses it (parser) and emits the handler
// function (codegen), writing the .go file next to the .ghp origin.
package transpiler

import (
	"fmt"
	"os"
	"path/filepath"

	"ghp/src/internal/parser"
	"ghp/src/internal/transpiler/codegen"
	"ghp/src/internal/transpiler/router"
)

// Transpile converts the .ghp page described by ghpFile into its
// generated .go file: it reads the page from src and writes the handler to
// out. Ex: src /srv/pages, out /tmp/ghp-x, blog/about.ghp ->
// /tmp/ghp-x/blog/about.go.
func Transpile(ghpFile router.GhpFile, src string, out string) error {
	content, err := os.ReadFile(ghpFile.Ghp(src))
	if err != nil {
		return fmt.Errorf("read %s: %w", ghpFile.Ghp(src), err)
	}

	program, err := parser.Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse %s: %w", ghpFile.Ghp(src), err)
	}

	goSrc, err := codegen.Assemble(ghpFile.PkgName, ghpFile.FuncName, program)
	if err != nil {
		return fmt.Errorf("codegen %s: %w", ghpFile.Ghp(src), err)
	}

	goPath := ghpFile.Go(out)
	if err := os.MkdirAll(filepath.Dir(goPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(goPath), err)
	}
	if err := os.WriteFile(goPath, []byte(goSrc), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", goPath, err)
	}
	return nil
}
