package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/semver"
)

var registries = map[string]string{
	"npm":   "registry.npmjs.org",
	"pypi":  "pypi.org",
	"cargo": "crates.io",
	"go":    "git tag (proxy.golang.org)",
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
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCURRENT\tNEXT\tTARGET")
			publishable, skipped := 0, 0
			for _, p := range pkgs {
				if p.Private {
					skipped++
					continue
				}
				next := "?"
				cur := p.Version
				if cur == "" && p.Ecosystem == "go" {
					cur = "v0.0.0" // first tag for a Go module
				}
				if v, err := semver.Parse(cur); err == nil {
					n, err := v.Bump(bump)
					if err != nil {
						return err
					}
					next = n.String()
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, orDash(p.Version), next, registries[p.Ecosystem])
				publishable++
			}
			if err := w.Flush(); err != nil {
				return err
			}
			cmd.Printf("\n%d package(s) would be released with a %s bump. %d private package(s) skipped.\n", publishable, bump, skipped)
			cmd.Println("This is a dry run. Nothing was changed.")
			return nil
		},
	}
	cmd.Flags().StringVar(&bump, "bump", semver.Patch, "version bump: major, minor or patch")
	return cmd
}
