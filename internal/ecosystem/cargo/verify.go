package cargo

import (
	"context"
	"io"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Verify runs `cargo package`, which unpacks the .crate into target/package and
// compiles it there, so only files inside the crate are available to the build.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, log io.Writer) error {
	return ecosystem.Run(ctx, log, dir, nil, "cargo", "package", "--allow-dirty", "--quiet")
}
