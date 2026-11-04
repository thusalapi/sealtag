package pypi

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Verify builds the wheel, installs it into a fresh virtual environment and imports
// every top-level module the wheel contains.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, log io.Writer) error {
	tmp, err := os.MkdirTemp("", "sealtag-verify-pypi-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	dist := filepath.Join(tmp, "dist")
	if err := build(dir, dist); err != nil {
		return err
	}
	wheels, _ := filepath.Glob(filepath.Join(dist, "*.whl"))
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
	venv := filepath.Join(tmp, "venv")
	if err := ecosystem.Run(ctx, log, tmp, nil, "uv", "venv", "--quiet", venv); err != nil {
		return err
	}
	py := filepath.Join(venv, "bin", "python")
	if runtime.GOOS == "windows" {
		py = filepath.Join(venv, "Scripts", "python.exe")
	}
	if err := ecosystem.Run(ctx, log, tmp, nil, "uv", "pip", "install", "--quiet", "--python", py, wheels[0]); err != nil {
		return err
	}
	// -I isolates Python from the current folder, so the import must come from the
	// installed wheel and not from the source tree.
	script := "import importlib\nfor m in " + pyList(modules) + ":\n    importlib.import_module(m)\n    print('imported', m)\n"
	return ecosystem.Run(ctx, log, tmp, nil, py, "-I", "-c", script)
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
