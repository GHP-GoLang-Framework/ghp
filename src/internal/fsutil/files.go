package fsutil

import (
	"io"
	"os"
	"path/filepath"
)

// CopyTree copies the tree under src to dst, calling itself again for each subdirectory. The .git directory is never copied. dst is created if missing. cb runs once per copied entry (file or directory), with the entry's absolute source path and whether it is a directory, after the copy. Ex: CopyTree("blog", "out", cb) copies blog/faq.md to out/faq.md and calls cb with the absolute /.../blog/faq.md and false.
func CopyTree(src string, dst string, cb func(path string, isDir bool)) error {
	if err := CreateDir(src); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err == nil {
		for _, e := range entries {
			srcPath := filepath.Join(src, e.Name())
			dstPath := filepath.Join(dst, e.Name())

			isDir := e.IsDir()
			if isDir {
				if e.Name() == ".git" {
					continue
				}
				if err := CopyTree(srcPath, dstPath, cb); err != nil {
					return err
				}
			} else if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
			cb(srcPath, isDir)
		}
	}

	return err
}

// CopyFile copies a single file from src to dst, preserving its mode.
func CopyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err == nil {
		defer out.Close()
		_, err = io.Copy(out, in)
	}

	return err
}

func CreateDir(path string) error {
	return os.Mkdir(path, 0o755)
}

func CreateFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
