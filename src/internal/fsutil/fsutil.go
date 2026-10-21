// Package fsutil provides shared filesystem utilities for walking directories
// and copying file trees. It is a leaf package with no internal dependencies.
package fsutil

import (
	"io/fs"
	"os"
	"path/filepath"
)

// WalkFunc is the callback for WalkDir. Return a non-nil error to stop the
// walk. Return SkipDir to skip a directory (like filepath.WalkDir).
type WalkFunc func(path string, d fs.DirEntry, err error) error

// SkipDir is returned by WalkFunc to skip a directory without visiting its
// contents.
var SkipDir = filepath.SkipDir

// WalkDir walks the directory tree rooted at root, calling fn for each file
// or directory. It is a thin wrapper around filepath.WalkDir that normalizes
// error handling and provides SkipFile semantics.
func WalkDir(root string, fn WalkFunc) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fn(path, d, err)
		}
		return fn(path, d, nil)
	})
}

// WalkFiles walks root and calls fn only for regular files (not directories).
// The skip function can return true to exclude a file from the walk. It
// receives the relative path from root and the DirEntry.
// Ex: WalkFiles(src, func(rel string, d fs.DirEntry) bool { return filepath.Ext(rel) != ".go" }, fn)
func WalkFiles(root string, skip func(rel string, d fs.DirEntry) bool, fn func(rel string, d fs.DirEntry) error) error {
	return WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if skip != nil && skip(rel, d) {
			return nil
		}
		return fn(rel, d)
	})
}

// CopyFile copies a single file from src to dst, creating parent directories
// as needed. It preserves the file mode.
func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, info.Mode())
}

// CopyTree copies the file tree from src to dst, calling skip for each file
// to decide whether to copy it. Directories are created as needed. The skip
// function receives the relative path from src and the DirEntry; return true
// to skip the file.
// Ex: CopyTree(src, dst, func(rel string, d fs.DirEntry) bool { return filepath.Ext(rel) != ".ghp" })
func CopyTree(src, dst string, skip func(rel string, d fs.DirEntry) bool) error {
	return WalkFiles(src, skip, func(rel string, d fs.DirEntry) error {
		srcPath := filepath.Join(src, rel)
		dstPath := filepath.Join(dst, rel)
		return CopyFile(srcPath, dstPath)
	})
}

// ReadFile is a convenience wrapper around os.ReadFile that includes the path
// in the error for better diagnostics.
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ReadError{Path: path, Err: err}
	}
	return data, nil
}

// WriteFile is a convenience wrapper around os.WriteFile that creates parent
// directories and includes the path in the error.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &WriteError{Path: path, Err: err}
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return &WriteError{Path: path, Err: err}
	}
	return nil
}

// ReadError wraps a file read error with the path.
type ReadError struct {
	Path string
	Err  error
}

func (e *ReadError) Error() string {
	return "read " + e.Path + ": " + e.Err.Error()
}

func (e *ReadError) Unwrap() error {
	return e.Err
}

// WriteError wraps a file write error with the path.
type WriteError struct {
	Path string
	Err  error
}

func (e *WriteError) Error() string {
	return "write " + e.Path + ": " + e.Err.Error()
}

func (e *WriteError) Unwrap() error {
	return e.Err
}

// MkdirAll is a convenience wrapper that includes the path in the error.
func MkdirAll(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return &MkdirError{Path: path, Err: err}
	}
	return nil
}

type MkdirError struct {
	Path string
	Err  error
}

func (e *MkdirError) Error() string {
	return "mkdir " + e.Path + ": " + e.Err.Error()
}

func (e *MkdirError) Unwrap() error {
	return e.Err
}
