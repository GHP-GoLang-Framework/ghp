// Package textutil provides shared string helpers for turning user input
// (file names, paths) into safe Go identifiers. Each helper answers one
// narrow question so pages, routes and codegen can reuse them without
// importing a kitchen-sink "utils" package: stripping accents, splicing
// word caps, dropping non-alphanumeric runes, rewriting route wildcards.
package textutil

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// nonWord matches any rune that is not a letter or digit. Everything else -
// punctuation, spaces, "_" - is removed so only alnum remains.
// Ex: "m-y_page" -> "mypage".
var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// StripNonWord removes every rune that is not a letter or digit.
// Ex: "[my-slug]v2" -> "myslugv2".
func StripNonWord(s string) string {
	return nonWord.ReplaceAllString(s, "")
}

// PascalCase capitalizes the first letter of each word (words split on any
// rune that is not a letter/digit/underscore), so a file name maps to a Go
// identifier. Ex: "my-page_v2" -> "MyPageV2".
func PascalCase(s string) string {
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
	}, s)
}

// NormalizeASCII strips accents by decomposing with NFD and dropping
// combining marks (Mn), keeping the ASCII letters. Ex: "órgão" -> "orgao".
func NormalizeASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, norm.NFD.String(s))
}

// paramRe matches a bracketed path parameter, allowing letters, digits and
// hyphens, for conversion to a ServeMux wildcard. Ex: "[my-slug]" -> "{my-slug}".
var paramRe = regexp.MustCompile(`\[([\w-]+)\]`)

// BracketsToBraces rewrites ServeMux-style wildcards from [name] to {name}.
// Segments without brackets are returned unchanged. Ex: "a/[id]-b" -> "a/{id}-b".
func BracketsToBraces(s string) string {
	return paramRe.ReplaceAllString(s, "{$1}")
}
