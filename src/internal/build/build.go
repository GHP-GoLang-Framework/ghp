// Package build implements the ghp build pipeline: it stages a project
// into a temp dir, transpiles every .ghp page, emits ghproutes.go into both
// trees and compiles the main package, dropping the binary at the requested
// output path.
package build

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ghp/src/internal/colors"
	"ghp/src/internal/pages"
	"ghp/src/internal/transpiler"
)

// binaryName is the default name of the compiled server when no output is
// given. Ex: /srv/pages -> /srv/pages/app.
const binaryName = "app"

// Project stages, transpiles and compiles the project at src in one shot and
// removes the staged dir when it returns: pages are discovered from src,
// handlers are generated next to them, ghproutes.go is written into the entry
// dir (the dir of the main package) and the binary lands at output. Empty
// entry defaults to src, empty output to <src>/app.
// Ex: Project("/srv/pages", "/srv/pages/cli", "/srv/bin/ghp-server") ->
// /srv/pages/cli/ghproutes.go, /srv/bin/ghp-server.
func Project(src, entry, output string) error {
	tmpDir, err := Stage(src)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if entry == "" {
		entry = src
	}
	if output == "" {
		output = filepath.Join(src, binaryName)
	}
	return Compile(src, tmpDir, entry, output)
}

// Compile runs the transpile + routes + compile pipeline inside an already
// staged dir so callers (dev/build) own the staged lifetime. Pages are
// discovered project-root-wide from stagedDir and their handlers generated
// next to them; ghproutes.go is written into the main package dir (entry)
// `go build .` runs there too, so a main package in a subdir works like one
// at the root. The binary lands at output. entry must sit inside src.
// Ex: Compile("/srv/pages", "/tmp/ghp-stage", "/srv/pages/src",
// "/srv/bin/ghp-server") -> /srv/pages/src/ghproutes.go, /srv/bin/ghp-server.
func Compile(src, stagedDir, entry, output string) error {
	entryRel, err := filepath.Rel(src, entry)
	if err != nil {
		return err
	}
	if entryRel == ".." || strings.HasPrefix(entryRel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("entry point %q is outside the project %q", entry, src)
	}

	projectPages := pages.Discover(stagedDir)
	// Pages under the entry dir belong to the main package: Go allows a
	// single package per dir, so entry-top pages join it (PkgName main) and
	// subdirs keep their own package while the route sheds the entry prefix.
	if entryRel != "." {
		for _, p := range projectPages {
			switch {
			case p.RelDir == entryRel:
				p.PkgName = "main"
				p.Route = pages.NewPage(p.FileName + ".ghp").Route
			case strings.HasPrefix(p.RelDir, entryRel+string(filepath.Separator)):
				rest := strings.TrimPrefix(p.RelDir, entryRel+string(filepath.Separator))
				p.Route = pages.NewPage(rest + string(filepath.Separator) + p.FileName + ".ghp").Route
			}
		}
	}

	failed := false
	fmt.Println("Project: ", colors.Magenta(stagedDir))
	for _, p := range projectPages {
		if err := transpiler.Transpile(p, stagedDir, stagedDir); err != nil {
			fmt.Fprintf(os.Stderr, "ghp: %s\n", err)
			failed = true
		}
	}

	buildDir := stagedDir
	if entryRel != "." {
		buildDir = filepath.Join(stagedDir, entryRel)
	}

	module := readModule(filepath.Join(stagedDir, "go.mod"))
	routesPath := filepath.Join(entry, "ghproutes.go")
	if err := writeRoutes(projectPages, module, filepath.Join(buildDir, "ghproutes.go"), routesPath); err != nil {
		return err
	}

	if failed {
		return fmt.Errorf("some pages failed to transpile; skipping binary build")
	}
	return compile(buildDir, output)
}

// RemoveStaged keeps the staged dir on disk, tells the user where it is and
// blocks until Enter is pressed, then removes it. When stdin ends (EOF, e.g.
// in tests or scripts) the dir is removed right away.
// Ex: staged dir /tmp/ghp-stage-123 -> "press Enter to delete
// /tmp/ghp-stage-123".
func RemoveStaged(tmpDir string) {
	fmt.Printf("Staged build in %s — press Enter to delete: ", tmpDir)
	reader := bufio.NewReader(os.Stdin)
	if _, err := reader.ReadString('\n'); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "ghp: reading stdin: %s\n", err)
	}
	if err := os.RemoveAll(tmpDir); err != nil {
		fmt.Fprintf(os.Stderr, "ghp: removing staged dir: %s\n", err)
	}
}

// Stage copies the buildable set of src into a fresh temp dir: every .ghp
// page, every .go file and the root go.mod, preserving their relative
// layout. Only directories that contain such files are recreated. Ex: src
// blog/about.ghp -> tmp/blog/about.ghp.
func Stage(src string) (string, error) {
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

// compile builds the main package of the staged module in tmpDir (the
// project root, `go build .`) and writes the binary to output, creating the
// output dir when needed. Ex: tmpDir /tmp/ghp-stage, output /out/app ->
// /out/app.
func compile(tmpDir, output string) error {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", output, ".")
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
