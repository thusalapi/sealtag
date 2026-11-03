package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/release"
)

func buildPlan(opts *options, bump string) (release.Plan, error) {
	pkgs, err := detect.Walk(opts.dir, detect.All())
	if err != nil {
		return release.Plan{}, err
	}
	return release.Build(opts.dir, pkgs, bump)
}

func newPlanCmd(opts *options) *cobra.Command {
	var bump string
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show what a release would publish, without changing anything",
		Long: "plan works out the next version of every package.\n\n" +
			"With --bump auto (the default) it reads Conventional Commits since each\n" +
			"package's last release tag: fix and perf give a patch, feat gives a minor,\n" +
			"and a breaking change (feat! or BREAKING CHANGE:) gives a major.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			plan, err := buildPlan(opts, bump)
			if err != nil {
				return err
			}
			if opts.json {
				return writeJSON(cmd.OutOrStdout(), plan)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tCURRENT\tNEXT\tBUMP\tCOMMITS\tTAG\tTARGET")
			releasing := 0
			for _, r := range plan.Releases {
				next, tag := r.Next, r.NextTag
				if next == "" {
					next, tag = "-", "-"
				} else {
					releasing++
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
					r.Package.Name, orDash(r.Current), next, orDash(r.Bump), len(r.Commits), tag, r.Target)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			for _, r := range plan.Releases {
				if r.Reason != "" {
					cmd.Printf("  note: %s: %s\n", r.Package.Name, r.Reason)
				}
			}
			cmd.Printf("\n%d of %d package(s) would be released. %d private package(s) skipped.\n",
				releasing, len(plan.Releases), len(plan.Skipped))
			cmd.Println("This is a dry run. Nothing was changed.")
			return nil
		},
	}
	cmd.Flags().StringVar(&bump, "bump", release.Auto, "auto (from commits), major, minor or patch")
	return cmd
}
