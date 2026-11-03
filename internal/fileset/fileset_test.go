package fileset

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func write(t *testing.T, root, rel, data string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func skip(n string) bool { return n == "node_modules" || n == ".git" }

func TestWalkFallback(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.txt", "a")
	write(t, root, "src/b.go", "b")
	write(t, root, "node_modules/x/c.js", "c")
	got, err := List(root, skip)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a.txt", "src/b.go"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestGitRespectsIgnore(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	write(t, root, ".gitignore", "dist/\n*.log\n")
	write(t, root, "main.go", "package main")
	write(t, root, "dist/bundle.js", "x")
	write(t, root, "debug.log", "x")
	write(t, root, ".env", "SECRET=1") // untracked but not ignored: must be listed
	got, err := List(root, skip)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{".env", ".gitignore", "main.go"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
