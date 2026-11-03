// Package npm detects packages defined by package.json.
package npm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Ecosystem implements ecosystem.Ecosystem for npm.
type Ecosystem struct{}

// Name returns "npm".
func (Ecosystem) Name() string { return "npm" }

type manifest struct {
	Name       string          `json:"name"`
	Version    string          `json:"version"`
	Private    bool            `json:"private"`
	Workspaces json.RawMessage `json:"workspaces"`
}

// Detect reads dir/package.json.
func (Ecosystem) Detect(dir string) (*ecosystem.Package, error) {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	// A workspace root without a name is only a container for other packages.
	if m.Name == "" {
		return nil, nil
	}
	p := &ecosystem.Package{
		Ecosystem: "npm",
		Name:      m.Name,
		Version:   m.Version,
		Dir:       dir,
		Manifest:  path,
		Private:   m.Private,
	}
	if m.Private && len(m.Workspaces) > 0 {
		p.Note = "workspace root"
	}
	return p, nil
}
