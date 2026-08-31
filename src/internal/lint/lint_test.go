package lint

import "testing"

func TestMessageValidate(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{name: "valid with scope", header: "feat(parser): add for tag", want: ""},
		{name: "valid without scope", header: "fix: correct echo spacing", want: ""},
		{name: "valid build", header: "build(deps): bump go version", want: ""},
		{name: "valid trailing newline", header: "docs: update readme\n", want: ""},

		{name: "missing type", header: "add for tag", want: "missing type: header must follow 'type(scope): subject'"},
		{name: "type not in enum", header: "feature: add tag", want: "type must be one of feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert"},
		{name: "type upper case", header: "feat(Parser): add tag", want: "scope must be lower-case"},
		{name: "subject trailing full stop", header: "feat: add tag.", want: "subject must not end with a full stop"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (Message{Header: tt.header}).Validate(); got != tt.want {
				t.Errorf("Validate(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestSplitType(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		wantType string
		wantRest string
		wantOK   bool
	}{
		{name: "with scope", header: "feat(scope): x", wantType: "feat", wantRest: "(scope): x", wantOK: true},
		{name: "no scope", header: "fix: x", wantType: "fix", wantRest: ": x", wantOK: true},
		{name: "no colon", header: "feat x", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, rest, ok := splitType(tt.header)
			if typ != tt.wantType || rest != tt.wantRest || ok != tt.wantOK {
				t.Errorf("splitType(%q) = (%q, %q, %v), want (%q, %q, %v)", tt.header, typ, rest, ok, tt.wantType, tt.wantRest, tt.wantOK)
			}
		})
	}
}

func TestSplitScope(t *testing.T) {
	tests := []struct {
		name      string
		rest      string
		wantScope string
		wantSubj  string
		wantOK    bool
	}{
		{name: "with scope", rest: "(parser): x", wantScope: "parser", wantSubj: "x", wantOK: true},
		{name: "without scope", rest: ": x", wantScope: "", wantSubj: "x", wantOK: true},
		{name: "no colon", rest: "foo", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, subj, ok := splitScope(tt.rest)
			if scope != tt.wantScope || subj != tt.wantSubj || ok != tt.wantOK {
				t.Errorf("splitScope(%q) = (%q, %q, %v), want (%q, %q, %v)", tt.rest, scope, subj, ok, tt.wantScope, tt.wantSubj, tt.wantOK)
			}
		})
	}
}

func TestContains(t *testing.T) {
	if !contains(Types, "feat") {
		t.Error("contains(Types, feat) = false, want true")
	}
	if contains(Types, "nope") {
		t.Error("contains(Types, nope) = true, want false")
	}
}

func TestMaxHeaderLength(t *testing.T) {
	long := "feat: " + string(make([]byte, 100))
	if got := (Message{Header: long}).Validate(); got == "" {
		t.Error("Validate(header > 100 chars) = empty, want an error")
	}
}
