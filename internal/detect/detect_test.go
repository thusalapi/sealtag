package detect

import (
	"testing"

	"github.com/thusalapi/shipwright/internal/ecosystem"
)

func TestWalkMonorepo(t *testing.T) {
	pkgs, err := Walk("../../testdata/monorepo", All())
	if err != nil {
		t.Fatal(err)
	}
	want := []ecosystem.Package{
		{Ecosystem: "npm", Name: "acme-root", Dir: ".", Private: true},
		{Ecosystem: "cargo", Name: "acme-core", Version: "0.7.2", Dir: "crates/core"},
		{Ecosystem: "cargo", Name: "acme-internal", Version: "0.1.0", Dir: "crates/internal", Private: true},
		{Ecosystem: "go", Name: "github.com/acme/sdk", Dir: "go/sdk"},
		{Ecosystem: "npm", Name: "@acme/web", Version: "2.4.1", Dir: "packages/web"},
		{Ecosystem: "pypi", Name: "acme-client", Dir: "py"},
	}
	if len(pkgs) != len(want) {
		t.Fatalf("got %d packages, want %d: %+v", len(pkgs), len(want), pkgs)
	}
	for i, w := range want {
		g := pkgs[i]
		if g.Ecosystem != w.Ecosystem || g.Name != w.Name || g.Version != w.Version || g.Dir != w.Dir || g.Private != w.Private {
			t.Errorf("package %d:\n got  %+v\n want %+v", i, g, w)
		}
	}
}

func TestSkipDir(t *testing.T) {
	for _, d := range []string{"node_modules", ".git", "target", "testdata"} {
		if !SkipDir(d) {
			t.Errorf("%s should be skipped", d)
		}
	}
	if SkipDir("packages") {
		t.Error("packages should not be skipped")
	}
}
