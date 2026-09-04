package router

import (
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// nonWord matches any rune that is not a letter or digit. Everything else - punctuation, spaces, "_" - is removed so only alnum remains. Ex: "m-y_page" -> "mypage".
var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// GhpFile describes one page: where it lives, the identifiers its generated
// handler will use and the ServeMux route it is served at.
type GhpFile struct {
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
func (g *GhpFile) Go(root string) string {
	return filepath.Join(root, g.RelDir, goName(g.FileName)+".go")
}

// goName strips characters that are not letters or digits so the result is a
// valid Go file base name. Ex: "[slug]" -> "slug".
func goName(fileName string) string {
	return nonWord.ReplaceAllString(fileName, "")
}

// Ghp returns the path to the source .ghp page, rooted at root. Ex: root "/srv/src", RelDir "blog", FileName "about" -> "/srv/src/blog/about.ghp".
func (g *GhpFile) Ghp(root string) string {
	return filepath.Join(root, g.RelDir, g.FileName+".ghp")
}

// route returns the ServeMux route this page is served at, fixed at build
// time so FileName can later be sanitized without losing the wildcard. Ex:
// blog/[slug].ghp -> "/blog/{slug}", blog/index.ghp -> "/blog".
func (g *GhpFile) route() string {
	return g.Route
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
		path += changeBrackets(name)
	}
	return path
}

// paramRe matches a bracketed path parameter, allowing letters, digits and
// hyphens, for conversion to a ServeMux wildcard. Ex: "[my-slug]" -> "{my-slug}".
var paramRe = regexp.MustCompile(`\[([\w-]+)\]`)

// changeBrackets rewrites ServeMux-style wildcards from [name] to {name}.
// Segments without brackets are returned unchanged. Ex: "a/[id]-b" ->
// "a/{id}-b".
func changeBrackets(s string) string {
	return paramRe.ReplaceAllString(s, "{$1}")
}

// NewGhpFile builds GhpFile from a page's relative path. The func name is the file, PascalCase'd; the package is the parent dir lower-cased, or "main" for a top-level page. Ex: "blog/about.ghp" -> "BlogAbout", "blog".
func NewGhpFile(relPath string) *GhpFile {
	parts := strings.Split(relPath, "/")
	last := len(parts) - 1

	pkgName := "main"
	if last > 0 {
		pkgName = strings.ToLower(normalizeASCII(parts[last-1]))
	}
	pkgName = nonWord.ReplaceAllString(pkgName, "")

	relDir := strings.Join(parts[:last], "/")
	fileName := strings.TrimSuffix(parts[last], ".ghp")

	return &GhpFile{
		RelDir:   relDir,
		FileName: fileName,
		FuncName: pascalCaseName(normalizeASCII(fileName)),
		PkgName:  pkgName,
		Route:    computeRoute(relDir, fileName),
	}
}

// pascalCaseName capitalizes the first letter of each word (words split on
// any rune that is not a letter/digit/underscore), so a file name maps to a
// Go identifier. Ex: "my-page_v2" -> "MyPageV2".
func pascalCaseName(name string) string {
	capitalizeNext := true

	return strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalizeNext = true
			return -1
		}

		if capitalizeNext {
			r = unicode.ToUpper(r)
			capitalizeNext = false
		} else {
			r = unicode.ToLower(r)
		}

		return r
	}, name)
}

// normalizeASCII strips accents by decomposing with NFD and dropping
// combining marks (Mn), keeping the ASCII letters. Ex: "órgão" -> "orgao".
func normalizeASCII(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, norm.NFD.String(name))
}
