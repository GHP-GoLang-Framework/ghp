package parser

import (
	"ghp/src/internal/ast"
	"strings"
)

// scanner walks over the source one tag/text run at a time, tracking the
// current byte offset and 1-base line number so nodes cna report where
// they came from
type scanner struct {
	src  string
	pos  int
	line int
}

func newScanner(src string) *scanner {
	return &scanner{src: src, line: 1}
}

func (s *scanner) eof() bool {
	return s.pos >= len(s.src)
}

// advance moves pos foward by n bytes, keeping line in sync with any
// newlines skipped over.
func (s *scanner) advance(n int) {
	s.line += strings.Count(s.src[s.pos:s.pos+n], "\n")
	s.pos += n
}

// tagToken is one <go ..> tag recognized in the source, with its raw
// payload (whatever sits between the head and the closing '>', already
// trimmed) and the line the tag started on.
type tagToken struct {
	kind    tagKind
	payload string
	line    int
}

// nextText consumes source up to (but not including) the next '<' that
// opens a recognized GHP tag, and return it as a Text node. ok is false
// when there's no text before EOF or the netx tag - i.e a tag follows
// immediately, or the source is already exhauted.
func (s *scanner) nextText() (node *ast.Text, ok bool) {
	start := s.pos
	startLine := s.line

	for !s.eof() {
		if s.src[s.pos] == '<' {
			if kind, _ := matchTagHead(s.src[s.pos+1:]); kind != tagNone {
				break
			}
		}
		s.pos++
	}

	// Count the newlines crossed in one pass over the whole run, instead of
	// calling advance per byte (which would recount a slice each time).
	s.line += strings.Count(s.src[start:s.pos], "\n")

	if s.pos == start {
		return nil, false
	}
	return ast.NewText(s.src[start:s.pos], startLine), true
}

// nextTagToken asusmes s.pos sits on the '<' of a recognized tag (the
// caller checks this via matchTagHead, as nextText already does) and
// consumes through the closing '/>', returning the tag's kind, payload and
// starting line.
func (s *scanner) nextTagToken() (tagToken, error) {
	line := s.line
	kind, headLen := matchTagHead(s.src[s.pos+1:])
	payloadStart := s.pos + 1 + headLen

	closeIdx := findTagClose(s.src, payloadStart)
	if closeIdx == -1 {
		return tagToken{}, &SyntaxError{Line: line, Message: "tag not closed with '/>"}
	}

	payload := strings.TrimSpace(s.src[payloadStart:closeIdx])
	// Every GHP tag is self-closing: the '/' right before the '>' is the
	// closing marker, not part of the payload. Strip it (and any whitespace
	// it carries) so <go:if cond/> yields "cond", not "cond/".
	if !strings.HasSuffix(payload, "/") {
		return tagToken{}, &SyntaxError{Line: line, Message: "tag not closed with '/>"}
	}
	payload = strings.TrimSpace(strings.TrimSuffix(payload, "/"))
	s.advance(closeIdx + 1 - s.pos)

	return tagToken{kind: kind, payload: payload, line: line}, nil
}
