package pages

import (
	"path/filepath"
	"strings"

	"ghp/src/internal/textutil"
)

// Page describes one .ghp page: where it lives, the identifiers its generated
// handler will use and the ServeMux route it is served at.
type Page struct {
	RelDir   string
	FileName string
	FuncName string
	PkgName  string
	Route    string
}

// Go returns the path to the generated Go file, rooted at root. The file base
// name drops any brackets so it stays a valid Go file name. Ex: root
// "/srv/src", RelDir "blog", FileName "about" -> "/srv/src/blog/about.go";
// FileName "[slug]" -> "/srv/src/blog/slug.go".
func (p *Page) Go(root string) string {
	return filepath.Join(root, p.RelDir, goName(p.FileName)+".go")
}

// goName strips characters that are not letters or digits so the result is a
// valid Go file base name. Ex: "[slug]" -> "slug".
func goName(fileName string) string {
	return textutil.StripNonWord(fileName)
}

// Ghp returns the path to the source .ghp page, rooted at root. Ex: root "/srv/src", RelDir "blog", FileName "about" -> "/srv/src/blog/about.ghp".
func (p *Page) Ghp(root string) string {
	return filepath.Join(root, p.RelDir, p.FileName+".ghp")
}

// route returns the ServeMux route this page is served at, fixed at build
// time so FileName can later be sanitized without losing the wildcard. Ex:
// blog/[slug].ghp -> "/blog/{slug}", blog/index.ghp -> "/blog".
func (p *Page) route() string {
	return p.Route
}

// computeRoute maps a page to its ServeMux route. An "index" page maps to
// its directory; a subdirectory page gets the directory as a prefix and its
// [param] wildcards become {param}. Ex: "blog/[slug].ghp" -> "/blog/{slug}",
// "blog/index.ghp" -> "/blog".
func computeRoute(relDir string, fileName string) string {
	path := "/" + relDir // /blog, /docs, /example/prt
	name := strings.ToLower(fileName)
	if name != "index" {
		if relDir != "" {
			path += "/"
		}
		path += textutil.BracketsToBraces(name)
	}
	return path
}

// NewPage builds Page from a page's relative path. The func name is the file, PascalCase'd; the package is the parent dir lower-cased, or "main" for a top-level page. Ex: "blog/about.ghp" -> "BlogAbout", "blog".
func NewPage(relPath string) *Page {
	parts := strings.Split(relPath, "/")
	last := len(parts) - 1

	pkgName := "main"
	if last > 0 {
		pkgName = strings.ToLower(textutil.NormalizeASCII(parts[last-1]))
	}
	pkgName = textutil.StripNonWord(pkgName)

	relDir := strings.Join(parts[:last], "/")
	fileName := strings.TrimSuffix(parts[last], ".ghp")

	return &Page{
		RelDir:   relDir,
		FileName: fileName,
		FuncName: textutil.PascalCase(textutil.NormalizeASCII(fileName)),
		PkgName:  pkgName,
		Route:    computeRoute(relDir, fileName),
	}
}
