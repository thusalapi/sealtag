package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/ecosystem"
	"github.com/thusalapi/sealtag/internal/semver"
)

var registries = map[string]string{
	"npm":   "registry.npmjs.org",
	"pypi":  "pypi.org",
	"cargo": "crates.io",
	"go":    "git tag (proxy.golang.org)",
}

// Release is one planned package release.
type Release struct {
	Package ecosystem.Package `json:"package"`
	Current string            `json:"current"`
	Next    string            `json:"next,omitempty"` // empty when the version is unknown
	Target  string            `json:"target"`
}

// Plan is the full dry-run result.
type Plan struct {
	Bump     string              `json:"bump"`
	Releases []Release           `json:"releases"`
	Skipped  []ecosystem.Package `json:"skipped"`
}

func buildPlan(pkgs []ecosystem.Package, bump string) (Plan, error) {
	plan := Plan{Bump: bump, Releases: []Release{}, Skipped: []ecosystem.Package{}}
	for _, p := range pkgs {
		if p.Private {
			plan.Skipped = append(plan.Skipped, p)
			continue
		}
		r := Release{Package: p, Current: p.Version, Target: registries[p.Ecosystem]}
		cur := p.Version
		if cur == "" && p.Ecosystem == "go" {
			cur = "v0.0.0" // a Go module with no tags starts from v0.0.0
		}
		if v, err := semver.Parse(cur); err == nil {
			n, err := v.Bump(bump)
			if err != nil {
				return Plan{}, err
			}
			r.Next = n.String()
		}
		plan.Releases = append(plan.Releases, r)
	}
	return plan, nil
}

func newPlanCmd(opts *options) *cobra.Command {
	var bump string
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show what a release would publish, without changing anything",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pkgs, err := detect.Walk(opts.dir, detect.All())
			if err != nil {
				return err
			}
			plan, err := buildPlan(pkgs, bump)
			if err != nil {
				return err
			}
			if opts.json {
				return writeJSON(cmd.OutOrStdout(), plan)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCURRENT\tNEXT\tTARGET")
			for _, r := range plan.Releases {
				next := r.Next
				if next == "" {
					next = "?"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Package.Name, orDash(r.Current), next, r.Target)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			cmd.Printf("\n%d package(s) would be released with a %s bump. %d private package(s) skipped.\n",
				len(plan.Releases), bump, len(plan.Skipped))
			cmd.Println("This is a dry run. Nothing was changed.")
			return nil
		},
	}
	cmd.Flags().StringVar(&bump, "bump", semver.Patch, "version bump: major, minor or patch")
	return cmd
}
