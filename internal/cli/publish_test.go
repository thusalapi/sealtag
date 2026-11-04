package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/thusalapi/sealtag/internal/testutil"
)

func TestTokenVars(t *testing.T) {
	env := map[string]string{"NPM_TOKEN": "x", "HOME": "/h"}
	got := tokenVars(func(k string) string { return env[k] })
	if len(got) != 1 || got[0] != "NPM_TOKEN" {
		t.Errorf("got %v", got)
	}
}

func TestPublishRefusesLongLivedToken(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("chore: init", map[string]string{"package.json": `{"name":"lib","version":"1.0.0"}`})
	t.Setenv("CARGO_REGISTRY_TOKEN", "cio-old")
	_, err := run(t, "publish", "--dry-run", "-C", r.Root)
	if !errors.Is(err, ErrLongLivedToken) || !strings.Contains(err.Error(), "CARGO_REGISTRY_TOKEN") {
		t.Errorf("err = %v", err)
	}
}

func TestTaggedTargets(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("chore: init", map[string]string{
		"web/package.json": `{"name":"@x/web","version":"1.2.0"}`,
		"sdk/go.mod":       "module example.com/sdk\n",
		"lib/package.json": `{"name":"@x/lib","version":"3.0.0"}`,
	})
	r.Git("tag", "@x/web@v1.2.0")
	r.Git("tag", "sdk/v0.4.0")
	r.Git("tag", "unrelated")

	targets, err := taggedTargets(r.Root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, tg := range targets {
		got[tg.Package.Name] = tg.Package.Version
	}
	if len(got) != 2 || got["@x/web"] != "1.2.0" || got["example.com/sdk"] != "v0.4.0" {
		t.Errorf("targets: %v", got)
	}

	r.Git("tag", "@x/lib@v3.1.0") // manifest still says 3.0.0
	if _, err := taggedTargets(r.Root); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Errorf("mismatch should fail, got %v", err)
	}
}

func TestPublishDryRunGo(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("feat: first", map[string]string{"go.mod": "module example.com/Mod\n", "m.go": "package mod\n"})
	r.Git("tag", "v0.1.0")
	out, err := run(t, "publish", "--dry-run", "-C", r.Root)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	// Upper-case letters in module paths are escaped as "!" plus lower case.
	if !strings.Contains(out, "proxy.golang.org/example.com/!mod/@v/v0.1.0.info") {
		t.Errorf("output:\n%s", out)
	}
}

func TestPublishBlocksOnSecret(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("feat: first", map[string]string{"go.mod": "module example.com/m\n", "id_rsa": "key"})
	r.Git("tag", "v0.1.0")
	if _, err := run(t, "publish", "--dry-run", "-C", r.Root); !errors.Is(err, ErrCheckFailed) {
		t.Errorf("err = %v, want ErrCheckFailed", err)
	}
}

func TestPublishNothingTagged(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("feat: first", map[string]string{"go.mod": "module example.com/m\n"})
	out, err := run(t, "publish", "--dry-run", "-C", r.Root)
	if err != nil || !strings.Contains(out, "Nothing to publish") {
		t.Errorf("%v %s", err, out)
	}
}

func TestDoctor(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("chore: init", map[string]string{"package.json": `{"name":"lib","version":"1.0.0"}`})
	r.Git("remote", "add", "origin", "git@github.com:octo/lib.git")
	out, err := run(t, "doctor", "-C", r.Root)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[ok]   origin remote is on GitHub (octo/lib)", "[todo] .github/workflows/release.yml exists", "npmjs.com/package/lib/access", "Owner: octo", "Repository: lib"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestPublishOnlyTriggeringTag(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit("feat: first", map[string]string{"go.mod": "module example.com/m\n", "m.go": "package m\n"})
	r.Git("tag", "v0.1.0")
	r.Git("tag", "-a", "v0", "-m", "floating") // not a semver tag: never a target

	out, err := run(t, "publish", "--dry-run", "--skip-verify", "--tag", "v0", "-C", r.Root)
	if err != nil || !strings.Contains(out, "Nothing to publish") {
		t.Errorf("--tag v0 should publish nothing: %v\n%s", err, out)
	}
	out, err = run(t, "publish", "--dry-run", "--skip-verify", "--tag", "v0.1.0", "-C", r.Root)
	if err != nil || !strings.Contains(out, "v0.1.0.info") {
		t.Errorf("--tag v0.1.0 should publish: %v\n%s", err, out)
	}
}
