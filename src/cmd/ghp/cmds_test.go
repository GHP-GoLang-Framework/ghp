package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
	os.MkdirAll(filepath.Join(dir, "blog"), 0o755)
	os.WriteFile(filepath.Join(dir, "blog", "[slug].ghp"), []byte(`<p><go= r.PathValue("slug") /></p>`), 0o644)

	var buf bytes.Buffer
	if code := Build([]string{dir}, &buf); code != 0 {
		t.Fatalf("Build exit = %d, want 0\nout:\n%s", code, buf.String())
	}

	for _, want := range []string{"Building GHP project...", "GHP project built"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("output missing %q\ngot: %s", want, buf.String())
		}
	}
	for _, name := range []string{"main.go", "go.mod", "pages/index.go", "pages/blog_slug.go"} {
		if _, err := os.Stat(filepath.Join(dir, "build", name)); err != nil {
			t.Errorf("missing generated file build/%s: %v", name, err)
		}
	}
}

func TestBuildDefaultDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
	t.Chdir(dir)

	var buf bytes.Buffer
	if code := Build(nil, &buf); code != 0 {
		t.Fatalf("Build exit = %d, want 0\nout:\n%s", code, buf.String())
	}
	for _, name := range []string{"main.go", "go.mod", "pages/index.go"} {
		if _, err := os.Stat(filepath.Join(dir, "build", name)); err != nil {
			t.Errorf("missing generated file build/%s: %v", name, err)
		}
	}
}

func TestBuildError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.ghp"), []byte(`<go:if x/>`), 0o644)

	var buf bytes.Buffer
	if code := Build([]string{dir}, &buf); code != 1 {
		t.Errorf("Build exit = %d, want 1\nout:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "ghp build: bad.ghp") {
		t.Errorf("output missing file reference\ngot: %s", buf.String())
	}
}

func TestFlagErrors(t *testing.T) {
	tests := []struct {
		name string
		run  func(args []string, stdout io.Writer) int
	}{
		{"build", Build},
		{"dev", Dev},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			code := tt.run([]string{"--bogus"}, &buf)
			if code != 2 {
				t.Errorf("%s exit = %d, want 2\nout:\n%s", tt.name, code, buf.String())
			}
		})
	}
}

func TestDevServesSlugRoute(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
	os.MkdirAll(filepath.Join(dir, "blog"), 0o755)
	os.WriteFile(filepath.Join(dir, "blog", "[slug].ghp"), []byte(`<p>Slug: <go= r.PathValue("slug") /></p>`), 0o644)

	port := freePort(t)
	t.Setenv("GHP_PORT", port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var buf bytes.Buffer
	done := make(chan int, 1)
	go func() { done <- runDev(ctx, []string{dir}, &buf) }()

	base := "http://127.0.0.1:" + port
	waitFor(t, base+"/")

	resp, err := http.Get(base + "/blog/ola")
	if err != nil {
		t.Fatalf("GET /blog/ola: %v", err)
	}
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /blog/ola status = %d, want 200", resp.StatusCode)
	}
	if !strings.Contains(body, "Slug: ola") {
		t.Errorf("GET /blog/ola body missing the slug\ngot: %s", body)
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("runDev exit = %d, want 0\nout:\n%s", code, buf.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("runDev did not stop after cancel\nout:\n%s", buf.String())
	}
}

func TestDevReloadsOnPageChange(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "index.ghp")
	os.WriteFile(page, []byte("<h1>Version 1</h1>"), 0o644)

	port := freePort(t)
	t.Setenv("GHP_PORT", port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var buf bytes.Buffer
	done := make(chan int, 1)
	go func() { done <- runDev(ctx, []string{dir}, &buf) }()

	base := "http://127.0.0.1:" + port
	waitFor(t, base+"/")
	waitForBody(t, base+"/", "Version 1")

	os.WriteFile(page, []byte("<h1>Version 2</h1>"), 0o644)
	waitForBody(t, base+"/", "Version 2")

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("runDev exit = %d, want 0\nout:\n%s", code, buf.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("runDev did not stop after cancel\nout:\n%s", buf.String())
	}
}

func TestDevKeepsOldServerOnBrokenEdit(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "index.ghp")
	os.WriteFile(page, []byte("<h1>Good</h1>"), 0o644)

	port := freePort(t)
	t.Setenv("GHP_PORT", port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var log syncBuf
	done := make(chan int, 1)
	go func() { done <- runDev(ctx, []string{dir}, &log) }()

	base := "http://127.0.0.1:" + port
	waitFor(t, base+"/")
	waitForBody(t, base+"/", "Good")

	os.WriteFile(page, []byte(`<go:if broken/>`), 0o644)
	waitForLog(t, &log, "keeping the current server running")
	waitForBody(t, base+"/", "Good")

	os.WriteFile(page, []byte("<h1>Fixed</h1>"), 0o644)
	waitForBody(t, base+"/", "Fixed")

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("runDev exit = %d, want 0\nout:\n%s", code, log.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("runDev did not stop after cancel\nout:\n%s", log.String())
	}
}

func TestParseDirRejectsExtraArgs(t *testing.T) {
	var buf bytes.Buffer
	if code := Build([]string{"a", "b"}, &buf); code != 2 {
		t.Errorf("Build exit = %d, want 2\nout:\n%s", code, buf.String())
	}
	if !strings.Contains(buf.String(), "too many arguments") {
		t.Errorf("output missing usage hint\ngot: %s", buf.String())
	}
}

func TestBuildWriteError(t *testing.T) {
	tests := []struct {
		name string
		dir  func(t *testing.T) string
	}{
		{
			name: "build dir is a file",
			dir: func(t *testing.T) string {
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
				os.WriteFile(filepath.Join(dir, "build"), []byte("in the way"), 0o644)
				return dir
			},
		},
		{
			name: "generated page path is a directory",
			dir: func(t *testing.T) string {
				dir := t.TempDir()
				os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
				os.MkdirAll(filepath.Join(dir, "build", "pages", "index.go"), 0o755)
				return dir
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if code := Build([]string{tt.dir(t)}, &buf); code != 1 {
				t.Fatalf("Build exit = %d, want 1\nout:\n%s", code, buf.String())
			}
		})
	}
}

func TestDevMkdirTempError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.ghp"), []byte("<h1>Home</h1>"), 0o644)
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	if code := runDev(ctx, []string{dir}, &buf); code != 1 {
		t.Fatalf("runDev exit = %d, want 1\nout:\n%s", code, buf.String())
	}
}

func TestDevBuildError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.ghp"), []byte(`<go:if x/>`), 0o644)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var buf bytes.Buffer
	if code := runDev(ctx, []string{dir}, &buf); code != 1 {
		t.Fatalf("runDev exit = %d, want 1\nout:\n%s", code, buf.String())
	}
}

func TestStartAppMissingBinary(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := startApp(ctx, filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("startApp on a missing binary should fail")
	}
}

func TestEqualSnapshots(t *testing.T) {
	base := map[string]int64{"index.ghp": 1}
	if !equalSnapshots(base, map[string]int64{"index.ghp": 1}) {
		t.Error("identical snapshots should be equal")
	}
	if equalSnapshots(base, map[string]int64{"index.ghp": 2}) {
		t.Error("snapshots with a changed mtime should differ")
	}
	if equalSnapshots(base, map[string]int64{"index.ghp": 1, "about.ghp": 3}) {
		t.Error("snapshots with a different page count should differ")
	}
}

// syncBuf is a bytes.Buffer safe for concurrent use, so a test can read
// the dev server's log while it keeps writing to it.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuf) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

// waitForLog polls w until its content contains want.
func waitForLog(t *testing.T, w fmt.Stringer, want string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if strings.Contains(w.String(), want) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("output never contained %q\ngot: %s", want, w.String())
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// freePort asks the OS for an unused TCP port and hands it back (with a
// small race window between closing and the child binding it, which is
// acceptable for tests).
func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

// waitFor polls url until the server answers with 200, failing after a
// timeout so the test reports a clear error instead of hanging.
func waitFor(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never answered %s", url)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b := new(bytes.Buffer)
	if _, err := b.ReadFrom(resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b.String()
}

// waitForBody polls url until its body contains want, so reload tests can
// wait for the running server to pick up a page change.
func waitForBody(t *testing.T, url, want string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		if resp, err := http.Get(url); err == nil {
			if body := readBody(t, resp); strings.Contains(body, want) {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("body of %s never contained %q", url, want)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
