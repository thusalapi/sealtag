package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// repo is a throwaway git repository for tests.
type repo struct {
	t    *testing.T
	root string
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	r := &repo{t: t, root: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// commit writes the files and commits them with the given message.
func (r *repo) commit(msg string, files map[string]string) {
	r.t.Helper()
	for name, body := range files {
		p := filepath.Join(r.root, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
	r.git("add", "-A")
	r.git("commit", "-q", "-m", msg)
}

func (r *repo) plan(bump string) Plan {
	r.t.Helper()
	pkgs, err := detect.Walk(r.root, detect.All())
	if err != nil {
		r.t.Fatal(err)
	}
	p, err := Build(r.root, pkgs, bump)
	if err != nil {
		r.t.Fatal(err)
	}
	return p
}

func find(p Plan, name string) Release {
	for _, r := range p.Releases {
		if r.Package.Name == name {
			return r
		}
	}
	return Release{}
}

func TestAutoBumpSinglePackage(t *testing.T) {
	r := newRepo(t)
	r.commit("chore: init", map[string]string{"package.json": `{"name":"lib","version":"1.2.0"}`})
	r.git("tag", "v1.2.0")

	if got := find(r.plan(Auto), "lib"); got.Next != "" || got.Reason == "" {
		t.Fatalf("no commits since tag: want no release, got %+v", got)
	}

	r.commit("fix: handle nil", map[string]string{"a.js": "1"})
	r.commit("docs: readme", map[string]string{"README.md": "x"})
	got := find(r.plan(Auto), "lib")
	if got.Bump != "patch" || got.Next != "1.2.1" || got.NextTag != "v1.2.1" || len(got.Commits) != 1 {
		t.Fatalf("after fix: %+v", got)
	}

	r.commit("feat(api): add stream()", map[string]string{"b.js": "2"})
	if got := find(r.plan(Auto), "lib"); got.Next != "1.3.0" {
		t.Errorf("after feat: next = %s, want 1.3.0", got.Next)
	}

	r.commit("feat!: remove callback API", map[string]string{"c.js": "3"})
	if got := find(r.plan(Auto), "lib"); got.Next != "2.0.0" {
		t.Errorf("after breaking: next = %s, want 2.0.0", got.Next)
	}
}

func TestBreakingBeforeOneIsMinor(t *testing.T) {
	r := newRepo(t)
	r.commit("feat!: first", map[string]string{"package.json": `{"name":"lib","version":"0.4.2"}`})
	got := find(r.plan(Auto), "lib")
	if got.Next != "0.5.0" || got.Bump != "minor" {
		t.Errorf("got %+v", got)
	}
	if got := find(r.plan("major"), "lib"); got.Next != "1.0.0" {
		t.Errorf("explicit major: next = %s", got.Next)
	}
}

func TestMonorepoScopesCommitsAndTags(t *testing.T) {
	r := newRepo(t)
	r.commit("chore: init", map[string]string{
		"packages/a/package.json": `{"name":"@x/a","version":"1.0.0"}`,
		"packages/b/package.json": `{"name":"@x/b","version":"3.1.0"}`,
		"sdk/go.mod":              "module example.com/sdk\n",
	})
	r.git("tag", "@x/a@v1.0.0")
	r.git("tag", "@x/b@v3.1.0")
	r.git("tag", "sdk/v0.2.0")

	r.commit("feat: a feature", map[string]string{"packages/a/x.js": "1"})
	r.commit("fix: sdk bug", map[string]string{"sdk/s.go": "package sdk\n"})

	p := r.plan(Auto)
	a, b, sdk := find(p, "@x/a"), find(p, "@x/b"), find(p, "example.com/sdk")
	if a.Next != "1.1.0" || a.NextTag != "@x/a@v1.1.0" {
		t.Errorf("a: %+v", a)
	}
	if b.Next != "" {
		t.Errorf("b has no changes but got next %q", b.Next)
	}
	if sdk.Current != "v0.2.0" || sdk.Next != "v0.2.1" || sdk.NextTag != "sdk/v0.2.1" {
		t.Errorf("sdk: %+v", sdk)
	}
}

func TestRootPackageIgnoresNestedPackages(t *testing.T) {
	r := newRepo(t)
	r.commit("chore: init", map[string]string{
		"package.json":        `{"name":"root-lib","version":"1.0.0"}`,
		"plugin/package.json": `{"name":"root-plugin","version":"1.0.0"}`,
	})
	r.git("tag", "root-lib@v1.0.0")
	r.git("tag", "root-plugin@v1.0.0")
	r.commit("feat: plugin only", map[string]string{"plugin/p.js": "1"})
	p := r.plan(Auto)
	if got := find(p, "root-lib"); got.Next != "" {
		t.Errorf("root package should not release for a plugin commit: %+v", got)
	}
	if got := find(p, "root-plugin"); got.Next != "1.1.0" {
		t.Errorf("plugin: %+v", got)
	}
}

func TestAutoNeedsGit(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"lib","version":"1.0.0"}`), 0o644)
	pkgs, _ := detect.Walk(dir, detect.All())
	if _, err := Build(dir, pkgs, Auto); err != ErrNoHistory {
		t.Errorf("err = %v, want ErrNoHistory", err)
	}
	if p, err := Build(dir, pkgs, "minor"); err != nil || p.Releases[0].Next != "1.1.0" {
		t.Errorf("explicit bump without git: %+v %v", p, err)
	}
}

func TestTagPrefix(t *testing.T) {
	cases := []struct {
		p    ecosystem.Package
		mono bool
		want string
	}{
		{ecosystem.Package{Ecosystem: "go", Dir: "."}, true, "v"},
		{ecosystem.Package{Ecosystem: "go", Dir: "sdk/go"}, false, "sdk/go/v"},
		{ecosystem.Package{Ecosystem: "npm", Name: "lib", Dir: "."}, false, "v"},
		{ecosystem.Package{Ecosystem: "npm", Name: "@x/a", Dir: "a"}, true, "@x/a@v"},
	}
	for _, c := range cases {
		if got := TagPrefix(c.p, c.mono); got != c.want {
			t.Errorf("TagPrefix(%+v, %v) = %q, want %q", c.p, c.mono, got, c.want)
		}
	}
}

func TestChangelog(t *testing.T) {
	r := newRepo(t)
	r.commit("chore: init", map[string]string{"package.json": `{"name":"lib","version":"1.0.0"}`})
	r.git("tag", "v1.0.0")
	r.commit("feat(cli): add --json", map[string]string{"a": "1"})
	r.commit("fix: crash on empty input", map[string]string{"b": "1"})
	r.commit("feat!: rename config file", map[string]string{"c": "1"})
	md := Changelog(find(r.plan(Auto), "lib"), "2026-11-03")
	for _, want := range []string{"## lib 2.0.0 (2026-11-03)", "### Breaking changes", "- rename config file", "### Features", "- **cli:** add --json", "### Fixes", "- crash on empty input"} {
		if !strings.Contains(md, want) {
			t.Errorf("changelog missing %q:\n%s", want, md)
		}
	}
}
