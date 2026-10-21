package fsutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestWalkDir(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.txt", "a")
	writeFile(t, root, "sub/b.txt", "b")
	writeFile(t, root, "sub/c.txt", "c")

	var visited []string
	err := WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel == "." {
			return nil // Skip root
		}
		visited = append(visited, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}

	// Order is not guaranteed, but all should be visited
	expected := map[string]bool{"a.txt": true, "sub": true, "sub/b.txt": true, "sub/c.txt": true}
	for _, v := range visited {
		if !expected[v] {
			t.Errorf("unexpected visit: %s", v)
		}
		delete(expected, v)
	}
	if len(expected) > 0 {
		for k := range expected {
			t.Errorf("missing visit: %s", k)
		}
	}
}

func TestWalkFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.txt", "a")
	writeFile(t, root, "sub/b.txt", "b")
	writeFile(t, root, "sub/c.go", "c")
	writeFile(t, root, "ignored.log", "log")

	var visited []string
	err := WalkFiles(root, func(rel string, d fs.DirEntry) bool {
		return filepath.Ext(rel) != ".txt"
	}, func(rel string, d fs.DirEntry) error {
		visited = append(visited, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkFiles: %v", err)
	}

	expected := []string{"a.txt", "sub/b.txt"}
	if len(visited) != len(expected) {
		t.Errorf("visited = %v, want %v", visited, expected)
	}
	for i, v := range visited {
		if v != expected[i] {
			t.Errorf("visited[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestWalkFilesSkipNil(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a.txt", "a")

	var visited []string
	err := WalkFiles(root, nil, func(rel string, d fs.DirEntry) error {
		visited = append(visited, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("WalkFiles: %v", err)
	}
	if len(visited) != 1 || visited[0] != "a.txt" {
		t.Errorf("visited = %v, want [a.txt]", visited)
	}
}

func TestCopyFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src.txt", "hello")

	dst := filepath.Join(root, "dst.txt")
	if err := CopyFile(filepath.Join(root, "src.txt"), dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("dst content = %q, want %q", string(data), "hello")
	}
}

func TestCopyFileCreatesParents(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "src.txt", "hello")

	dst := filepath.Join(root, "sub", "dir", "dst.txt")
	if err := CopyFile(filepath.Join(root, "src.txt"), dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("dst content = %q, want %q", string(data), "hello")
	}
}

func TestCopyTree(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.txt", "a")
	writeFile(t, src, "sub/b.txt", "b")
	writeFile(t, src, "sub/c.go", "c")
	writeFile(t, src, "ignored.log", "log")

	dst := t.TempDir()
	err := CopyTree(src, dst, func(rel string, d fs.DirEntry) bool {
		return filepath.Ext(rel) != ".txt"
	})
	if err != nil {
		t.Fatalf("CopyTree: %v", err)
	}

	// Check copied files
	for _, rel := range []string{"a.txt", "sub/b.txt"} {
		data, err := os.ReadFile(filepath.Join(dst, rel))
		if err != nil {
			t.Errorf("read %s: %v", rel, err)
		}
		expectedContent := filepath.Base(rel)[0:1] // "a" or "b"
		if string(data) != expectedContent {
			t.Errorf("%s content = %q, want %q", rel, string(data), expectedContent)
		}
	}

	// Check ignored files not copied
	for _, rel := range []string{"sub/c.go", "ignored.log"} {
		if _, err := os.Stat(filepath.Join(dst, rel)); !os.IsNotExist(err) {
			t.Errorf("%s should not exist in dst", rel)
		}
	}
}

func TestCopyTreeSkipNil(t *testing.T) {
	src := t.TempDir()
	writeFile(t, src, "a.txt", "a")

	dst := t.TempDir()
	err := CopyTree(src, dst, nil)
	if err != nil {
		t.Fatalf("CopyTree: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "a" {
		t.Errorf("dst content = %q, want %q", string(data), "a")
	}
}

func TestReadFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "test.txt", "content")

	data, err := ReadFile(filepath.Join(root, "test.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("ReadFile = %q, want %q", string(data), "content")
	}
}

func TestReadFileError(t *testing.T) {
	root := t.TempDir()
	_, err := ReadFile(filepath.Join(root, "missing.txt"))
	if err == nil {
		t.Fatal("expected error")
	}
	readErr, ok := err.(*ReadError)
	if !ok {
		t.Fatalf("error type = %T, want *ReadError", err)
	}
	if readErr.Path != filepath.Join(root, "missing.txt") {
		t.Errorf("ReadError.Path = %q, want path", readErr.Path)
	}
}

func TestWriteFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "sub", "out.txt")

	err := WriteFile(path, []byte("data"), 0o644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "data" {
		t.Errorf("WriteFile data = %q, want %q", string(data), "data")
	}
}

func TestWriteFileError(t *testing.T) {
	root := t.TempDir()
	// Create a directory where we want to write a file
	dir := filepath.Join(root, "out.txt")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	err := WriteFile(dir, []byte("data"), 0o644)
	if err == nil {
		t.Fatal("expected error")
	}
	writeErr, ok := err.(*WriteError)
	if !ok {
		t.Fatalf("error type = %T, want *WriteError", err)
	}
	if writeErr.Path != dir {
		t.Errorf("WriteError.Path = %q, want path", writeErr.Path)
	}
}

func TestMkdirAll(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a", "b", "c")

	err := MkdirAll(path, 0o755)
	if err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !info.IsDir() {
		t.Error("path is not a directory")
	}
}

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
