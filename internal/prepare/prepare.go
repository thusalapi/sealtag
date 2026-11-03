// Package prepare applies a release plan to the working tree: it writes the new
// versions into manifests and the release notes into CHANGELOG.md files. It does not
// publish anything. Committing and tagging are separate, explicit steps.
package prepare

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/release"
)

// Result describes what Apply changed.
type Result struct {
	Releases []release.Release `json:"releases"` // only the ones that are released
	Changed  []string          `json:"changed"`  // files written, relative to root
	Tags     []string          `json:"tags"`     // tags to create after committing
}

const changelogHeader = "# Changelog\n\nAll notable changes to this package are documented here. This file is updated by sealtag.\n"

// Apply writes versions and changelogs for every release in plan that has a next
// version. date is the release date (YYYY-MM-DD).
func Apply(root string, plan release.Plan, date string) (Result, error) {
	res := Result{Releases: []release.Release{}, Changed: []string{}, Tags: []string{}}
	if err := checkSharedVersions(plan); err != nil {
		return res, err
	}
	changed := map[string]bool{}
	for _, r := range plan.Releases {
		if r.Next == "" {
			continue
		}
		res.Releases = append(res.Releases, r)
		res.Tags = append(res.Tags, r.NextTag)

		// Go has no version file: the tag is the release.
		if v, ok := detect.ByName(r.Package.Ecosystem).(ecosystem.Versioner); ok {
			files, err := v.SetVersion(root, r.Package, strings.TrimPrefix(r.Next, "v"))
			if err != nil {
				return res, fmt.Errorf("%s: setting version: %w", r.Package.Name, err)
			}
			for _, f := range files {
				changed[f] = true
			}
		}
		notes := release.Changelog(r, date)
		cl := path.Join(r.Package.Dir, "CHANGELOG.md")
		if err := prependChangelog(filepath.Join(root, filepath.FromSlash(cl)), notes); err != nil {
			return res, fmt.Errorf("%s: writing changelog: %w", r.Package.Name, err)
		}
		changed[cl] = true
	}
	for f := range changed {
		res.Changed = append(res.Changed, f)
	}
	sort.Strings(res.Changed)
	return res, nil
}

// checkSharedVersions stops two crates that inherit one workspace version from being
// given different versions in the same release.
func checkSharedVersions(plan release.Plan) error {
	seen := map[string]string{}
	for _, r := range plan.Releases {
		if r.Next == "" || r.Package.Ecosystem != "cargo" || !strings.Contains(r.Package.Note, "inherited") {
			continue
		}
		if prev, ok := seen["cargo-workspace"]; ok && prev != r.Next {
			return fmt.Errorf("crates sharing the workspace version would get different versions (%s and %s): pass one --bump for all of them", prev, r.Next)
		}
		seen["cargo-workspace"] = r.Next
	}
	return nil
}

// prependChangelog inserts notes below the file's title, creating the file if needed.
func prependChangelog(path, notes string) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		data = []byte(changelogHeader)
	} else if err != nil {
		return err
	}
	// Insert before the first existing "## " release heading, or at the end.
	idx := bytes.Index(data, []byte("\n## "))
	var out []byte
	if idx < 0 {
		out = append(bytes.TrimRight(data, "\n"), []byte("\n\n"+notes)...)
	} else {
		out = append(append([]byte{}, data[:idx+1]...), []byte(notes+"\n")...)
		out = append(out, data[idx+1:]...)
	}
	return os.WriteFile(path, out, 0o644)
}
