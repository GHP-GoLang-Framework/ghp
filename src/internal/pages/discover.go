package pages

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Discover recursively walks absPath and returns a Page for every .ghp
// file found. The relPath inside each Page is computed relative to absPath.
// A root that cannot be read is an error: silently returning no pages would
// turn a mistyped project path into an empty build instead of a diagnosis.
func Discover(absPath string) ([]*Page, error) {
	var discovered []*Page

	err := filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".ghp") {
			return nil
		}

		rel, err := filepath.Rel(absPath, path)
		if err != nil {
			return err
		}

		discovered = append(discovered, NewPage(filepath.ToSlash(rel)))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("discover %s: %w", absPath, err)
	}

	return discovered, nil
}
