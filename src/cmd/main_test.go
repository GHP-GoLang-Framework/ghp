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

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// minProject mirrors build.minProject: scaffold + one page.
func minProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/site\n\ngo 1.26\n")
	writeFile(t, root, "main.go", `package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	AddRoutes(mux)
}
`)
	writeFile(t, root, "about.ghp", "<h1>About</h1>\n")
	return root
}

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"no args prints usage", nil, 2},
		{"help", []string{"help"}, 0},
		{"unknown command", []string{"bogus"}, 2},
		{"version", []string{"version"}, 0},
		{"--version", []string{"--version"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := run(tt.args, &bytes.Buffer{}); got != tt.want {
				t.Errorf("run(%v) = %d, want %d", tt.args, got, tt.want)
			}
		})
	}
}

func TestVersionPrintsValue(t *testing.T) {
	var buf bytes.Buffer
	if got := run([]string{"version"}, &buf); got != 0 {
		t.Errorf("run(version) = %d, want 0", got)
	}
	if got, want := buf.String(), version+"\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}

func TestRunBuildWithoutGoMod(t *testing.T) {
	dir := t.TempDir()
	if got := run([]string{"build", dir}, &bytes.Buffer{}); got != 1 {
		t.Errorf("run(build, empty dir) = %d, want 1", got)
	}
}

func TestRunDevWithoutGoMod(t *testing.T) {
	dir := t.TempDir()
	if got := run([]string{"dev", dir}, &bytes.Buffer{}); got != 1 {
		t.Errorf("run(dev, empty dir) = %d, want 1", got)
	}
}

func TestBuildSuccess(t *testing.T) {
	dir := minProject(t)
	if got := Build([]string{dir}); got != 0 {
		t.Fatalf("Build = %d, want 0", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "app")); err != nil {
		t.Errorf("app not built: %v", err)
	}
}

func TestBuildProjectError(t *testing.T) {
	dir := minProject(t)
	writeFile(t, dir, "broken.ghp", "<go:if n == 1/>\n")
	if got := Build([]string{dir}); got != 1 {
		t.Errorf("Build = %d, want 1", got)
	}
}

func TestDevSuccess(t *testing.T) {
	dir := minProject(t)
	if got := Dev([]string{dir}); got != 0 {
		t.Fatalf("Dev = %d, want 0", got)
	}
}

func TestResolveParameters(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    ghpParameters
		wantErr string
	}{
		{"loose path sets project", []string{"site"}, ghpParameters{projectPath: "site"}, ""},
		{"project flag", []string{"-p", "site"}, ghpParameters{projectPath: "site"}, ""},
		{"project flag equals form", []string{"--project=site"}, ghpParameters{projectPath: "site"}, ""},
		{"output flag", []string{"-o", "bin/app"}, ghpParameters{output: "bin/app"}, ""},
		{"output equals form", []string{"--output=bin/app"}, ghpParameters{output: "bin/app"}, ""},
		{"entry point flag", []string{"-e", "cli"}, ghpParameters{entryPointDir: "cli"}, ""},
		{"entry point equals form", []string{"--entry-point=cmd/server"}, ghpParameters{entryPointDir: "cmd/server"}, ""},
		{"mixed flags and loose path", []string{"-o", "bin/app", "site"}, ghpParameters{projectPath: "site", output: "bin/app"}, ""},
		{"flag value may start with dash", []string{"-o", "-net"}, ghpParameters{output: "-net"}, ""},
		{"duplicate project path", []string{"site", "other"}, ghpParameters{}, "set twice"},
		{"project flag plus loose path", []string{"-p", "site", "other"}, ghpParameters{}, "set twice"},
		{"missing value", []string{"-o"}, ghpParameters{}, "missing value"},
		{"unknown flag", []string{"--bogus", "x"}, ghpParameters{}, "unknown flag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveParameters(tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("resolveParameters(%v) error = %v, want containing %q", tt.args, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveParameters(%v) = %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("resolveParameters(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestResolveBuildPaths(t *testing.T) {
	proj := minProject(t)
	entry := filepath.Join(proj, "cli")
	writeFile(t, entry, "main.go", "package main\n")

	src, gotEntry, output, err := resolveBuildPaths(ghpParameters{projectPath: proj})
	if err != nil {
		t.Fatalf("resolveBuildPaths(project): %v", err)
	}
	if src != proj {
		t.Errorf("src = %q, want %q", src, proj)
	}
	if gotEntry != proj {
		t.Errorf("entry default = %q, want project root %q", gotEntry, proj)
	}
	if want := filepath.Join(proj, "app"); output != want {
		t.Errorf("output default = %q, want %q", output, want)
	}

	out := filepath.Join(proj, "bin", "server")
	src, gotEntry, output, err = resolveBuildPaths(ghpParameters{projectPath: proj, entryPointDir: entry, output: out})
	if err != nil {
		t.Fatalf("resolveBuildPaths(all): %v", err)
	}
	if gotEntry != entry {
		t.Errorf("entry = %q, want %q", gotEntry, entry)
	}
	if output != out {
		t.Errorf("output = %q, want %q", output, out)
	}

	if _, _, _, err := resolveBuildPaths(ghpParameters{projectPath: t.TempDir()}); err == nil {
		t.Error("resolveBuildPaths(dir without go.mod) should fail")
	}

	noMain := minProject(t)
	if err := os.Remove(filepath.Join(noMain, "main.go")); err != nil {
		t.Fatalf("remove main.go: %v", err)
	}
	if _, _, _, err := resolveBuildPaths(ghpParameters{projectPath: noMain}); err == nil {
		t.Error("resolveBuildPaths(dir without main.go) should fail")
	}
}

func TestResolveBuildPathsDefaultsToCwd(t *testing.T) {
	// The package dir has no go.mod, so the default "." must fail cleanly.
	if _, _, _, err := resolveBuildPaths(ghpParameters{}); err == nil {
		t.Error("resolveBuildPaths() with no args in the package dir should fail")
	}
}

// TestMainExitsWithUsage covers the os.Exit path that is untestable in-process
// (calling main() directly would kill the test runner). It re-executes the test
// binary as a subprocess with a marker env var so the child calls main() for
// real and terminates after printing usage. os.Args[0] is the only way to reach
// main() — the documented exception to the "no os.Args globals" convention.
func TestMainExitsWithUsage(t *testing.T) {
	if os.Getenv("GHP_TEST_MAIN") == "1" {
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainExitsWithUsage$")
	cmd.Env = append(os.Environ(), "GHP_TEST_MAIN=1")
	out, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("main() should os.Exit, got: %v", err)
	}
	if got, want := exitErr.ExitCode(), 2; got != want {
		t.Errorf("main() exit code = %d, want %d (usage)", got, want)
	}
	for _, want := range []string{"Usage:", "build [flags]", "version", "-o, --output"} {
		if !bytes.Contains(out, []byte(want)) {
			t.Errorf("main() output missing %q:\n%s", want, out)
		}
	}
}
