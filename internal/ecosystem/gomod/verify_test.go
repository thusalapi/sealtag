package gomod

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"testing"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/testutil"
)

// TestVerifyCatchesIgnoredFile shows the bug verify exists for: the module builds in
// the repository, but a file it needs is gitignored, so the published zip is broken.
func TestVerifyCatchesIgnoredFile(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not installed")
	}
	r := testutil.NewRepo(t)
	r.Commit("feat: first", map[string]string{
		"go.mod":     "module example.com/m\n\ngo 1.22\n",
		"m.go":       "package m\n\nfunc A() string { return generated }\n",
		".gitignore": "gen.go\n",
	})
	p := ecosystem.Package{Ecosystem: "go", Name: "example.com/m", Dir: "."}

	if err := (Ecosystem{}).Verify(context.Background(), r.Root, p, ecosystem.VerifyOptions{Log: io.Discard}); err == nil {
		t.Fatal("verify should fail: m.go needs a constant that does not exist")
	}

	r.Write(map[string]string{"gen.go": "package m\n\nconst generated = \"x\"\n"}) // ignored by git
	if out, err := exec.Command("go", "-C", r.Root, "build", "./...").CombinedOutput(); err != nil {
		t.Fatalf("the module should build in the repo: %v\n%s", err, out)
	}
	err := (Ecosystem{}).Verify(context.Background(), r.Root, p, ecosystem.VerifyOptions{Log: io.Discard})
	if err == nil || !strings.Contains(err.Error(), "go build") {
		t.Fatalf("verify should still fail because gen.go is not published, got %v", err)
	}

	r.Commit("fix: commit generated file", map[string]string{".gitignore": ""})
	if err := (Ecosystem{}).Verify(context.Background(), r.Root, p, ecosystem.VerifyOptions{Log: io.Discard}); err != nil {
		t.Errorf("verify should pass once gen.go is tracked: %v", err)
	}
}
