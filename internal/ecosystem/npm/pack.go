package npm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

type packResult struct {
	Files []struct {
		Path string `json:"path"`
		Size int64  `json:"size"`
	} `json:"files"`
}

// ShippedFiles asks npm itself which files `npm publish` would include. npm applies
// the "files" field, .npmignore, .gitignore and its own always-included and
// never-included lists, so asking npm is the only fully accurate answer.
//
// --ignore-scripts matters: without it, npm runs the package's prepack and prepare
// scripts, which would let a malicious package run code during a "dry run".
func (Ecosystem) ShippedFiles(dir string) ([]ecosystem.File, error) {
	npm, err := exec.LookPath("npm")
	if err != nil {
		return nil, fmt.Errorf("npm: %w", ecosystem.ErrToolMissing)
	}
	cmd := exec.Command(npm, "pack", "--dry-run", "--json", "--ignore-scripts")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("npm pack failed in %s: %s", dir, bytes.TrimSpace(stderr.Bytes()))
		}
		return nil, err
	}
	return parsePack(out)
}

func parsePack(out []byte) ([]ecosystem.File, error) {
	var res []packResult
	if err := json.Unmarshal(out, &res); err != nil {
		return nil, fmt.Errorf("reading npm pack output: %w", err)
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("npm pack returned %d results, want 1", len(res))
	}
	files := make([]ecosystem.File, 0, len(res[0].Files))
	for _, f := range res[0].Files {
		files = append(files, ecosystem.File{Path: f.Path, Size: f.Size})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}
