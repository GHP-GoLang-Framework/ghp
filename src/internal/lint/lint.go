// Package lint validates Conventional Commit messages. It reproduces the
// rules the repository used to enforce through commitlint, so the git
// commit-msg hook can run it without needing Node.js.
//
// The whole commit message is accepted; only the header (first line) is
// checked, matching how conventional commits are treated by commitlint.
package lint

import (
	"fmt"
	"strings"
)

// Types is the accepted set of commit types, mirroring the old commitlint
// type-enum rule.
var Types = []string{
	"feat",
	"fix",
	"docs",
	"style",
	"refactor",
	"perf",
	"test",
	"build",
	"ci",
	"chore",
	"revert",
}

// MaxHeaderLength caps the header line, matching commitlint's
// header-max-length rule.
const MaxHeaderLength = 100

// Message holds the commit header plus a validator that checks it against
// the Conventional Commits rules.
type Message struct {
	Header string
}

// Validate reports the first rule violation found in the message, or an
// empty string if the header is a valid Conventional Commit.
//
// Ex: "Feat(x): y" -> "type must be lower-case"
func (m Message) Validate() string {
	header := strings.TrimSuffix(m.Header, "\n")

	if len(header) > MaxHeaderLength {
		return fmt.Sprintf("header must not be longer than %d characters", MaxHeaderLength)
	}

	typ, rest, ok := splitType(header)
	if !ok {
		return "missing type: header must follow 'type(scope): subject'"
	}
	if !contains(Types, typ) {
		return fmt.Sprintf("type must be one of %s", strings.Join(Types, ", "))
	}

	scope, subject, ok := splitScope(rest)
	if !ok {
		return "missing subject: header must follow 'type(scope): subject'"
	}
	if scope != "" && scope != strings.ToLower(scope) {
		return "scope must be lower-case"
	}
	if subject == "" {
		return "subject must not be empty"
	}
	if strings.HasSuffix(subject, ".") {
		return "subject must not end with a full stop"
	}
	return ""
}

// splitType peels the leading type token off header.
// Ex: "feat(scope): x" -> ("feat", "(scope): x", true)
func splitType(header string) (typ, rest string, ok bool) {
	i := strings.IndexByte(header, '(')
	if i > 0 {
		return header[:i], header[i:], true
	}
	if j := strings.IndexByte(header, ':'); j >= 0 {
		return header[:j], header[j:], true
	}
	return "", "", false
}

// splitScope splits a "[(scope)]: subject" suffix into its parts; ok is
// false when there is no ':'. Ex: "(scope): x" -> ("scope", "x", true).
func splitScope(rest string) (scope, subject string, ok bool) {
	j := strings.IndexByte(rest, ':')
	if j < 0 {
		return "", "", false
	}
	left := rest[:j]
	subject = strings.TrimSpace(rest[j+1:])
	if strings.HasPrefix(left, "(") && strings.HasSuffix(left, ")") {
		scope = left[1 : len(left)-1]
	}
	return scope, subject, true
}

// contains reports whether s is present in list.
func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
