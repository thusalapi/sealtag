// Package commits parses Conventional Commits (https://www.conventionalcommits.org)
// and decides which semver bump a set of commits needs.
//
//	feat(parser): add arrays          -> minor
//	fix: handle empty input           -> patch
//	feat!: drop Node 16               -> major
//	chore: update deps                -> no release
package commits

import (
	"regexp"
	"strings"
)

// Commit is one parsed commit message.
type Commit struct {
	Hash        string `json:"hash"`
	Type        string `json:"type"`            // "feat", "fix", ... or "" if not conventional
	Scope       string `json:"scope,omitempty"` // "parser" in feat(parser): ...
	Breaking    bool   `json:"breaking"`
	Description string `json:"description"`
	Subject     string `json:"subject"` // the original first line
}

var header = regexp.MustCompile(`^([a-zA-Z]+)(?:\(([^()]*)\))?(!)?:\s+(.+)$`)

// Parse reads a commit subject (first line) and body.
func Parse(hash, subject, body string) Commit {
	c := Commit{Hash: hash, Subject: subject, Description: subject}
	m := header.FindStringSubmatch(strings.TrimSpace(subject))
	if m == nil {
		return c
	}
	c.Type = strings.ToLower(m[1])
	c.Scope = m[2]
	c.Breaking = m[3] == "!"
	c.Description = m[4]
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "BREAKING CHANGE:") || strings.HasPrefix(line, "BREAKING-CHANGE:") {
			c.Breaking = true
		}
	}
	return c
}

// Bump levels, ordered so that a larger value wins.
const (
	None  = ""
	Patch = "patch"
	Minor = "minor"
	Major = "major"
)

var rank = map[string]int{None: 0, Patch: 1, Minor: 2, Major: 3}

// BumpFor returns the bump a single commit asks for.
func BumpFor(c Commit) string {
	switch {
	case c.Breaking:
		return Major
	case c.Type == "feat":
		return Minor
	case c.Type == "fix" || c.Type == "perf":
		return Patch
	}
	return None // docs, chore, ci, test, refactor, style, build, or not conventional
}

// Bump returns the largest bump any of the commits needs, or None.
func Bump(cs []Commit) string {
	best := None
	for _, c := range cs {
		if b := BumpFor(c); rank[b] > rank[best] {
			best = b
		}
	}
	return best
}

// Releasable returns only the commits that cause a release, keeping their order.
func Releasable(cs []Commit) []Commit {
	var out []Commit
	for _, c := range cs {
		if BumpFor(c) != None {
			out = append(out, c)
		}
	}
	return out
}
