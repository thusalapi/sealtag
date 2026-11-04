package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/gitlog"
	"github.com/thusalapi/sealtag/internal/oidc"
	"github.com/thusalapi/sealtag/internal/release"
	"github.com/thusalapi/sealtag/internal/scan"
	"github.com/thusalapi/sealtag/internal/semver"
)

// longLivedTokenVars are environment variables that usually hold a long-lived
// registry token. Their presence means the workflow is not using trusted publishing.
var longLivedTokenVars = []string{"NPM_TOKEN", "NODE_AUTH_TOKEN", "CARGO_REGISTRY_TOKEN", "TWINE_PASSWORD", "UV_PUBLISH_TOKEN", "PYPI_TOKEN", "PYPI_API_TOKEN"}

// ErrLongLivedToken is returned when a stored token is found and not explicitly allowed.
var ErrLongLivedToken = errors.New("long-lived registry token found in the environment")

// Target is one package release found at HEAD.
type Target struct {
	Package ecosystem.Package `json:"package"`
	Tag     string            `json:"tag"`
	Version string            `json:"version"` // as written in the changelog heading
}

func newPublishCmd(opts *options) *cobra.Command {
	var dryRun, allowToken, skipVerify bool
	var onlyTag string
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish the packages whose release tags point at HEAD, using trusted publishing",
		Long: "publish is meant to run in CI after `git push --follow-tags`. It finds the release\n" +
			"tags on the current commit, checks that each manifest matches its tag, runs\n" +
			"`sealtag check` and `sealtag verify` on those packages, and uploads them using short-lived OIDC tokens.\n\n" +
			"It refuses to run when a long-lived registry token is set in the environment,\n" +
			"unless --allow-token is given.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if found := tokenVars(os.Getenv); len(found) > 0 && !allowToken {
				return fmt.Errorf("%w: %s. Remove it and configure trusted publishing (see `sealtag doctor`), or pass --allow-token",
					ErrLongLivedToken, strings.Join(found, ", "))
			}
			targets, err := taggedTargets(opts.dir)
			if err != nil {
				return err
			}
			if onlyTag != "" {
				targets = filterTag(targets, onlyTag)
			}
			log := cmd.OutOrStdout()
			if len(targets) == 0 {
				fmt.Fprintln(log, "No release tags point at HEAD. Nothing to publish.")
				return nil
			}
			allow, err := scan.LoadAllow(opts.dir)
			if err != nil {
				return err
			}
			// Fail closed: check every package before uploading any of them, so a
			// problem in one does not leave the release half published.
			for _, t := range targets {
				r := checkPackage(opts.dir, t.Package, allow)
				if r.Status != "ok" {
					printCheck(cmd, []PackageCheck{r}, false)
					return fmt.Errorf("%s: %w", t.Package.Name, ErrCheckFailed)
				}
			}
			if !skipVerify {
				for _, t := range targets {
					fmt.Fprintf(log, "verifying %s in a clean environment\n", t.Package.Name)
					if r := verifyPackage(opts.dir, t.Package, log); r.Status == "failed" {
						return fmt.Errorf("%s: %w: %s", t.Package.Name, ErrVerifyFailed, r.Error)
					}
				}
			}
			popts := ecosystem.PublishOptions{DryRun: dryRun, Log: log, OIDC: oidc.Client{}}
			for _, t := range targets {
				mode := ""
				if dryRun {
					mode = " (dry run)"
				}
				fmt.Fprintf(log, "\n==> %s %s to %s%s\n", t.Package.Name, t.Tag, t.Package.Ecosystem, mode)
				pub, ok := detect.ByName(t.Package.Ecosystem).(ecosystem.Publisher)
				if !ok {
					return fmt.Errorf("%s: publishing is not supported for %s yet", t.Package.Name, t.Package.Ecosystem)
				}
				dir := filepath.Join(opts.dir, filepath.FromSlash(t.Package.Dir))
				if err := pub.Publish(context.Background(), dir, t.Package, popts); err != nil {
					return fmt.Errorf("%s: %w", t.Package.Name, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "check and build everything, but do not upload")
	cmd.Flags().BoolVar(&allowToken, "allow-token", false, "allow a long-lived registry token in the environment")
	cmd.Flags().StringVar(&onlyTag, "tag", "", "publish only the package for this tag (in CI: the tag that triggered the run)")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "do not install and load packages in a clean environment first")
	return cmd
}

func tokenVars(getenv func(string) string) []string {
	var found []string
	for _, k := range longLivedTokenVars {
		if getenv(k) != "" {
			found = append(found, k)
		}
	}
	return found
}

// taggedTargets matches release tags on HEAD to packages, and checks that each
// manifest agrees with its tag. A mismatch usually means the tag was created by
// hand without running `sealtag prepare`.
func taggedTargets(root string) ([]Target, error) {
	tags, err := gitlog.TagsAtHead(root)
	if err != nil {
		return nil, err
	}
	return matchTags(root, tags, true)
}

// taggedTargetsAll matches every tag in the repository, without the manifest check
// (an old tag's version is naturally different from today's manifest).
func taggedTargetsAll(root string) ([]Target, error) {
	tags, err := gitlog.AllTags(root)
	if err != nil {
		return nil, err
	}
	return matchTags(root, tags, false)
}

func matchTags(root string, tags []string, checkManifest bool) ([]Target, error) {
	pkgs, err := detect.Walk(root, detect.All())
	if err != nil {
		return nil, err
	}
	var public []ecosystem.Package
	for _, p := range pkgs {
		if !p.Private {
			public = append(public, p)
		}
	}
	var targets []Target
	for _, p := range public {
		prefix := release.TagPrefix(p, len(public) > 1)
		for _, tag := range tags {
			rest, ok := strings.CutPrefix(tag, prefix)
			if !ok {
				continue
			}
			tv, err := semver.Parse(rest)
			if err != nil {
				continue
			}
			t := Target{Package: p, Tag: tag, Version: tv.String()}
			if p.Ecosystem == "go" {
				t.Version = "v" + tv.String()
				t.Package.Version = t.Version
			} else if checkManifest {
				if mv, err := semver.Parse(p.Version); err != nil || mv.Compare(tv) != 0 {
					return nil, fmt.Errorf("%s: tag %s does not match the manifest version %q: run `sealtag prepare`", p.Name, tag, p.Version)
				}
			}
			targets = append(targets, t)
		}
	}
	return targets, nil
}

// filterTag keeps only the target for one tag. Without it, a run triggered by one
// tag would also publish every other release tag on the same commit, including a
// floating tag such as v0 that points at the latest v0.x release.
func filterTag(targets []Target, tag string) []Target {
	var out []Target
	for _, t := range targets {
		if t.Tag == tag {
			out = append(out, t)
		}
	}
	return out
}
