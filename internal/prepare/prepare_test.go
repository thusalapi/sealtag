package prepare

import (
	"strings"
	"testing"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/release"
	"github.com/thusalapi/sealtag/internal/testutil"
)

func plan(t *testing.T, r *testutil.Repo, bump string) release.Plan {
	t.Helper()
	pkgs, err := detect.Walk(r.Root, detect.All())
	if err != nil {
		t.Fatal(err)
	}
	p, err := release.Build(r.Root, pkgs, bump)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestApplyMonorepo(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("chore: init", map[string]string{
		"web/package.json":    "{\n  \"name\": \"@x/web\",\n  \"version\": \"2.4.1\",\n  \"engines\": { \"version\": \"keep\" }\n}\n",
		"Cargo.toml":          "[workspace]\nmembers = [\"core\"]\n\n[workspace.package]\nversion = \"0.7.2\"\n",
		"core/Cargo.toml":     "[package]\nname = \"x-core\"\nversion.workspace = true\n",
		"core/src/lib.rs":     "",
		"py/pyproject.toml":   "[project]\nname = \"x-py\"\ndynamic = [\"version\"]\n\n[tool.hatch.version]\npath = \"x_py/__init__.py\"\n",
		"py/x_py/__init__.py": "__version__ = \"1.0.0\"\n",
		"sdk/go.mod":          "module example.com/sdk\n",
		"web/CHANGELOG.md":    "# Changelog\n\nIntro.\n\n## @x/web 2.4.1 (2026-01-01)\n\n- old entry\n",
	})
	for _, tag := range []string{"@x/web@v2.4.1", "x-core@v0.7.2", "x-py@v1.0.0", "sdk/v0.1.0"} {
		r.Git("tag", tag)
	}
	r.Commit("feat: web feature", map[string]string{"web/a.js": "1"})
	r.Commit("fix: core fix", map[string]string{"core/src/lib.rs": "pub fn a() {}"})
	r.Commit("fix(py): py fix", map[string]string{"py/x_py/a.py": "1"})
	r.Commit("feat!: sdk break", map[string]string{"sdk/s.go": "package sdk"})

	res, err := Apply(r.Root, plan(t, r, release.Auto), "2026-11-03")
	if err != nil {
		t.Fatal(err)
	}

	if got := r.Read("web/package.json"); !strings.Contains(got, `"version": "2.5.0"`) || !strings.Contains(got, `"version": "keep"`) {
		t.Errorf("package.json:\n%s", got)
	}
	if got := r.Read("Cargo.toml"); !strings.Contains(got, `version = "0.7.3"`) {
		t.Errorf("workspace Cargo.toml:\n%s", got)
	}
	if got := r.Read("py/x_py/__init__.py"); got != "__version__ = \"1.0.1\"\n" {
		t.Errorf("__init__.py: %q", got)
	}
	cl := r.Read("web/CHANGELOG.md")
	newIdx, oldIdx := strings.Index(cl, "## @x/web 2.5.0 (2026-11-03)"), strings.Index(cl, "## @x/web 2.4.1")
	if newIdx < 0 || oldIdx < 0 || newIdx > oldIdx || !strings.Contains(cl, "Intro.") {
		t.Errorf("web changelog order wrong:\n%s", cl)
	}
	if !strings.HasPrefix(r.Read("sdk/CHANGELOG.md"), "# Changelog") {
		t.Error("new changelog should get a title")
	}

	wantTags := map[string]bool{"@x/web@v2.5.0": true, "x-core@v0.7.3": true, "x-py@v1.0.1": true, "sdk/v0.2.0": true}
	for _, tg := range res.Tags {
		delete(wantTags, tg)
	}
	if len(wantTags) != 0 {
		t.Errorf("missing tags %v, got %v", wantTags, res.Tags)
	}
	for _, f := range []string{"Cargo.toml", "web/package.json", "py/x_py/__init__.py", "sdk/CHANGELOG.md"} {
		if !contains(res.Changed, f) {
			t.Errorf("Changed is missing %s: %v", f, res.Changed)
		}
	}
	if contains(res.Changed, "sdk/go.mod") {
		t.Error("go.mod must not change: Go versions live in tags")
	}
}

func TestNothingToRelease(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("chore: init", map[string]string{"package.json": `{"name":"a","version":"1.0.0"}`})
	r.Git("tag", "v1.0.0")
	res, err := Apply(r.Root, plan(t, r, release.Auto), "2026-11-03")
	if err != nil || len(res.Releases) != 0 || len(res.Changed) != 0 {
		t.Errorf("got %+v, %v", res, err)
	}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
