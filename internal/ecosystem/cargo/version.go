package cargo

import (
	"fmt"
	"path/filepath"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/textedit"
)

// SetVersion updates [package].version in Cargo.toml. A crate that inherits its
// version from the workspace gets [workspace.package].version updated instead, which
// changes the version of every crate that inherits it.
func (Ecosystem) SetVersion(root string, p ecosystem.Package, version string) ([]string, error) {
	manifest, table := p.Manifest, "package"
	if p.Note == noteInherited {
		_, wsRoot := workspaceVersion(filepath.Join(root, filepath.FromSlash(p.Dir)))
		rel, err := filepath.Rel(root, filepath.Join(wsRoot, "Cargo.toml"))
		if err != nil {
			return nil, fmt.Errorf("locating workspace root: %w", err)
		}
		manifest, table = filepath.ToSlash(rel), "workspace.package"
	}
	ok, err := ecosystem.EditFile(root, manifest, func(b []byte) ([]byte, error) {
		return textedit.SetTOMLString(b, table, "version", version)
	})
	if err != nil || !ok {
		return nil, err
	}
	return []string{manifest}, nil
}
