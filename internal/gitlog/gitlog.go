// Package gitlog reads commits and tags from git for one package in a repository.
package gitlog

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/thusalapi/sealtag/internal/commits"
	"github.com/thusalapi/sealtag/internal/semver"
)

// Field and record separators. They cannot appear in normal commit messages.
const (
	fieldSep  = "\x1f"
	recordSep = "\x1e"
)

// Log returns the commits after sinceTag (or all commits when sinceTag is empty)
// that touch dir, newest first. exclude lists folders under dir that belong to other
// packages, so a root package does not pick up its sub-packages' commits.
func Log(root, sinceTag, dir string, exclude []string) ([]commits.Commit, error) {
	args := []string{"log", "--format=%H" + fieldSep + "%s" + fieldSep + "%b" + recordSep}
	if sinceTag != "" {
		args = append(args, sinceTag+"..HEAD")
	}
	args = append(args, "--", pathspec(dir))
	for _, e := range exclude {
		args = append(args, ":(exclude)"+pathspec(e))
	}
	out, err := git(root, args...)
	if err != nil {
		return nil, err
	}
	var cs []commits.Commit
	for _, rec := range strings.Split(out, recordSep) {
		rec = strings.TrimLeft(rec, "\n")
		if rec == "" {
			continue
		}
		parts := strings.SplitN(rec, fieldSep, 3)
		if len(parts) < 3 {
			continue
		}
		cs = append(cs, commits.Parse(parts[0], parts[1], strings.TrimSpace(parts[2])))
	}
	return cs, nil
}

// LatestTag returns the highest semver tag that starts with prefix, and its version.
func LatestTag(root, prefix string) (tag string, version semver.Version, ok bool) {
	out, err := git(root, "tag", "--list", prefix+"*")
	if err != nil {
		return "", semver.Version{}, false
	}
	for _, t := range strings.Fields(out) {
		v, err := semver.Parse(strings.TrimPrefix(t, prefix))
		if err != nil {
			continue
		}
		if !ok || v.Compare(version) > 0 {
			tag, version, ok = t, v, true
		}
	}
	return tag, version, ok
}

// IsRepo reports whether root is inside a git work tree with at least one commit.
func IsRepo(root string) bool {
	_, err := git(root, "rev-parse", "--verify", "HEAD")
	return err == nil
}

func pathspec(dir string) string {
	if dir == "" || dir == "." {
		return "."
	}
	return dir
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}
