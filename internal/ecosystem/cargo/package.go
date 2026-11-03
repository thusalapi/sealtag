package cargo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// ShippedFiles asks cargo which files `cargo publish` would include. cargo applies the
// "include" and "exclude" manifest keys and .gitignore rules.
// --allow-dirty lets this run with uncommitted changes, since nothing is uploaded.
func (Ecosystem) ShippedFiles(dir string) ([]ecosystem.File, error) {
	cargo, err := exec.LookPath("cargo")
	if err != nil {
		return nil, fmt.Errorf("cargo: %w", ecosystem.ErrToolMissing)
	}
	cmd := exec.Command(cargo, "package", "--list", "--allow-dirty", "--quiet")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("cargo package failed in %s: %s", dir, bytes.TrimSpace(stderr.Bytes()))
		}
		return nil, err
	}
	return parseList(dir, out)
}

func parseList(dir string, out []byte) ([]ecosystem.File, error) {
	var files []ecosystem.File
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		p := strings.TrimSpace(sc.Text())
		if p == "" {
			continue
		}
		p = filepath.ToSlash(p)
		f := ecosystem.File{Path: p}
		// Cargo adds files that are not in the crate folder: Cargo.toml.orig,
		// .cargo_vcs_info.json, and Cargo.lock copied from the workspace root.
		info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(p)))
		switch {
		case err == nil:
			f.Size = info.Size()
		case errors.Is(err, os.ErrNotExist):
			f.Generated = true
		default:
			return nil, err
		}
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, sc.Err()
}
