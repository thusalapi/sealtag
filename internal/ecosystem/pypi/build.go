package pypi

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// buildTimeout stops a build backend that hangs, for example while downloading.
const buildTimeout = 5 * time.Minute

// maxRead caps how much of each archive member is kept in memory for scanning.
const maxRead = 2 << 20

// ShippedFiles builds the sdist and the wheel into a temporary folder and lists
// what is inside each archive. Python has no "dry run" that reports the file list:
// what ships is decided by the build backend (hatchling, setuptools, flit, ...),
// each with its own include rules. Building is the only accurate answer.
//
// Building runs the project's build backend, which is code. That is acceptable for
// a project you are about to publish, but see docs/decisions/0005.
func (Ecosystem) ShippedFiles(dir string) ([]ecosystem.File, error) {
	out, err := os.MkdirTemp("", "sealtag-pypi-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(out)

	if err := build(dir, out); err != nil {
		return nil, err
	}
	artifacts, err := filepath.Glob(filepath.Join(out, "*"))
	if err != nil {
		return nil, err
	}
	var files []ecosystem.File
	for _, a := range artifacts {
		var fs []ecosystem.File
		switch {
		case strings.HasSuffix(a, ".tar.gz"):
			fs, err = readSdist(a)
		case strings.HasSuffix(a, ".whl"):
			fs, err = readWheel(a)
		default:
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", filepath.Base(a), err)
		}
		files = append(files, fs...)
	}
	if len(files) == 0 {
		return nil, errors.New("the build produced no sdist or wheel")
	}
	sort.SliceStable(files, func(i, j int) bool {
		if files[i].Artifact != files[j].Artifact {
			return files[i].Artifact < files[j].Artifact
		}
		return files[i].Path < files[j].Path
	})
	return files, nil
}

// build prefers uv (fast, isolated) and falls back to the standard "build" module.
// Like PyPI, both build the wheel from the sdist, not from the source tree.
func build(dir, out string) error {
	ctx, cancel := context.WithTimeout(context.Background(), buildTimeout)
	defer cancel()
	var cmd *exec.Cmd
	if uv, err := exec.LookPath("uv"); err == nil {
		cmd = exec.CommandContext(ctx, uv, "build", "--out-dir", out, "--quiet", ".")
	} else if py := python(); py != "" {
		cmd = exec.CommandContext(ctx, py, "-m", "build", "--outdir", out, ".")
	} else {
		return fmt.Errorf("uv or python with the build module: %w", ecosystem.ErrToolMissing)
	}
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("build timed out after %s", buildTimeout)
		}
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 2000 {
			msg = msg[len(msg)-2000:]
		}
		return fmt.Errorf("build failed in %s: %s", dir, msg)
	}
	return nil
}

// python returns a Python interpreter that has the "build" module, or "".
func python() string {
	for _, name := range []string{"python3", "python"} {
		p, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if exec.Command(p, "-c", "import build").Run() == nil {
			return p
		}
	}
	return ""
}

// readSdist lists a .tar.gz sdist. Every member sits under "<name>-<version>/",
// which is stripped so paths match the source tree.
func readSdist(path string) ([]ecosystem.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	name := filepath.Base(path)
	var files []ecosystem.File
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		_, rel, ok := strings.Cut(h.Name, "/")
		if !ok {
			rel = h.Name
		}
		data, err := io.ReadAll(io.LimitReader(tr, maxRead))
		if err != nil {
			return nil, err
		}
		files = append(files, ecosystem.File{Path: rel, Size: h.Size, Artifact: name, Data: data})
	}
	return files, nil
}

// readWheel lists a .whl, which is a zip file.
func readWheel(path string) ([]ecosystem.File, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	name := filepath.Base(path)
	var files []ecosystem.File
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(io.LimitReader(rc, maxRead))
		rc.Close()
		if err != nil {
			return nil, err
		}
		files = append(files, ecosystem.File{Path: zf.Name, Size: int64(zf.UncompressedSize64), Artifact: name, Data: data})
	}
	return files, nil
}
