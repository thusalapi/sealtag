// Package testutil holds helpers shared by tests.
package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Repo is a throwaway git repository in a temporary folder.
type Repo struct {
	T    testing.TB
	Root string
}

// NewRepo creates an empty repository on branch main. The test is skipped when git
// is not installed.
func NewRepo(t testing.TB) *Repo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	r := &Repo{T: t, Root: t.TempDir()}
	r.Git("init", "-q", "-b", "main")
	// Local identity so commits work on machines without a global git config.
	r.Git("config", "user.name", "sealtag-test")
	r.Git("config", "user.email", "test@example.com")
	r.Git("config", "commit.gpgsign", "false")
	r.Git("config", "tag.gpgsign", "false")
	r.Git("config", "core.autocrlf", "false")
	return r
}

// Git runs a git command in the repository and returns its trimmed output.
func (r *Repo) Git(args ...string) string {
	r.T.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Root
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.T.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// Write creates or replaces files, relative to the repository root.
func (r *Repo) Write(files map[string]string) {
	r.T.Helper()
	for name, body := range files {
		p := filepath.Join(r.Root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			r.T.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			r.T.Fatal(err)
		}
	}
}

// Commit writes the files and commits everything with the given message.
func (r *Repo) Commit(msg string, files map[string]string) {
	r.T.Helper()
	r.Write(files)
	r.Git("add", "-A")
	r.Git("commit", "-q", "-m", msg)
}

// Read returns a file's content.
func (r *Repo) Read(name string) string {
	r.T.Helper()
	b, err := os.ReadFile(filepath.Join(r.Root, filepath.FromSlash(name)))
	if err != nil {
		r.T.Fatal(err)
	}
	return string(b)
}
