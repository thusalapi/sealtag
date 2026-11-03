package pypi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/oidc"
)

// Index is the PyPI instance to publish to. Tests and rehearsals use TestPyPI.
var Index = "https://pypi.org"

// Publish builds the sdist and wheel, exchanges a GitHub OIDC token for a PyPI
// token, and uploads both files with uv.
func (Ecosystem) Publish(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.PublishOptions) error {
	out, err := os.MkdirTemp("", "sealtag-publish-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(out)
	if err := build(dir, out); err != nil {
		return err
	}
	files, _ := filepath.Glob(filepath.Join(out, "*"))
	if opts.DryRun {
		for _, f := range files {
			fmt.Fprintf(opts.Log, "would upload %s to %s\n", filepath.Base(f), Index)
		}
		return nil
	}
	aud, err := opts.OIDC.PyPIAudience(ctx, Index)
	if err != nil {
		return err
	}
	jwt, err := opts.OIDC.GitHubToken(ctx, aud)
	if err != nil {
		return err
	}
	token, err := opts.OIDC.PyPIToken(ctx, Index, jwt)
	if err != nil {
		return err
	}
	oidc.Mask(opts.Log, token)
	args := append([]string{"publish", "--publish-url", UploadURL(Index)}, files...)
	return ecosystem.Run(ctx, opts.Log, dir, []string{"UV_PUBLISH_TOKEN=" + token}, "uv", args...)
}

// UploadURL returns the upload endpoint for an index. PyPI serves uploads from a
// separate host, upload.pypi.org.
func UploadURL(index string) string {
	if index == "https://pypi.org" {
		return "https://upload.pypi.org/legacy/"
	}
	return index + "/legacy/"
}
