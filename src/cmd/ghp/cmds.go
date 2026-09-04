package main

import (
	"fmt"
	"os"
	"path/filepath"

	"ghp/src/internal/colors"
	"ghp/src/internal/fsutil"
	"ghp/src/internal/transpiler"
	"ghp/src/internal/transpiler/router"
)

func Build(args []string) int {
	src, err := path(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp build: %v\n", err)
		return 1
	}

	tmpDir, err := os.MkdirTemp("", "ghp-build-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp build: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	return _build(src, tmpDir)
}

func Dev(args []string) int {
	src, err := path(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp dev: %v\n", err)
		return 1
	}

	tmpDir, err := os.MkdirTemp("", "ghp-dev-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp dev: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)

	ghps := []router.GhpFile{}
	err = fsutil.CopyTree(src, tmpDir, func(path string, isDir bool) {
		if isDir {
			// watch
		} else if filepath.Ext(path) == ".ghp" {
			rel, err := filepath.Rel(src, path)
			if err == nil {
				ghps = append(ghps, router.NewGhpFile(filepath.ToSlash(rel)))
			}
			// watch
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ghp dev: %v\n", err)
		return 1
	}

	return _build(src, tmpDir)
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

func path(args []string) (string, error) {
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
	return path, nil
}
