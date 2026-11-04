package pypi

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

func TestTopLevelModules(t *testing.T) {
	files := []ecosystem.File{
		{Path: "demo/__init__.py"}, {Path: "demo/sub/__init__.py"}, {Path: "single.py"},
		{Path: "demo-1.0.dist-info/METADATA"}, {Path: "demo-1.0.data/scripts/x"},
	}
	if got := topLevelModules(files); !slices.Equal(got, []string{"demo", "single"}) {
		t.Errorf("got %v", got)
	}
}

func TestVerifyRealBuild(t *testing.T) {
	if _, err := exec.LookPath("uv"); err != nil || testing.Short() {
		t.Skip("needs uv")
	}
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src", "demo"), 0o755)
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"demo\"\nversion = \"1.0.0\"\n\n[build-system]\nrequires = [\"hatchling\"]\nbuild-backend = \"hatchling.build\"\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "demo", "__init__.py"), []byte("from . import missing\n"), 0o644)
	p := ecosystem.Package{Ecosystem: "pypi", Name: "demo"}
	if err := (Ecosystem{}).Verify(context.Background(), dir, p, io.Discard); err == nil {
		t.Fatal("import of a missing submodule should fail")
	}
	os.WriteFile(filepath.Join(dir, "src", "demo", "missing.py"), []byte("X = 1\n"), 0o644)
	if err := (Ecosystem{}).Verify(context.Background(), dir, p, io.Discard); err != nil {
		t.Errorf("verify should pass: %v", err)
	}
}
