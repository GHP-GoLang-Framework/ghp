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

// GhpFile describes one page: where it lives and the identifiers its generated handler will use.
type GhpFile struct {
	RelDir   string
	FileName string
	FuncName string
	PkgName  string
}

// Go returns the path to the generated Go file, rooted at root. Ex: root "/srv/src", RelDir "blog", FileName "about" -> "/srv/src/blog/about.go".
func (g GhpFile) Go(root string) string {
	return filepath.Join(root, g.RelDir, g.FileName+".go")
}

// Ghp returns the path to the source .ghp page, rooted at root. Ex: root "/srv/src", RelDir "blog", FileName "about" -> "/srv/src/blog/about.ghp".
func (g GhpFile) Ghp(root string) string {
	return filepath.Join(root, g.RelDir, g.FileName+".ghp")
}

// NewGhpFile builds GhpFile from a page's relative path. The func name is the file, PascalCase'd; the package is the parent dir lower-cased, or "main" for a top-level page. Ex: "blog/about.ghp" -> "BlogAbout", "blog".
func NewGhpFile(relPath string) GhpFile {
	parts := strings.Split(relPath, "/")
	last := len(parts) - 1

	pkgName := "main"
	if last > 0 {
		pkgName = strings.ToLower(normalizeASCII(parts[last-1]))
	}
	pkgName = nonWord.ReplaceAllString(pkgName, "")

	fileName := strings.TrimSuffix(parts[last], ".ghp")

	return GhpFile{
		RelDir:   strings.Join(parts[:last], "/"),
		FileName: fileName,
		FuncName: pascalCaseName(normalizeASCII(fileName)),
		PkgName:  pkgName,
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
