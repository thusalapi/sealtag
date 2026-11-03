// Package release works out what the next release of each package should be:
// which tag marks the last release, which commits came after it, and which version
// comes next.
package release

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/thusalapi/sealtag/internal/commits"
	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/gitlog"
	"github.com/thusalapi/sealtag/internal/semver"
)

// Auto asks sealtag to choose the bump from Conventional Commits.
const Auto = "auto"

var registries = map[string]string{
	"npm":   "registry.npmjs.org",
	"pypi":  "pypi.org",
	"cargo": "crates.io",
	"go":    "git tag (proxy.golang.org)",
}

// Release is the plan for one package.
type Release struct {
	Package   ecosystem.Package `json:"package"`
	Target    string            `json:"target"`
	TagPrefix string            `json:"tag_prefix"`
	LastTag   string            `json:"last_tag,omitempty"`
	Current   string            `json:"current"`
	Bump      string            `json:"bump"`           // "", "patch", "minor" or "major"
	Next      string            `json:"next,omitempty"` // empty when nothing to release or unknown
	NextTag   string            `json:"next_tag,omitempty"`
	Reason    string            `json:"reason,omitempty"` // why there is no release, or a caveat
	Commits   []commits.Commit  `json:"commits"`          // releasable commits since LastTag
}

// Plan is the full dry-run result.
type Plan struct {
	Bump     string              `json:"bump"`
	Releases []Release           `json:"releases"`
	Skipped  []ecosystem.Package `json:"skipped"`
}

// ErrNoHistory means automatic bumps were asked for outside a git repository.
var ErrNoHistory = errors.New("automatic bumps need git history: run inside a git repository with at least one commit, or pass --bump major|minor|patch")

// TagPrefix returns the tag prefix used for a package's releases. See ADR 0006.
//
//   - Go modules must use "<dir>/v" (or "v" at the root). The Go tooling requires it.
//   - In a repository with one publishable package, others use plain "v".
//   - In a monorepo, others use "<name>@v", for example "@acme/web@v2.4.1".
func TagPrefix(p ecosystem.Package, monorepo bool) string {
	if p.Ecosystem == "go" {
		if p.Dir == "." || p.Dir == "" {
			return "v"
		}
		return p.Dir + "/v"
	}
	if !monorepo {
		return "v"
	}
	return p.Name + "@v"
}

// Build creates the plan. bump is Auto or an explicit semver bump.
func Build(root string, pkgs []ecosystem.Package, bump string) (Plan, error) {
	plan := Plan{Bump: bump, Releases: []Release{}, Skipped: []ecosystem.Package{}}
	var public []ecosystem.Package
	for _, p := range pkgs {
		if p.Private {
			plan.Skipped = append(plan.Skipped, p)
		} else {
			public = append(public, p)
		}
	}
	hasGit := gitlog.IsRepo(root)
	if bump == Auto && !hasGit && len(public) > 0 {
		return Plan{}, ErrNoHistory
	}
	for _, p := range public {
		r, err := plan1(root, p, public, bump, hasGit)
		if err != nil {
			return Plan{}, fmt.Errorf("%s: %w", p.Name, err)
		}
		plan.Releases = append(plan.Releases, r)
	}
	return plan, nil
}

func plan1(root string, p ecosystem.Package, all []ecosystem.Package, bump string, hasGit bool) (Release, error) {
	r := Release{Package: p, Target: registries[p.Ecosystem], TagPrefix: TagPrefix(p, len(all) > 1), Commits: []commits.Commit{}}

	var tagVersion semver.Version
	hasTag := false
	if hasGit {
		r.LastTag, tagVersion, hasTag = gitlog.LatestTag(root, r.TagPrefix)
	}

	// The manifest is the source of truth for the current version. Go has no
	// manifest version, so its current version is the last tag.
	r.Current = p.Version
	if p.Ecosystem == "go" {
		r.Current = "v0.0.0"
		if hasTag {
			tagVersion.VPrefix = true // Go versions are always written with a "v"
			r.Current = tagVersion.String()
		}
	}
	cur, err := semver.Parse(r.Current)
	if err != nil {
		r.Reason = "current version unknown: " + firstNonEmpty(p.Note, "no version in the manifest")
		return r, nil
	}
	if hasTag && p.Ecosystem != "go" && strings.TrimPrefix(tagVersion.String(), "v") != strings.TrimPrefix(cur.String(), "v") {
		r.Reason = fmt.Sprintf("manifest says %s but the last tag is %s", cur, r.LastTag)
	}

	if hasGit {
		cs, err := gitlog.Log(root, r.LastTag, p.Dir, nestedDirs(p, all))
		if err != nil {
			return r, err
		}
		r.Commits = nonNil(commits.Releasable(cs))
	}

	r.Bump = bump
	if bump == Auto {
		r.Bump = commits.Bump(r.Commits)
		if r.Bump == commits.None {
			since := "the start of history"
			if r.LastTag != "" {
				since = r.LastTag
			}
			r.Reason = "no feat, fix, perf or breaking commits since " + since
			return r, nil
		}
		// Before 1.0.0 a breaking change bumps the minor version, the convention used
		// by Cargo and most release tools. Going to 1.0.0 is a deliberate choice:
		// pass --bump major.
		if r.Bump == commits.Major && cur.Major == 0 {
			r.Bump = commits.Minor
			r.Reason = "breaking change before 1.0.0: bumping minor"
		}
	}
	next, err := cur.Bump(r.Bump)
	if err != nil {
		return r, err
	}
	r.Next = next.String()
	r.NextTag = r.TagPrefix + strings.TrimPrefix(next.String(), "v")
	return r, nil
}

// nestedDirs returns the folders of other packages that sit inside p's folder.
func nestedDirs(p ecosystem.Package, all []ecosystem.Package) []string {
	var out []string
	for _, o := range all {
		if o.Dir == p.Dir {
			continue
		}
		if p.Dir == "." || strings.HasPrefix(o.Dir, p.Dir+"/") {
			out = append(out, path.Clean(o.Dir))
		}
	}
	return out
}

func nonNil(cs []commits.Commit) []commits.Commit {
	if cs == nil {
		return []commits.Commit{}
	}
	return cs
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
