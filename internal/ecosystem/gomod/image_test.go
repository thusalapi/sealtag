package gomod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImageFollowsGoDirective(t *testing.T) {
	dir := t.TempDir()
	if got := image(dir); got != "golang:1" {
		t.Errorf("no go.mod: %q", got)
	}
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module m\n\ngo 1.24.3\n"), 0o644)
	if got := image(dir); got != "golang:1.24" {
		t.Errorf("got %q", got)
	}
}
