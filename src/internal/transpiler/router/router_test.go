package router

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writePages creates one empty file per path (relative to dir) under
// dir, creating any parent directories needed. Scan only looks at paths,
// never at file contents, so an empty file is enough to test it.
func writePages(t *testing.T, dir string, relPaths ...string) {
	t.Helper()
	for _, rel := range relPaths {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(full, nil, 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", full, err)
		}
	}
}

func TestScanDerivesPagesForEachFile(t *testing.T) {
	dir := t.TempDir()
	writePages(t, dir, "index.ghp", "about.ghp", "blog/[slug].ghp")

	pages, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(pages) != 3 {
		t.Fatalf("len(pages) = %d, want 3: %+v", len(pages), pages)
	}

	byGhpPath := make(map[string]Page)
	for _, p := range pages {
		byGhpPath[p.GhpPath] = p
	}

	if got := byGhpPath["index.ghp"]; got.Route != "/" || got.FuncName != "Index" {
		t.Errorf("index.ghp = %+v", got)
	}
	if got := byGhpPath["about.ghp"]; got.Route != "/about" || got.FuncName != "About" {
		t.Errorf("about.ghp = %+v", got)
	}
	if got := byGhpPath["blog/[slug].ghp"]; got.Route != "/blog/{slug}" || got.FuncName != "BlogSlug" {
		t.Errorf("blog/[slug].ghp = %+v", got)
	}
}

func TestScanIgnoresNonGhpFiles(t *testing.T) {
	dir := t.TempDir()
	writePages(t, dir, "index.ghp")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), nil, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	pages, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("len(pages) = %d, want 1: %+v", len(pages), pages)
	}
}

func TestScanEmptyDir(t *testing.T) {
	pages, err := Scan(t.TempDir())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("len(pages) = %d, want 0", len(pages))
	}
}

func TestScanPropagatesWalkErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root, permission 0000 does not block reads")
	}

	dir := t.TempDir()
	blocked := filepath.Join(dir, "no-permission")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	writePages(t, blocked, "index.ghp")
	if err := os.Chmod(blocked, 0o000); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(blocked, 0o755) })

	if _, err := Scan(dir); err == nil {
		t.Fatal("Scan() = nil error, want propagated permission error")
	}
}

func TestScanDetectsRouteConflict(t *testing.T) {
	// blog.ghp -> /blog and blog/index.ghp -> /blog: two different
	// pages, same route - a direct consequence of the "drop index"
	// convention, not an isolated bug.
	dir := t.TempDir()
	writePages(t, dir, "blog.ghp", "blog/index.ghp")

	_, err := Scan(dir)
	if err == nil {
		t.Fatal("Scan() = nil error, want route conflict")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("error = %T, want *ConflictError", err)
	}
	if conflict.Value != "/blog" {
		t.Errorf("Value = %q, want %q", conflict.Value, "/blog")
	}

	// filepath.WalkDir visits "blog/" (directory) before "blog.ghp" at
	// the same level - "blog" is a prefix of "blog.ghp", so it comes
	// first in lexical order - and descends into it completely before
	// continuing, so "blog/index.ghp" is discovered first.
	want := `blog/index.ghp and blog.ghp share the same route: "/blog"`
	if conflict.Error() != want {
		t.Errorf("Error() = %q, want %q", conflict.Error(), want)
	}
}

func TestScanDetectsFuncNameConflict(t *testing.T) {
	// blog-post.ghp and blog_post.ghp derive different routes
	// (/blog-post and /blog_post, no conflict there), but hyphen and
	// underscore become the same word boundary in deriveFuncName - the
	// two collide on "BlogPost". Without also checking FuncName, the
	// second generated <go:import> would never compile (redeclared
	// function).
	dir := t.TempDir()
	writePages(t, dir, "blog-post.ghp", "blog_post.ghp")

	_, err := Scan(dir)
	if err == nil {
		t.Fatal("Scan() = nil error, want function name conflict")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("error = %T, want *ConflictError", err)
	}
	if conflict.Value != "BlogPost" {
		t.Errorf("Value = %q, want %q", conflict.Value, "BlogPost")
	}
}

func TestScanDetectsUnderscoreSlashFuncNameConflict(t *testing.T) {
	// blog_post.ghp (root) and blog/post.ghp (subfolder) derive
	// different routes (/blog_post and /blog/post, no conflict there),
	// but both derive the same FuncName "BlogPost", because a "_" inside
	// one segment is indistinguishable from a "/" separating two
	// segments. They also derive the same GoFile ("blog_post.go"), yet
	// Scan only needs the FuncName check to reject the pair: GoFile is a
	// function of the same joined segments that define FuncName, so any
	// GoFile collision is necessarily a FuncName collision too.
	dir := t.TempDir()
	writePages(t, dir, "blog_post.ghp", "blog/post.ghp")

	_, err := Scan(dir)
	if err == nil {
		t.Fatal("Scan() = nil error, want func name conflict")
	}

	var conflict *ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("error = %T, want *ConflictError", err)
	}
	if conflict.Value != "BlogPost" {
		t.Errorf("Value = %q, want %q", conflict.Value, "BlogPost")
	}
}

func TestScanRejectsInvalidFuncName(t *testing.T) {
	// "2024.ghp" on its own (without another path segment to give it a
	// leading letter) derives FuncName "2024" - a valid Go numeric
	// literal, but not a valid identifier. Without this check,
	// transpiler would generate "mux.HandleFunc(\"/2024\", pages.2024)":
	// valid Go syntax, but semantically broken, only discovered later
	// with `go build`.
	dir := t.TempDir()
	writePages(t, dir, "2024.ghp")

	if _, err := Scan(dir); err == nil {
		t.Fatal("Scan() = nil error, want invalid function name error")
	}
}

func TestScanRejectsMalformedDynamicSegment(t *testing.T) {
	tests := []string{
		"blog/[slug.ghp", // missing the ']'
		"blog/slug].ghp", // missing the '['
		"blog/[].ghp",    // empty parameter name
	}

	for _, rel := range tests {
		t.Run(rel, func(t *testing.T) {
			dir := t.TempDir()
			writePages(t, dir, rel)

			if _, err := Scan(dir); err == nil {
				t.Fatalf("Scan() = nil error for %q, want malformed segment error", rel)
			}
		})
	}
}
