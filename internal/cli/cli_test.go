package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thusalapi/sealtag/internal/release"
	"github.com/thusalapi/sealtag/internal/testutil"
)

// run executes the CLI in-process and returns stdout and the error.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestDetectJSON(t *testing.T) {
	out, err := run(t, "detect", "--json", "-C", "../../testdata/monorepo")
	if err != nil {
		t.Fatal(err)
	}
	var pkgs []map[string]any
	if err := json.Unmarshal([]byte(out), &pkgs); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(pkgs) != 6 {
		t.Errorf("got %d packages, want 6", len(pkgs))
	}
}

func TestPlanJSON(t *testing.T) {
	out, err := run(t, "plan", "--json", "--bump", "major", "-C", "../../testdata/monorepo")
	if err != nil {
		t.Fatal(err)
	}
	var p release.Plan
	if err := json.Unmarshal([]byte(out), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Releases) != 4 || len(p.Skipped) != 2 {
		t.Fatalf("releases %d skipped %d", len(p.Releases), len(p.Skipped))
	}
	for _, r := range p.Releases {
		if r.Package.Name == "@acme/web" && r.Next != "3.0.0" {
			t.Errorf("@acme/web next = %q, want 3.0.0", r.Next)
		}
	}
}

func TestPlanBadBump(t *testing.T) {
	if _, err := run(t, "plan", "--bump", "huge", "-C", "../../testdata/monorepo"); err == nil {
		t.Error("expected an error for an unknown bump")
	}
}

func TestScanFindsAndAllows(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.js"), []byte("console.log(1)\n"), 0o644)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("X=1\n"), 0o644)

	out, err := run(t, "scan", "-C", dir)
	if !errors.Is(err, ErrFindings) || !strings.Contains(out, "env-file") {
		t.Fatalf("expected env-file finding, err=%v out=%s", err, out)
	}

	os.WriteFile(filepath.Join(dir, ".sealtag-allow"), []byte(".env env-file local dev only, excluded by files field\n"), 0o644)
	out, err = run(t, "scan", "--json", "-C", dir)
	if err != nil {
		t.Fatalf("expected no error after allowlisting, got %v\n%s", err, out)
	}
	var res ScanResult
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 0 || len(res.Allowed) != 1 {
		t.Errorf("findings %d allowed %d", len(res.Findings), len(res.Allowed))
	}
}

func TestCheckNpmPackage(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not installed")
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"pk","version":"1.0.0"}`), 0o644)
	os.WriteFile(filepath.Join(dir, "index.js"), []byte("module.exports = 1\n"), 0o644)
	// No "files" field and no .npmignore: npm would publish this .env file.
	os.WriteFile(filepath.Join(dir, ".env"), []byte("X=1\n"), 0o644)

	out, err := run(t, "check", "--json", "-C", dir)
	if !errors.Is(err, ErrCheckFailed) {
		t.Fatalf("expected check to fail, got %v\n%s", err, out)
	}
	var res []PackageCheck
	if err := json.Unmarshal([]byte(out), &res); err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].Status != "findings" || res[0].Findings[0].File != ".env" {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{0: "0 B", 1023: "1023 B", 1024: "1.0 KiB", 5 << 20: "5.0 MiB"}
	for n, want := range cases {
		if got := humanSize(n); got != want {
			t.Errorf("humanSize(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestPrepareCommitAndTag(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("chore: init", map[string]string{"package.json": "{\n  \"name\": \"lib\",\n  \"version\": \"1.0.0\"\n}\n"})
	r.Git("tag", "v1.0.0")
	r.Commit("feat: new thing", map[string]string{"index.js": "1"})

	r.Write(map[string]string{"scratch.txt": "uncommitted"})
	if _, err := run(t, "prepare", "--commit", "-C", r.Root); !errors.Is(err, ErrDirty) {
		t.Fatalf("dirty tree: err = %v, want ErrDirty", err)
	}
	r.Git("rm", "-q", "--cached", "--ignore-unmatch", "scratch.txt")
	os.Remove(filepath.Join(r.Root, "scratch.txt"))

	out, err := run(t, "prepare", "--commit", "--date", "2026-11-03", "-C", r.Root)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := r.Git("tag", "--points-at", "HEAD"); got != "v1.1.0" {
		t.Errorf("tag on HEAD = %q", got)
	}
	if got := r.Git("log", "-1", "--format=%s"); got != "chore(release): lib 1.1.0" {
		t.Errorf("commit message = %q", got)
	}
	if r.Git("status", "--porcelain") != "" {
		t.Error("tree should be clean after --commit")
	}
	// The release commit is not releasable, so the next plan is empty.
	out, _ = run(t, "plan", "--json", "-C", r.Root)
	var p release.Plan
	json.Unmarshal([]byte(out), &p)
	if p.Releases[0].Next != "" || p.Releases[0].LastTag != "v1.1.0" {
		t.Errorf("after release: %+v", p.Releases[0])
	}
}
