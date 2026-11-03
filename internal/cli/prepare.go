package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/gitlog"
	"github.com/thusalapi/sealtag/internal/prepare"
	"github.com/thusalapi/sealtag/internal/release"
)

// ErrDirty stops --commit from mixing a release with unrelated work.
var ErrDirty = errors.New("the working tree has uncommitted changes: commit or stash them first, or run without --commit")

func newPrepareCmd(opts *options) *cobra.Command {
	var bump, date string
	var commit bool
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Write the next versions into manifests and update CHANGELOG.md",
		Long: "prepare applies the plan shown by `sealtag plan`: it writes each new version\n" +
			"into the package manifest and adds the release notes to the package's\n" +
			"CHANGELOG.md. It never publishes.\n\n" +
			"With --commit it also commits exactly those files as \"chore(release): ...\"\n" +
			"and creates one annotated tag per package. Push the tags to trigger publishing.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if commit {
				dirty, err := gitlog.Dirty(opts.dir)
				if err != nil {
					return err
				}
				if dirty {
					return ErrDirty
				}
			}
			plan, err := buildPlan(opts, bump)
			if err != nil {
				return err
			}
			if date == "" {
				date = time.Now().Format("2006-01-02")
			}
			res, err := prepare.Apply(opts.dir, plan, date)
			if err != nil {
				return err
			}
			if len(res.Releases) > 0 && commit {
				if err := gitlog.CommitFiles(opts.dir, res.Changed, releaseMessage(res)); err != nil {
					return fmt.Errorf("committing: %w", err)
				}
				for _, r := range res.Releases {
					if err := gitlog.Tag(opts.dir, r.NextTag, r.Package.Name+" "+r.Next); err != nil {
						return fmt.Errorf("tagging %s: %w", r.NextTag, err)
					}
				}
			}
			if opts.json {
				return writeJSON(cmd.OutOrStdout(), res)
			}
			printPrepare(cmd, res, commit)
			return nil
		},
	}
	cmd.Flags().StringVar(&bump, "bump", release.Auto, "auto (from commits), major, minor or patch")
	cmd.Flags().StringVar(&date, "date", "", "release date for the changelog (default today, YYYY-MM-DD)")
	cmd.Flags().BoolVar(&commit, "commit", false, "commit the changes and create the release tags")
	return cmd
}

func releaseMessage(res prepare.Result) string {
	var names []string
	for _, r := range res.Releases {
		names = append(names, r.Package.Name+" "+r.Next)
	}
	return "chore(release): " + strings.Join(names, ", ")
}

func printPrepare(cmd *cobra.Command, res prepare.Result, committed bool) {
	if len(res.Releases) == 0 {
		cmd.Println("Nothing to release. Run `sealtag plan` to see why.")
		return
	}
	for _, r := range res.Releases {
		cmd.Printf("%s: %s -> %s\n", r.Package.Name, orDash(r.Current), r.Next)
	}
	cmd.Println("\nChanged files:")
	for _, f := range res.Changed {
		cmd.Println("  " + f)
	}
	if committed {
		cmd.Println("\nCommitted and tagged:")
		for _, t := range res.Tags {
			cmd.Println("  " + t)
		}
		cmd.Println("\nPush with: git push --follow-tags")
		return
	}
	cmd.Println("\nReview the changes, then run `sealtag prepare --commit` on a clean tree,")
	cmd.Println("or commit them yourself and create these tags:")
	for _, t := range res.Tags {
		cmd.Println("  " + t)
	}
}
