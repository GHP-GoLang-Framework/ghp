package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBuild(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)

	var out bytes.Buffer
	if code := run([]string{"build", dir}, &out); code != 0 {
		t.Fatalf("run(build dir) exit = %d, want 0\nout:\n%s", code, out.String())
	}
	for _, name := range []string{"main.go", "go.mod", "pages/index.go"} {
		if _, err := os.Stat(filepath.Join(dir, "build", name)); err != nil {
			t.Errorf("missing generated file build/%s: %v", name, err)
		}
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
		wantContains []string
	}{
		{
			name:         "no args prints usage and exits 2",
			args:         nil,
			wantExitCode: 2,
			wantContains: []string{"ghp <command>"},
		},
		{
			name:         "help prints usage",
			args:         []string{"help"},
			wantExitCode: 0,
			wantContains: []string{"ghp <command>"},
		},
		{
			name:         "unknown command prints usage and exits 2",
			args:         []string{"bogus"},
			wantExitCode: 2,
			wantContains: []string{"Command unknown", "ghp <command>"},
		},
		{
			name:         "dev with a bogus flag exits 2",
			args:         []string{"dev", "--bogus"},
			wantExitCode: 2,
			wantContains: []string{"ghp dev: unexpected flag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer

			gotExitCode := run(tt.args, &out)

			if gotExitCode != tt.wantExitCode {
				t.Errorf("run(%v) exit code = %d, want %d", tt.args, gotExitCode, tt.wantExitCode)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(out.String(), want) {
					t.Errorf("run(%v) output missing %q\ngot: %s", tt.args, want, out.String())
				}
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	var out bytes.Buffer

	printUsage(&out)

	for _, want := range []string{"ghp <command>", "dev", "build", "help"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("printUsage() output missing %q\ngot: %s", want, out.String())
		}
	}
}
