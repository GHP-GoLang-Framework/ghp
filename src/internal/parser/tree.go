package parser

import (
	"fmt"
	"ghp/src/internal/ast"
	"strings"
)

var tagKindNames = map[tagKind]string{
	tagNone:        "text",
	tagImport:      "<go:import/>",
	tagStatement:   "<go/>",
	tagEcho:        "<go=/>",
	tagIf:          "<go:if/>",
	tagElse:        "<go:else/>",
	tagElif:        "<go:elif/>",
	tagSwitch:      "<go:switch/>",
	tagCase:        "<go:case/>",
	tagDefault:     "<go:default/>",
	tagFor:         "<go:for/>",
	tagCloseIf:     "<go:endif/>",
	tagCloseSwitch: "<go:endswitch/>",
	tagCloseFor:    "<go:endfor/>",
}

func (k tagKind) String() string {
	if name, ok := tagKindNames[k]; ok {
		return name
	}
	return "unknown tag"
}

// The stop sets parseNodes pauses on are shared instead of rebuilt per
// call: they are only ever read, and a <go:for>/<go:if>/<go:switch> in a
// deep tree would otherwise allocate one map each.
var (
	stopEndFor    = map[tagKind]bool{tagCloseFor: true}
	stopEndIf     = map[tagKind]bool{tagCloseIf: true}
	stopElifOrEnd = map[tagKind]bool{tagElif: true, tagElse: true, tagCloseIf: true}
	stopElseBody  = map[tagKind]bool{tagElif: true, tagCloseIf: true}
	stopEndSwitch = map[tagKind]bool{tagCase: true, tagDefault: true, tagCloseSwitch: true}
)

// parseNodes parses a sequence of nodes until EOF or one of the stop kinds
// is reached. The stop tag itself is consumed from the scanner but not
// turned into a node - it's returned to the caller instead, so callers
// like buildIf can tell a <go:else/> apart from a <go:endif/>. A nil stop
// map means "run to EOF" (used at the top level), in which case the
// returned token is always the zero tagToken.
func parseNodes(s *scanner, stop map[tagKind]bool) ([]ast.Node, tagToken, error) {
	var nodes []ast.Node

	for {
		if text, ok := s.nextText(); ok {
			nodes = append(nodes, text)
		}

		if s.eof() {
			return nodes, tagToken{}, nil
		}

		tok, err := s.nextTagToken()
		if err != nil {
			return nil, tagToken{}, err
		}

		if stop[tok.kind] {
			return nodes, tok, nil
		}

		node, err := buildNode(s, tok)
		if err != nil {
			return nil, tagToken{}, err
		}
		nodes = append(nodes, node)
	}
}

// buildNode consumes the body (and any nested tags) belonging to tok and
// returns the finished node. tok has already been read from s; buildNode
// is responsible for reading everything up to and including tok's
// matching close tag, for tags that have one.
func buildNode(s *scanner, tok tagToken) (ast.Node, error) {
	switch tok.kind {
	case tagImport:
		paths, err := parseImportPaths(tok.payload)
		if err != nil {
			return nil, &SyntaxError{Line: tok.line, Message: err.Error()}
		}
		return ast.NewImport(paths, tok.line), nil

	case tagStatement:
		return ast.NewStatement(tok.payload, tok.line), nil

	case tagEcho:
		// An empty <go= /> would generate fmt.Sprint() - valid Go that
		// renders an empty string, so the typo would go unnoticed.
		if tok.payload == "" {
			return nil, &SyntaxError{Line: tok.line, Message: "<go=> requires an expression"}
		}
		return ast.NewEcho(tok.payload, tok.line), nil

	case tagIf:
		// Checked here, before the body is read, so the report points
		// at the <go:if> itself even when the body is what's broken.
		if tok.payload == "" {
			return nil, &SyntaxError{Line: tok.line, Message: "<go:if> requires a condition"}
		}
		return buildIf(s, tok)

	case tagSwitch:
		return buildSwitch(s, tok)

	case tagFor:
		// An empty <go:for/> compiles into `for {}` - an infinite loop
		// that builds clean and hangs the request, so it has to be
		// caught here.
		if tok.payload == "" {
			return nil, &SyntaxError{Line: tok.line, Message: "<go:for> requires a loop expression"}
		}
		body, close, err := parseNodes(s, stopEndFor)
		if err != nil {
			return nil, err
		}
		if close.kind == tagNone {
			return nil, &SyntaxError{Line: tok.line, Message: "<go:for> missing <go:endfor/>"}
		}
		if err := rejectPayload(close); err != nil {
			return nil, err
		}
		return ast.NewFor(tok.payload, body, tok.line), nil

	default:
		return nil, &SyntaxError{Line: tok.line, Message: fmt.Sprintf("unexpected %s here", tok.kind)}
	}
}

// rejectPayload reports the tags that mean something only on their own -
// the closers <go:endif/>, <go:endfor/> and <go:endswitch/> - when one is
// written carrying a payload. It would otherwise be dropped on the floor
// without a word. <go:else/> has its own, more specific message in buildIf.
func rejectPayload(tok tagToken) error {
	if tok.payload == "" {
		return nil
	}
	return &SyntaxError{Line: tok.line, Message: fmt.Sprintf("%s does not take a payload", tok.kind)}
}

// buildIf handles a <go:if> that already had its head read (tok). It reads
// the "then" body up to <go:elif/>, <go:else/>, or <go:endif/>; when it
// stops at <go:elif/>, it collects the elif condition and body, then loops
// until a non-elif stop tag is reached.  Elifs, Else stay nil (not empty
// slices) when there are no <go:elif/> or <go:else/> tags respectively,
// matching ast.If's contract.
func buildIf(s *scanner, tok tagToken) (ast.Node, error) {
	then, close, err := parseNodes(s, stopElifOrEnd)
	if err != nil {
		return nil, err
	}
	if close.kind == tagNone {
		return nil, &SyntaxError{Line: tok.line, Message: "<go:if> missing <go:endif/>"}
	}

	var elifs []ast.ElseIf
	for close.kind == tagElif {
		if close.payload == "" {
			return nil, &SyntaxError{Line: close.line, Message: "<go:elif> requires a condition"}
		}
		elifBody, after, err := parseNodes(s, stopElifOrEnd)
		if err != nil {
			return nil, err
		}
		elifs = append(elifs, ast.ElseIf{Cond: close.payload, Body: elifBody, Line: close.line})
		close = after
	}

	if close.kind != tagElse && close.kind != tagCloseIf {
		return nil, &SyntaxError{Line: tok.line, Message: "<go:if> missing <go:endif/>"}
	}

	var els []ast.Node
	if close.kind == tagElse {
		if close.payload != "" {
			return nil, &SyntaxError{Line: close.line, Message: "<go:else> does not take a condition - chained else-if isn't supported, use nested <go:if> instead"}
		}

		// The else body stops on <go:endif/> like any other, but also
		// on <go:elif/>: a bare elif after the else is a real (and
		// otherwise silent) mistake - the condition would be dropped,
		// since the chain above is already closed - so it deserves its
		// own message instead of "unexpected <go:elif/> here".
		els, close, err = parseNodes(s, stopElseBody)
		if err != nil {
			return nil, err
		}
		if close.kind == tagElif {
			return nil, &SyntaxError{Line: close.line, Message: "<go:elif> must come before <go:else/>"}
		}
		if close.kind != tagCloseIf {
			return nil, &SyntaxError{Line: tok.line, Message: "<go:if> missing <go:endif/>"}
		}
	}

	if err := rejectPayload(close); err != nil {
		return nil, err
	}

	return ast.NewIf(tok.payload, then, elifs, els, tok.line), nil
}

// buildSwitch handles a <go:switch> that already had its head read (tok).
// Its body is a run of <go:case/>/<go:default/> blocks up to
// <go:endswitch/>; any text sitting directly between <go:switch> and the
// first case is rejected rather than dropped, since it has nowhere to
// live on the Switch node - a reader would see it in the .ghp and never
// in the page.
func buildSwitch(s *scanner, tok tagToken) (ast.Node, error) {
	sw := ast.NewSwitch(tok.payload, nil, nil, tok.line)

	before, next, err := parseNodes(s, stopEndSwitch)
	if err != nil {
		return nil, err
	}
	// An unterminated switch outranks anything the body may also be
	// missing, so the report stays on the absent <go:endswitch/>.
	if next.kind == tagNone {
		return nil, &SyntaxError{Line: tok.line, Message: "<go:switch> missing <go:endswitch/>"}
	}
	if hasContent(before) {
		// The line points at the <go:switch> rather than at the stray
		// content: a Text node reports the line it starts on, which
		// includes the newline left over from the tag above it.
		return nil, &SyntaxError{Line: tok.line, Message: "<go:switch> body must start with <go:case/> or <go:default/>"}
	}

	sawDefault := false
	for next.kind != tagCloseSwitch {
		switch next.kind {
		case tagCase:
			caseTok := next
			if caseTok.payload == "" {
				return nil, &SyntaxError{Line: caseTok.line, Message: "<go:case> requires at least one value"}
			}
			body, after, err := parseNodes(s, stopEndSwitch)
			if err != nil {
				return nil, err
			}
			sw.Cases = append(sw.Cases, ast.Case{Value: caseTok.payload, Body: body, Line: caseTok.line})
			next = after

		case tagDefault:
			if sawDefault {
				return nil, &SyntaxError{Line: next.line, Message: "duplicate <go:default>"}
			}
			sawDefault = true
			body, after, err := parseNodes(s, stopEndSwitch)
			if err != nil {
				return nil, err
			}
			sw.Default = body
			next = after

		default:
			return nil, &SyntaxError{Line: tok.line, Message: "<go:switch> missing <go:endswitch/>"}
		}
	}

	if err := rejectPayload(next); err != nil {
		return nil, err
	}
	return sw, nil
}

// hasContent reports whether nodes hold anything a reader of the page
// would actually see, skipping the whitespace-only Text nodes that are
// just the layout between two tags.
func hasContent(nodes []ast.Node) bool {
	for _, n := range nodes {
		if text, ok := n.(*ast.Text); ok && strings.TrimSpace(text.Value) == "" {
			continue
		}
		return true
	}
	return false
}

// parseImportPaths splits a <go:import (...)/> payload into its individual
// paths. GHP mirrors Go's own import syntax: parentheses hold one path per
// line (or comma-separated), each optionally prefixed with an alias.
func parseImportPaths(payload string) ([]ast.ImportPath, error) {
	inner := strings.TrimSpace(payload)
	inner = strings.TrimPrefix(inner, "(")
	inner = strings.TrimSuffix(inner, ")")

	var paths []ast.ImportPath
	for _, item := range strings.FieldsFunc(inner, func(r rune) bool {
		return r == '\n' || r == ','
	}) {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		var alias, path string
		switch fields := strings.Fields(item); len(fields) {
		case 1:
			path = fields[0]
		case 2:
			alias, path = fields[0], fields[1]
		default:
			// An item holding more than "alias path" is two imports
			// written on one line with nothing between them - reading
			// them as a single one would silently invent an alias out
			// of the first path, so it is reported instead.
			return nil, fmt.Errorf("<go:import> %s is not a single import path (use one path per line, or separate them with commas)", item)
		}

		if strings.ContainsAny(alias, `"'`) {
			return nil, fmt.Errorf("<go:import> alias %s cannot be quoted", alias)
		}

		paths = append(paths, ast.ImportPath{Alias: alias, Path: strings.Trim(path, `"`)})
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("empty <go:import>")
	}
	return paths, nil
}
