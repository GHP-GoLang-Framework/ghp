package pages

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"ghp/src/internal/fsutil"
)

// Discover recursively walks absPath and returns a Page for every .ghp
// file found. The relPath inside each Page is computed relative to absPath.
// A root that cannot be read is an error: silently returning no pages would
// turn a mistyped project path into an empty build instead of a diagnosis.
func Discover(absPath string) ([]*Page, error) {
	var discovered []*Page

	err := fsutil.WalkFiles(absPath, func(rel string, d fs.DirEntry) bool {
		return filepath.Ext(rel) != ".ghp"
	}, func(rel string, d fs.DirEntry) error {
		discovered = append(discovered, NewPage(filepath.ToSlash(rel)))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover %s: %w", absPath, err)
	}

	return discovered, nil
}
