package pypi

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writeSdist(t *testing.T, path string, members map[string]string) {
	t.Helper()
	f, _ := os.Create(path)
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for name, body := range members {
		tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg})
		tw.Write([]byte(body))
	}
	tw.Close()
	gz.Close()
}

func TestReadSdist(t *testing.T) {
	p := filepath.Join(t.TempDir(), "demo-1.0.0.tar.gz")
	writeSdist(t, p, map[string]string{"demo-1.0.0/PKG-INFO": "Name: demo\n", "demo-1.0.0/src/demo/__init__.py": "x = 1\n"})
	files, err := readSdist(p)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, f := range files {
		got[f.Path] = string(f.Data)
		if f.Artifact != "demo-1.0.0.tar.gz" {
			t.Errorf("artifact = %q", f.Artifact)
		}
	}
	if got["src/demo/__init__.py"] != "x = 1\n" || got["PKG-INFO"] == "" {
		t.Errorf("got %v", got)
	}
}

func TestReadWheel(t *testing.T) {
	p := filepath.Join(t.TempDir(), "demo-1.0.0-py3-none-any.whl")
	f, _ := os.Create(p)
	zw := zip.NewWriter(f)
	w, _ := zw.Create("demo/__init__.py")
	w.Write([]byte("x = 1\n"))
	zw.Close()
	f.Close()
	files, err := readWheel(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "demo/__init__.py" || files[0].Size != 6 {
		t.Errorf("got %+v", files)
	}
}

func TestDetect(t *testing.T) {
	p, err := Ecosystem{}.Detect("../../../testdata/monorepo/py")
	if err != nil || p == nil || p.Name != "acme-client" || p.Version != "0.3.0" || p.VersionFile == "" {
		t.Fatalf("got %+v, %v", p, err)
	}
	if p, _ := (Ecosystem{}).Detect("../../../testdata/monorepo/tools"); p != nil {
		t.Errorf("tool-config-only pyproject should be ignored: %+v", p)
	}
}

// TestShippedFilesRealBuild builds a real project. It shows why sealtag exists:
// without a .gitignore, hatchling puts a stray .env file into the sdist.
func TestShippedFilesRealBuild(t *testing.T) {
	if _, err := exec.LookPath("uv"); err != nil {
		t.Skip("uv not installed")
	}
	if testing.Short() {
		t.Skip("builds a package, skipped with -short")
	}
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src", "demo"), 0o755)
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"demo\"\nversion = \"1.0.0\"\n\n[build-system]\nrequires = [\"hatchling\"]\nbuild-backend = \"hatchling.build\"\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "demo", "__init__.py"), []byte("x = 1\n"), 0o644)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("SECRET=1\n"), 0o644)

	files, err := Ecosystem{}.ShippedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	inSdist, inWheel := map[string]bool{}, map[string]bool{}
	for _, f := range files {
		if filepath.Ext(f.Artifact) == ".whl" {
			inWheel[f.Path] = true
		} else {
			inSdist[f.Path] = true
		}
	}
	if !inSdist[".env"] {
		t.Errorf("expected hatchling to include .env in the sdist: %v", inSdist)
	}
	if !inWheel["demo/__init__.py"] {
		t.Errorf("wheel missing module: %v", inWheel)
	}
}
