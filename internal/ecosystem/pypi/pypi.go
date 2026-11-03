// Package pypi detects Python packages defined by pyproject.toml.
package pypi

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"

	"github.com/thusalapi/shipwright/internal/ecosystem"
)

// Ecosystem implements ecosystem.Ecosystem for PyPI.
type Ecosystem struct{}

// Name returns "pypi".
func (Ecosystem) Name() string { return "pypi" }

type manifest struct {
	Project struct {
		Name       string   `toml:"name"`
		Version    string   `toml:"version"`
		Dynamic    []string `toml:"dynamic"`
		Classifier []string `toml:"classifiers"`
	} `toml:"project"`
	Tool struct {
		Poetry struct {
			Name    string `toml:"name"`
			Version string `toml:"version"`
		} `toml:"poetry"`
	} `toml:"tool"`
}

// Detect reads dir/pyproject.toml. It supports the standard [project] table and the
// older [tool.poetry] table.
func (Ecosystem) Detect(dir string) (*ecosystem.Package, error) {
	path := filepath.Join(dir, "pyproject.toml")
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	p := &ecosystem.Package{Ecosystem: "pypi", Dir: dir, Manifest: path}
	switch {
	case m.Project.Name != "":
		p.Name, p.Version = m.Project.Name, m.Project.Version
		if slices.Contains(m.Project.Dynamic, "version") {
			p.Note = "version is dynamic (set by the build backend)"
		}
	case m.Tool.Poetry.Name != "":
		p.Name, p.Version = m.Tool.Poetry.Name, m.Tool.Poetry.Version
	default:
		// pyproject.toml used only for tool config (ruff, black, ...).
		return nil, nil
	}
	// PyPI rejects uploads with this classifier, so treat it as private.
	if slices.Contains(m.Project.Classifier, "Private :: Do Not Upload") {
		p.Private = true
	}
	return p, nil
}
