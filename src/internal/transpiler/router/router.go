package router

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// SearchGhpFiles recursively walks absPath and returns a GhpFile for every
// .ghp file found. The relPath inside each GhpFile is computed relative to
// absPath.
func SearchGhpFiles(absPath string) []GhpFile {
	var files []GhpFile

	filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".ghp") {
			return nil
		}

		rel, err := filepath.Rel(absPath, path)
		if err == nil {
			file := NewGhpFile(filepath.ToSlash(rel))
			files = append(files, file)
		}

		return nil
	})

	return files
}
