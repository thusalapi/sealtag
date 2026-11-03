package cargo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseList(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.WriteFile(filepath.Join(dir, "src", "lib.rs"), []byte("pub fn a() {}\n"), 0o644)
	out := []byte(".cargo_vcs_info.json\nCargo.toml\nCargo.toml.orig\nsrc/lib.rs\n")
	files, err := parseList(dir, out)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 || files[3].Path != "src/lib.rs" || files[3].Size != 14 || files[3].Generated {
		t.Errorf("got %+v", files)
	}
	if !files[0].Generated || !files[2].Generated { // .cargo_vcs_info.json, Cargo.toml.orig
		t.Errorf("generated files not marked: %+v", files)
	}
}

func TestShippedFilesRealCargo(t *testing.T) {
	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo not installed")
	}
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "src"), 0o755)
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"demo\"\nversion = \"0.1.0\"\nedition = \"2021\"\nlicense = \"MIT\"\ndescription = \"demo\"\nexclude = [\"secret.txt\"]\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "lib.rs"), []byte("pub fn a() {}\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("x\n"), 0o644)

	files, err := Ecosystem{}.ShippedFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f.Path] = true
	}
	if !got["src/lib.rs"] || !got["Cargo.toml"] || got["secret.txt"] {
		t.Errorf("unexpected files: %+v", files)
	}
}

func TestDetectWorkspaceInheritance(t *testing.T) {
	p, err := Ecosystem{}.Detect("../../../testdata/monorepo/crates/core")
	if err != nil || p == nil {
		t.Fatalf("detect: %v %v", p, err)
	}
	if p.Version != "0.7.2" || p.Name != "acme-core" {
		t.Errorf("got %+v", p)
	}
	if p, _ := (Ecosystem{}).Detect("../../../testdata/monorepo"); p != nil {
		t.Errorf("virtual workspace root should not be a package: %+v", p)
	}
}
