// Package cargo detects Rust crates defined by Cargo.toml.
package cargo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

const noteInherited = "version inherited from workspace"

// Ecosystem implements ecosystem.Ecosystem for crates.io.
type Ecosystem struct{}

// Name returns "cargo".
func (Ecosystem) Name() string { return "cargo" }

type manifest struct {
	Package   map[string]any `toml:"package"`
	Workspace struct {
		Package struct {
			Version string `toml:"version"`
		} `toml:"package"`
	} `toml:"workspace"`
	hasWorkspace bool
}

func load(path string) (*manifest, error) {
	var m manifest
	md, err := toml.DecodeFile(path, &m)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	m.hasWorkspace = md.IsDefined("workspace")
	return &m, nil
}

// Detect reads dir/Cargo.toml. A virtual workspace root (no [package]) returns nil.
func (Ecosystem) Detect(dir string) (*ecosystem.Package, error) {
	path := filepath.Join(dir, "Cargo.toml")
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	m, err := load(path)
	if err != nil {
		return nil, err
	}
	if m.Package == nil {
		return nil, nil
	}
	name, _ := m.Package["name"].(string)
	p := &ecosystem.Package{Ecosystem: "cargo", Name: name, Dir: dir, Manifest: path}

	// version = "1.2.3"  or  version.workspace = true
	switch v := m.Package["version"].(type) {
	case string:
		p.Version = v
	case map[string]any:
		if inherit, _ := v["workspace"].(bool); inherit {
			ver, root := workspaceVersion(dir)
			p.Version = ver
			if root != "" {
				p.Note = noteInherited
			} else {
				p.Note = "version.workspace = true but no workspace root found"
			}
		}
	case nil:
		// Since Cargo 1.75 a missing version means 0.0.0 and publish = false.
		p.Version = "0.0.0"
		p.Private = true
	}

	// publish = false  or  publish = []  means "never publish".
	switch v := m.Package["publish"].(type) {
	case bool:
		p.Private = p.Private || !v
	case []any:
		p.Private = p.Private || len(v) == 0
	}
	return p, nil
}

// workspaceRoot returns the folder of the nearest Cargo.toml above or at dir that
// declares [workspace], or "" when the crate is not in a workspace.
func workspaceRoot(dir string) string {
	cur := dir
	for {
		if m, err := load(filepath.Join(cur, "Cargo.toml")); err == nil && m.hasWorkspace {
			return cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

// workspaceVersion walks up from dir looking for [workspace.package].version.
func workspaceVersion(dir string) (version, root string) {
	cur := filepath.Dir(dir)
	for {
		path := filepath.Join(cur, "Cargo.toml")
		if m, err := load(path); err == nil && m.hasWorkspace {
			return m.Workspace.Package.Version, cur
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", ""
		}
		cur = parent
	}
}
