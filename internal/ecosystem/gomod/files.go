package gomod

import (
	"os"
	"path/filepath"
	"sort"

	modzip "golang.org/x/mod/zip"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/fileset"
)

// ShippedFiles applies the official module zip rules (golang.org/x/mod/zip), the same
// code the Go module proxy uses. The proxy builds the zip from the git commit, so when
// dir is inside a git repository only tracked, non-ignored files are kept.
func (Ecosystem) ShippedFiles(dir string) ([]ecosystem.File, error) {
	checked, err := modzip.CheckDir(dir)
	if err != nil {
		return nil, err
	}
	tracked, inGit := fileset.GitFiles(dir)
	keep := map[string]bool{}
	for _, f := range tracked {
		keep[f] = true
	}
	var files []ecosystem.File
	for _, p := range checked.Valid {
		// CheckDir returns paths joined with dir. Make them relative again.
		r, err := filepath.Rel(dir, p)
		if err != nil {
			return nil, err
		}
		rel := filepath.ToSlash(r)
		if inGit && !keep[rel] {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		files = append(files, ecosystem.File{Path: rel, Size: info.Size()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}
