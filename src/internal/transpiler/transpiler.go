// Package transpiler converts one .ghp page into its generated Go file:
// it reads the source (pages), parses it (parser) and emits the handler
// function (codegen), writing the .go file next to the .ghp origin.
package transpiler

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/term"

	"ghp/src/internal/colors"
	"ghp/src/internal/pages"
	"ghp/src/internal/parser"
	"ghp/src/internal/transpiler/codegen"
)

// Transpile converts the .ghp page described by page into its generated
// .go file: it reads the page from src and writes the handler to out.
// Ex: src /srv/pages, out /tmp/ghp-x, blog/about.ghp ->
// /tmp/ghp-x/blog/about.go.
func Transpile(page *pages.Page, src string, out string) error {
	err := transpile(page, src, out)
	name := colors.Yellow(page.FileName + ".ghp")
	if page.RelDir != "" {
		name = page.RelDir + "/" + name
	}

	status := ""
	if err != nil {
		status = colors.Red("FAIL")
	} else {
		status = colors.Green("OK")
	}

	line := "  " + colors.Cyan(name)
	fmt.Printf("%s%s\n", colors.Dots(line, statusCol(width())), "["+status+"]")
	return err
}

// width returns the terminal's column count, falling back to 80 when it
// cannot be determined (e.g. output redirected).
func width() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// statusCol places the status at roughly 80% of the terminal width.
func statusCol(w int) int {
	return w * 8 / 10
}

func transpile(page *pages.Page, src string, out string) error {
	content, err := os.ReadFile(page.Ghp(src))
	if err != nil {
		return fmt.Errorf("read %s: %w", page.Ghp(src), err)
	}

	program, err := parser.Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse %s: %w", page.Ghp(src), err)
	}

	goSrc, err := codegen.Assemble(page.PkgName, page.FuncName, program)
	if err != nil {
		return fmt.Errorf("codegen %s: %w", page.Ghp(src), err)
	}

	goPath := page.Go(out)
	if err := os.MkdirAll(filepath.Dir(goPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(goPath), err)
	}

	if err := os.WriteFile(goPath, []byte(goSrc), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", goPath, err)
	}
	return nil
}
