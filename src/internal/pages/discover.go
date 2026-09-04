package pages

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// Discover recursively walks absPath and returns a Page for every .ghp
// file found. The relPath inside each Page is computed relative to absPath.
func Discover(absPath string) []*Page {
	var pages []*Page

	filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".ghp") {
			return nil
		}

		rel, err := filepath.Rel(absPath, path)
		if err == nil {
			pages = append(pages, NewPage(filepath.ToSlash(rel)))
		}

		return nil
	})

	return pages
}
