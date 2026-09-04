package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ghp/src/internal/colors"
	"ghp/src/internal/transpiler"
	"ghp/src/internal/transpiler/router"
)

func Build(args []string) int {
	src, err := getPath(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp build: %v\n", err)
		return 1
	}

	tmpDir, err := transcribe(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp build: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	status := _build(tmpDir, tmpDir)
	if err := writeRoutes(tmpDir, filepath.Join(src, "ghproutes.go")); err != nil {
		fmt.Fprintf(os.Stderr, "ghp build: %v\n", err)
		status = 1
	}
	if status == 0 {
		if err := buildBinary(tmpDir, src, "app"); err != nil {
			fmt.Fprintf(os.Stderr, "ghp build: %v\n", err)
			status = 1
		}
	}
	return status
}

func Dev(args []string) int {
	return Build(args)
}

// _build transpiles every .ghp page under src, printing a clear error that names the failing page for each one that fails, while still transpiling the rest of the pages. The generated .go files land in out. It returns a non-zero exit code when any page failed. Ex: "ghp build: blog/index.ghp: parse line 3: ...".
func _build(src string, out string) int {
	fmt.Println("Project: ", colors.Magenta(src))
	status := 0
	for _, f := range router.SearchGhpFiles(src) {
		if err := transpiler.Transpile(f, src, out); err != nil {
			fmt.Fprintf(os.Stderr, "ghp build: %s\n", err)
			status = 1
		}
	}
	return status
}

func getPath(args []string) (string, error) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve build path: %w", err)
	}

	if _, err := os.Stat(filepath.Join(path, "go.mod")); err != nil {
		return "", fmt.Errorf("go.mod not found in %s: %w", path, err)
	}

	if _, err := os.Stat(filepath.Join(path, "main.go")); err != nil {
		return "", fmt.Errorf("main.go not found in %s: %w", path, err)
	}

	return path, nil
}

// transcribe stages the buildable set of src into a fresh temp dir: every
// .ghp page, every .go file and the root go.mod, preserving their relative
// layout. Only directories that contain such files are recreated. Ex:
// src blog/about.ghp -> tmp/blog/about.ghp.
func transcribe(src string) (string, error) {
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

// buildBinary compiles the main package of the staged module in tmpDir and
// places the resulting executable name inside src. Ex: tmp /tmp/ghp, src
// /srv/pages, name "app" -> /srv/pages/app.
func buildBinary(tmpDir string, src string, name string) error {
	cmd := exec.Command("go", "build", "-o", filepath.Join(src, name), ".")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build: %w\n%s", err, out)
	}
	return nil
}

// writeRoutes generates ghproutes.go from the staged pages and writes it to
// the temp dir (where the build lives) and to routesPath, the route file the
// user's tree keeps. Ex: tmp /tmp/ghp, routesPath /srv/pages/ghproutes.go ->
// /tmp/ghp/ghproutes.go and /srv/pages/ghproutes.go.
func writeRoutes(tmpDir string, routesPath string) error {
	content := router.GenRoutes(router.SearchGhpFiles(tmpDir), modulePath(filepath.Join(tmpDir, "go.mod")))
	for _, f := range []string{filepath.Join(tmpDir, "ghproutes.go"), routesPath} {
		if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// modulePath returns the module directive of a go.mod file, or "" when none
// is declared. Ex: "module example.com/site\n" -> "example.com/site".
func modulePath(gomod string) string {
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
