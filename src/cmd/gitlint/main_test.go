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
