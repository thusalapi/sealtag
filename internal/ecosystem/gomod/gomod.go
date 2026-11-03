// Package gomod detects Go modules defined by go.mod.
//
// Go has no upload step. A version is published by pushing a git tag, so the current
// version is the highest matching tag. Modules in a subdirectory use prefixed tags,
// for example "sdk/v1.2.0".
package gomod

import (
	"bufio"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/semver"
)

// Ecosystem implements ecosystem.Ecosystem for Go modules.
type Ecosystem struct{}

// Name returns "go".
func (Ecosystem) Name() string { return "go" }

// Detect reads dir/go.mod and looks up the latest version tag with git.
func (Ecosystem) Detect(dir string) (*ecosystem.Package, error) {
	path := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	name := ModulePath(data)
	if name == "" {
		return nil, nil
	}
	p := &ecosystem.Package{Ecosystem: "go", Name: name, Dir: dir, Manifest: path}
	if v, ok := latestTag(dir); ok {
		p.Version = v
	} else {
		p.Note = "no version tag yet"
	}
	return p, nil
}

// ModulePath returns the path from the "module" line of a go.mod file.
func ModulePath(gomod []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(gomod))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if rest, ok := strings.CutPrefix(line, "module"); ok && (rest == "" || rest[0] == ' ' || rest[0] == '\t') {
			return strings.Trim(strings.TrimSpace(rest), `"`)
		}
	}
	return ""
}

// latestTag returns the highest semver tag for the module in dir, without its prefix.
// It returns false when dir is not in a git repository or has no matching tags.
func latestTag(dir string) (string, bool) {
	top, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(top, dir)
	if err != nil {
		return "", false
	}
	prefix := ""
	if rel != "." {
		prefix = filepath.ToSlash(rel) + "/"
	}
	out, err := git(dir, "tag", "--list", prefix+"v*")
	if err != nil {
		return "", false
	}
	return HighestTag(strings.Fields(out), prefix)
}

// HighestTag picks the highest valid semver tag that starts with prefix.
func HighestTag(tags []string, prefix string) (string, bool) {
	var best semver.Version
	found := false
	for _, t := range tags {
		v, err := semver.Parse(strings.TrimPrefix(t, prefix))
		if err != nil || !strings.HasPrefix(t, prefix) {
			continue
		}
		if !found || v.Compare(best) > 0 {
			best, found = v, true
		}
	}
	return best.String(), found
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
