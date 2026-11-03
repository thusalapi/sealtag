// Package detect walks a repository and collects every package that any supported
// ecosystem recognises.
package detect

import (
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/thusalapi/shipwright/internal/ecosystem"
	"github.com/thusalapi/shipwright/internal/ecosystem/cargo"
	"github.com/thusalapi/shipwright/internal/ecosystem/gomod"
	"github.com/thusalapi/shipwright/internal/ecosystem/npm"
	"github.com/thusalapi/shipwright/internal/ecosystem/pypi"
)

// All returns every supported ecosystem.
func All() []ecosystem.Ecosystem {
	return []ecosystem.Ecosystem{npm.Ecosystem{}, pypi.Ecosystem{}, cargo.Ecosystem{}, gomod.Ecosystem{}}
}

// SkipDir reports whether a directory below the root should never be searched.
// These hold dependencies, build output or test fixtures, not the user's packages.
func SkipDir(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", "node_modules", "vendor", "target", "dist", "build",
		".venv", "venv", "__pycache__", ".tox", ".mypy_cache", "testdata":
		return true
	}
	return false
}

// Walk finds packages under root. Results are sorted by directory, then ecosystem.
// Dir and Manifest in the results are relative to root.
func Walk(root string, ecosystems []ecosystem.Ecosystem) ([]ecosystem.Package, error) {
	var pkgs []ecosystem.Package
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && SkipDir(d.Name()) {
			return filepath.SkipDir
		}
		for _, eco := range ecosystems {
			p, err := eco.Detect(path)
			if err != nil {
				return err
			}
			if p == nil {
				continue
			}
			p.Dir = rel(root, p.Dir)
			p.Manifest = rel(root, p.Manifest)
			pkgs = append(pkgs, *p)
		}
		return nil
	})
	sort.SliceStable(pkgs, func(i, j int) bool {
		if pkgs[i].Dir != pkgs[j].Dir {
			return pkgs[i].Dir < pkgs[j].Dir
		}
		return pkgs[i].Ecosystem < pkgs[j].Ecosystem
	})
	return pkgs, err
}

func rel(root, path string) string {
	r, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(r)
}
