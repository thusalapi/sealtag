package cargo

import (
	"context"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Image is the container used by `verify --container`.
const Image = "rust:1-slim"

// Verify runs `cargo package`, which unpacks the .crate into target/package and
// compiles it there, so only files inside the crate are available to the build.
// In a container, CARGO_HOME is empty, so no host registry cache is reused.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.VerifyOptions) error {
	sb := ecosystem.Sandbox{Opts: opts, Image: Image, Root: dir}
	if opts.Container {
		sb.Env = []string{"CARGO_HOME=/tmp/cargo"}
	}
	return sb.Run(ctx, ".", "cargo", "package", "--allow-dirty", "--quiet")
}
