// Package build implements the ghp build pipeline: it stages a project
// into a temp dir, transpiles every .ghp page, emits ghproutes.go into
// both trees and compiles the main package, dropping the binary on the
// project root.
package build

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ghp/src/internal/colors"
	"ghp/src/internal/pages"
	"ghp/src/internal/transpiler"
)

// binaryName is the name of the compiled server dropped on the project
// root. Ex: /srv/pages -> /srv/pages/app.
const binaryName = "app"

// Project stages, transpiles and compiles the project at src. Stages it in
// a temp dir, transpiles every page (reporting failures inline without
// stopping the rest), writes ghproutes.go into the temp tree and the user's
// tree, then compiles the main package. Any failed page skips the binary
// step. Ex: src /srv/pages -> /srv/pages/ghproutes.go, /srv/pages/app.
func Project(src string) error {
	tmpDir, err := stage(src)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	projectPages := pages.Discover(tmpDir)

	failed := false
	fmt.Println("Project: ", colors.Magenta(tmpDir))
	for _, p := range projectPages {
		if err := transpiler.Transpile(p, tmpDir, tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "ghp: %s\n", err)
			failed = true
		}
	}

	module := readModule(filepath.Join(tmpDir, "go.mod"))
	if err := writeRoutes(projectPages, module, filepath.Join(tmpDir, "ghproutes.go"), filepath.Join(src, "ghproutes.go")); err != nil {
		return err
	}

	if failed {
		return fmt.Errorf("some pages failed to transpile; skipping binary build")
	}

	return compile(tmpDir, src)
}

// stage copies the buildable set of src into a fresh temp dir: every .ghp
// page, every .go file and the root go.mod, preserving their relative
// layout. Only directories that contain such files are recreated. Ex: src
// blog/about.ghp -> tmp/blog/about.ghp.
func stage(src string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ghp-stage-*")
	if err != nil {
		return "", err
	}

	err = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if filepath.Ext(name) != ".ghp" && filepath.Ext(name) != ".go" && name != "go.mod" {
			return nil
		}

		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		out := filepath.Join(tmpDir, rel)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, data, 0o644)
	})
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", err
	}
	return tmpDir, nil
}

// compile builds the main package of the staged module in tmpDir and places
// the resulting executable on src. Ex: tmpDir /tmp/ghp-stage, src
// /srv/pages -> /srv/pages/app.
func compile(tmpDir string, src string) error {
	cmd := exec.Command("go", "build", "-o", filepath.Join(src, binaryName), ".")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build: %w\n%s", err, out)
	}
	return nil
}

// writeRoutes generates ghproutes.go from the staged pages and writes it to
// tmpRoutes (where the build lives) and to routesPath, the route file the
// user's tree keeps. Ex: tmpRoutes /tmp/ghp/ghproutes.go, routesPath
// /srv/pages/ghproutes.go -> both files carry the same generated source.
func writeRoutes(files []*pages.Page, module string, tmpRoutes string, routesPath string) error {
	content := pages.GenRoutes(files, module)
	for _, f := range []string{tmpRoutes, routesPath} {
		if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// readModule returns the module directive of a go.mod file, or "" when none
// is declared. Ex: "module example.com/site\n" -> "example.com/site".
func readModule(gomod string) string {
	data, err := os.ReadFile(gomod)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	return ""
}
