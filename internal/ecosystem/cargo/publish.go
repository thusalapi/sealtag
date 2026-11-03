package cargo

import (
	"context"
	"fmt"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/oidc"
)

// Registry is the crates.io API base URL.
var Registry = "https://crates.io"

// Publish exchanges a GitHub OIDC token for a crates.io token, runs
// `cargo publish`, then revokes the token so it cannot be reused even within its
// short lifetime.
func (Ecosystem) Publish(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.PublishOptions) error {
	if opts.DryRun {
		return ecosystem.Run(ctx, opts.Log, dir, nil, "cargo", "publish", "--dry-run", "--allow-dirty")
	}
	jwt, err := opts.OIDC.GitHubToken(ctx, "crates.io")
	if err != nil {
		return err
	}
	token, err := opts.OIDC.CratesToken(ctx, Registry, jwt)
	if err != nil {
		return err
	}
	oidc.Mask(opts.Log, token)
	defer func() {
		if err := opts.OIDC.RevokeCratesToken(context.WithoutCancel(ctx), Registry, token); err != nil {
			fmt.Fprintf(opts.Log, "warning: could not revoke crates.io token: %v\n", err)
		}
	}()
	return ecosystem.Run(ctx, opts.Log, dir, []string{"CARGO_REGISTRY_TOKEN=" + token}, "cargo", "publish")
}
