package gomod

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShippedFilesModuleRules(t *testing.T) {
	dir := t.TempDir()
	w := func(rel, data string) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(data), 0o644)
	}
	w("go.mod", "module example.com/m\n")
	w("m.go", "package m\n")
	w("vendor/x/x.go", "package x\n")             // vendor is never in a module zip
	w("sub/go.mod", "module example.com/m/sub\n") // nested module is excluded
	w("sub/s.go", "package sub\n")

	files, err := Ecosystem{}.ShippedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f.Path] = true
	}
	if !got["go.mod"] || !got["m.go"] || got["vendor/x/x.go"] || got["sub/s.go"] {
		t.Errorf("unexpected files: %+v", files)
	}
}
