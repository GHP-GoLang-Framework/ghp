package parser

import "testing"

func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantErr string
	}{
		{"unclosed if", `<go:if a == b/>text`, "line 1: <go:if> missing <go:endif/>"},
		{"tag without />", `<go:if a == b>text<go:endif/>`, "line 1: tag not closed with '/>"},
		{"tag never closed", `<go:if a == b`, "line 1: tag not closed with '/>"},
		{"stray close", `text <go:endif/> more text`, "line 1: unexpected <go:endif/> here"},
		{"else without if", `<go:else/>hi<go:endif/>`, "line 1: unexpected <go:else/> here"},
		{"duplicate default", "<go:switch v/>\n<go:default/>a\n<go:default/>b\n<go:endswitch/>", "line 3: duplicate <go:default>"},
		{"if greater-than in cond needs parens", `<go:if (a > b)/>x<go:endif/>`, ""},
		{"unclosed for", `<go:for i := range xs/>text`, "line 1: <go:for> missing <go:endfor/>"},
		{"empty import", `<go:import ()/>`, "line 1: empty <go:import>"},
		{"error propagated from inside the for", `<go:for i/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"error propagated from the if's then body", `<go:if a/><go:for i/>x`, "line 1: <go:for> missing <go:endfor/>"},
		{"error propagated from the if's else body", `<go:if a/>x<go:else/><go:for i/>y`, "line 1: <go:for> missing <go:endfor/>"},
		{"error propagated from an elif body", `<go:if a/>x<go:elif b/><go:if c/>y`, "line 1: <go:if> missing <go:endif/>"},
		{"if with else but never closed", `<go:if a/>x<go:else/>y`, "line 1: <go:if> missing <go:endif/>"},
		{"error before the switch's first case", `<go:switch v/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"error inside a case body", `<go:switch v/><go:case 1/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"error inside the default body", `<go:switch v/><go:default/><go:if a/>x`, "line 1: <go:if> missing <go:endif/>"},
		{"switch with no case/default/close", `<go:switch v/>text`, "line 1: <go:switch> missing <go:endswitch/>"},
		{"chained else-if via go:else still rejected", `<go:if a/>x<go:else b/>y<go:endif/>`, "line 1: <go:else> does not take a condition - chained else-if isn't supported, use nested <go:if> instead"},
		{"elif without condition", `<go:if a/>x<go:elif/>y<go:endif/>`, "line 1: <go:elif> requires a condition"},
		{"unclosed if after elif", `<go:if a/>x<go:elif b/>y`, "line 1: <go:if> missing <go:endif/>"},

		// Payloads that only failed later, if at all: an empty
		// <go:for/> compiled into `for {}` (valid Go, so it built clean
		// and hung the request), while an empty <go:if/>/<go= /> only
		// surfaced as "assembled file is not valid Go", pointing at the
		// generated .go line instead of the .ghp.
		{"empty for", `<go:for/>x<go:endfor/>`, "line 1: <go:for> requires a loop expression"},
		{"empty if condition", `<go:if/>x<go:endif/>`, "line 1: <go:if> requires a condition"},
		{"empty echo", `<go= />`, "line 1: <go=> requires an expression"},

		// Payloads on tags that mean something only on their own - they
		// used to be dropped on the floor without a word.
		{"endif with payload", `<go:if a/>x<go:endif y/>`, "line 1: <go:endif/> does not take a payload"},
		{"endfor with payload", `<go:for i/>x<go:endfor y/>`, "line 1: <go:endfor/> does not take a payload"},
		{"endswitch with payload", `<go:switch v/><go:endswitch x/>`, "line 1: <go:endswitch/> does not take a payload"},
		{"case without value", `<go:switch v/><go:case/>x<go:endswitch/>`, "line 1: <go:case> requires at least one value"},

		// Content between <go:switch> and its first case has nowhere to
		// live on the Switch node - it used to vanish from the page.
		{"stray content before the first case", "<go:switch v/>\nnope\n<go:case 1/>x\n<go:endswitch/>", "line 1: <go:switch> body must start with <go:case/> or <go:default/>"},
		{"whitespace before the first case is fine", "<go:switch v/>\n\t<go:case 1/>x\n<go:endswitch/>", ""},

		{"elif after else", `<go:if a/>x<go:else/>y<go:elif b/>z<go:endif/>`, "line 1: <go:elif> must come before <go:else/>"},

		// Import problems used to be reported by the codegen, after the
		// whole tree was already built - now Parse rejects them.
		{"import alias that is quoted", `<go:import ("f" "fmt")/>`, `line 1: <go:import> alias "f" cannot be quoted`},
		{"two paths on one line with nothing between them", `<go:import ("strings" f "fmt")/>`, `line 1: <go:import> "strings" f "fmt" is not a single import path (use one path per line, or separate them with commas)`},
		{"same path under two aliases", "<go:import (a \"strings\")/>\n<go:import (b \"strings\")/>x", `line 2: "strings" is imported with alias "b", but line 1 already imports it with "a"`},
		{"alias on an automatic package", `<go:import (myio "io")/>x<go= y />`, `line 1: "io" is managed automatically by the page and cannot be imported with an alias ("myio")`},
		{"import nested two bodies deep", `<go:if a/><go:for i/><go:import ("fmt")/>x<go:endfor/><go:endif/>`, "line 1: <go:import> is not at the top level of the file - conditional imports do not exist in Go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.src)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Parse(%q) unexpected error: %v", tt.src, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Parse(%q) = nil error, want %q", tt.src, tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Parse(%q) error = %q, want %q", tt.src, err.Error(), tt.wantErr)
			}
		})
	}
}
