package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	msgFile := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")

	tests := []struct {
		name     string
		args     []string
		fileData string
		wantCode int
		wantErr  string
	}{
		{name: "valid message", args: []string{"--edit", msgFile}, fileData: "feat(parser): add for tag\n", wantCode: 0},
		{name: "valid with body", args: []string{"--edit", msgFile}, fileData: "feat(parser): add for tag\n\nLong body over many lines explains the why.\n", wantCode: 0},
		{name: "invalid message", args: []string{"--edit", msgFile}, fileData: "not a commit\n", wantCode: 1, wantErr: "missing type: header must follow 'type(scope): subject'"},
		{name: "missing edit flag", args: nil, wantCode: 2, wantErr: "usage: gitlint --edit <file>"},
		{name: "file does not exist", args: []string{"--edit", filepath.Join(t.TempDir(), "nope")}, wantCode: 1, wantErr: "gitlint: open "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fileData != "" {
				if err := os.WriteFile(msgFile, []byte(tt.fileData), 0o600); err != nil {
					t.Fatalf("write message file: %v", err)
				}
			} else if err := os.Remove(msgFile); err != nil && !os.IsNotExist(err) {
				t.Fatalf("remove message file: %v", err)
			}

			var stdout, stderr bytes.Buffer
			if got := run(tt.args, &stdout, &stderr); got != tt.wantCode {
				t.Errorf("run() code = %d, want %d", got, tt.wantCode)
			}
			if tt.wantErr != "" && !strings.Contains(stderr.String(), tt.wantErr) {
				t.Errorf("run() stderr = %q, want it to contain %q", stderr.String(), tt.wantErr)
			}
		})
	}
}

func TestMainFunction(t *testing.T) {
	if os.Getenv("GITLINT_TEST_MAIN") == "1" {
		main()
		t.Fatal("main returned without exiting")
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainFunction")
	cmd.Env = append(os.Environ(), "GITLINT_TEST_MAIN=1")
	if err := cmd.Run(); err == nil {
		t.Fatal("main() with no --edit should exit non-zero")
	} else {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}
		if code := ee.ProcessState.ExitCode(); code != 2 {
			t.Errorf("main() exit code = %d, want 2", code)
		}
	}
}

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
		{name: "missing subject after scope", header: "feat(", want: "missing subject: header must follow 'type(scope): subject'"},
		{name: "subject empty", header: "feat:", want: "subject must not be empty"},
		{name: "scope upper case", header: "feat(Parser): add tag", want: "scope must be lower-case"},
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
