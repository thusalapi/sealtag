package cli

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/release"
)

func newChangelogCmd(opts *options) *cobra.Command {
	var bump, date string
	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Print the changelog for the pending release of each package",
		Long: "changelog prints Markdown release notes built from Conventional Commits\n" +
			"since each package's last release. It does not write any file yet.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			plan, err := buildPlan(opts, bump)
			if err != nil {
				return err
			}
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}
			notes := map[string]string{}
			for _, r := range plan.Releases {
				if md := release.Changelog(r, date); md != "" {
					notes[r.Package.Name] = md
				}
			}
			if opts.json {
				return writeJSON(cmd.OutOrStdout(), notes)
			}
			if len(notes) == 0 {
				cmd.Println("Nothing to release.")
				return nil
			}
			first := true
			for _, r := range plan.Releases {
				md, ok := notes[r.Package.Name]
				if !ok {
					continue
				}
				if !first {
					cmd.Println()
				}
				cmd.Print(md)
				first = false
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&bump, "bump", release.Auto, "auto (from commits), major, minor or patch")
	cmd.Flags().StringVar(&date, "date", "", "release date to print (default today, YYYY-MM-DD)")
	return cmd
}
