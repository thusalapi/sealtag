package pypi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Image is the container used by `verify --container`.
const Image = "python:3.13-slim"

// Verify builds the wheel on the host, installs it into a fresh virtual
// environment and imports every top-level module the wheel contains. With
// opts.Container the venv lives in a new container that only sees the wheel.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.VerifyOptions) error {
	tmp, err := os.MkdirTemp("", "sealtag-verify-pypi-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := build(dir, filepath.Join(tmp, "dist")); err != nil {
		return err
	}
	wheels, _ := filepath.Glob(filepath.Join(tmp, "dist", "*.whl"))
	if len(wheels) != 1 {
		return fmt.Errorf("build produced %d wheels, want 1", len(wheels))
	}
	files, err := readWheel(wheels[0])
	if err != nil {
		return err
	}
	modules := topLevelModules(files)
	if len(modules) == 0 {
		return fmt.Errorf("the wheel contains no importable module")
	}
	// -I isolates Python from the current folder, so the import must come from the
	// installed wheel and not from the source tree.
	script := "import importlib\nfor m in " + pyList(modules) + ":\n    importlib.import_module(m)\n    print('imported', m)\n"
	sb := ecosystem.Sandbox{Opts: opts, Image: Image, Root: tmp}
	wheel := sb.Path("dist/" + filepath.Base(wheels[0]))

	if opts.Container {
		// The python image has pip but not uv. One shell keeps the venv between steps.
		sh := "python -m venv /tmp/venv && /tmp/venv/bin/pip install --quiet " + shellQuote(wheel) +
			" && /tmp/venv/bin/python -I -c " + shellQuote(script)
		return sb.Run(ctx, ".", "sh", "-c", sh)
	}
	venv := filepath.Join(tmp, "venv")
	if err := sb.Run(ctx, ".", "uv", "venv", "--quiet", venv); err != nil {
		return err
	}
	py := filepath.Join(venv, "bin", "python")
	if runtime.GOOS == "windows" {
		py = filepath.Join(venv, "Scripts", "python.exe")
	}
	if err := sb.Run(ctx, ".", "uv", "pip", "install", "--quiet", "--python", py, wheel); err != nil {
		return err
	}
	return sb.Run(ctx, ".", py, "-I", "-c", script)
}

// shellQuote wraps s in single quotes for sh.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// topLevelModules returns the importable top-level names in a wheel.
func topLevelModules(files []ecosystem.File) []string {
	seen := map[string]bool{}
	for _, f := range files {
		first, rest, nested := strings.Cut(f.Path, "/")
		switch {
		case strings.HasSuffix(first, ".dist-info") || strings.HasSuffix(first, ".data"):
			continue
		case nested && rest == "__init__.py":
			seen[first] = true
		case !nested && strings.HasSuffix(first, ".py"):
			seen[strings.TrimSuffix(first, ".py")] = true
		}
	}
	var out []string
	for m := range seen {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

func pyList(items []string) string {
	q := make([]string, len(items))
	for i, s := range items {
		q[i] = "'" + s + "'"
	}
	return "[" + strings.Join(q, ", ") + "]"
}
