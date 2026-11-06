package cargo

import (
	"context"
	"path/filepath"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Image is the container used by `verify --container`.
const Image = "rust:1-slim"

// Verify runs `cargo package`, which unpacks the .crate into target/package and
// compiles it there, so only files inside the crate are available to the build.
// In a container, CARGO_HOME is empty, so no host registry cache is reused.
//
// A workspace member needs its workspace root (for inherited fields and the lock
// file), so the container mounts the workspace and works in the crate's folder.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.VerifyOptions) error {
	root, rel := dir, "."
	if ws := workspaceRoot(dir); ws != "" {
		if r, err := filepath.Rel(ws, dir); err == nil {
			root, rel = ws, filepath.ToSlash(r)
		}
	}
	sb := ecosystem.Sandbox{Opts: opts, Image: Image, Root: root}
	if opts.Container {
		sb.Env = []string{"CARGO_HOME=/tmp/cargo"}
	}
	return sb.Run(ctx, rel, "cargo", "package", "--allow-dirty", "--quiet")
}
