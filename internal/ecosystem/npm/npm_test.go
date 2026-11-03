package npm

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParsePack(t *testing.T) {
	out := []byte(`[{"id":"pk@1.0.0","files":[{"path":"package.json","size":47},{"path":"lib/a.js","size":2}]}]`)
	files, err := parsePack(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Path != "lib/a.js" || files[1].Size != 47 {
		t.Errorf("got %+v", files)
	}
	if _, err := parsePack([]byte(`[]`)); err == nil {
		t.Error("empty result should fail")
	}
}

// TestShippedFilesRealNpm runs the real npm to check that the "files" field is honoured.
func TestShippedFilesRealNpm(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not installed")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"pk","version":"1.0.0","files":["lib"]}`), 0o644)
	os.MkdirAll(filepath.Join(dir, "lib"), 0o755)
	os.WriteFile(filepath.Join(dir, "lib", "a.js"), []byte("x\n"), 0o644)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=1\n"), 0o644) // not in "files"

	files, err := Ecosystem{}.ShippedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f.Path] = true
	}
	if !got["lib/a.js"] || !got["package.json"] || got[".env"] {
		t.Errorf("unexpected file list: %+v", files)
	}
}
