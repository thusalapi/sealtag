package npm

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/semver"
)

// minNpm is the first npm release that performs the trusted publishing exchange
// itself during `npm publish`.
const minNpm = "11.5.1"

// Publish runs `npm publish`. npm itself requests the OIDC token and exchanges it,
// and it attaches provenance automatically when publishing through trusted
// publishing from a public repository.
func (Ecosystem) Publish(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.PublishOptions) error {
	if err := checkNpmVersion(); err != nil {
		if !opts.DryRun {
			return err
		}
		fmt.Fprintf(opts.Log, "warning: %v\n", err)
	}
	args := []string{"publish", "--access", "public"}
	if opts.DryRun {
		args = append(args, "--dry-run")
	} else if !opts.OIDC.InGitHubActions() {
		return fmt.Errorf("npm trusted publishing needs GitHub Actions with `id-token: write`")
	}
	return ecosystem.Run(ctx, opts.Log, dir, nil, "npm", args...)
}

func checkNpmVersion() error {
	out, err := exec.Command("npm", "--version").Output()
	if err != nil {
		return fmt.Errorf("npm: %w", ecosystem.ErrToolMissing)
	}
	have, err := semver.Parse(strings.TrimSpace(string(out)))
	if err != nil {
		return fmt.Errorf("reading npm version: %w", err)
	}
	want, _ := semver.Parse(minNpm)
	if have.Compare(want) < 0 {
		return fmt.Errorf("npm %s is too old for trusted publishing: need %s or newer (npm install -g npm@latest)", have, minNpm)
	}
	return nil
}
