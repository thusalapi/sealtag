package gomod

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/mod/module"
	modzip "golang.org/x/mod/zip"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Verify builds the module zip exactly as the proxy would, extracts it into an empty
// folder and runs `go build ./...` there. A module that only builds because of an
// untracked or excluded file fails here instead of failing for users.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, log io.Writer) error {
	tmp, err := os.MkdirTemp("", "sealtag-verify-go-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	files, err := (Ecosystem{}).ShippedFiles(dir)
	if err != nil {
		return err
	}
	var zfiles []modzip.File
	for _, f := range files {
		zfiles = append(zfiles, diskFile{root: dir, path: f.Path})
	}
	zipPath := filepath.Join(tmp, "module.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	// The version only labels the zip contents. v0.0.0 is always valid.
	mv := module.Version{Path: p.Name, Version: "v0.0.0"}
	if err := modzip.Create(zf, mv, zfiles); err != nil {
		zf.Close()
		return err
	}
	zf.Close()
	out := filepath.Join(tmp, "src")
	if err := modzip.Unzip(out, mv, zipPath); err != nil {
		return err
	}
	// Unzip makes files read-only, like the module cache. `go build` may need to
	// add missing requirements to go.mod, and the temp folder must be removable.
	if err := makeWritable(out); err != nil {
		return err
	}
	return ecosystem.Run(ctx, log, out, []string{"GOFLAGS=-mod=mod"}, "go", "build", "./...")
}

// diskFile adapts a file on disk to the modzip.File interface.
type diskFile struct{ root, path string }

func (f diskFile) Path() string { return f.path }
func (f diskFile) Lstat() (os.FileInfo, error) {
	return os.Lstat(filepath.Join(f.root, filepath.FromSlash(f.path)))
}
func (f diskFile) Open() (io.ReadCloser, error) {
	return os.Open(filepath.Join(f.root, filepath.FromSlash(f.path)))
}

func makeWritable(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if d.IsDir() {
			mode = 0o755
		}
		return os.Chmod(path, mode)
	})
}
