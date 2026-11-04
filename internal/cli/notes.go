package cli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/release"
)

func newNotesCmd(opts *options) *cobra.Command {
	var tag string
	cmd := &cobra.Command{
		Use:   "notes",
		Short: "Print the release notes for a released tag, from its CHANGELOG.md",
		Long: "notes finds the package that owns --tag and prints that release's section from\n" +
			"the package's CHANGELOG.md. Use it as the body of a GitHub release.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if tag == "" {
				return fmt.Errorf("--tag is required")
			}
			targets, err := taggedTargetsAll(opts.dir)
			if err != nil {
				return err
			}
			for _, t := range targets {
				if t.Tag != tag {
					continue
				}
				data, err := os.ReadFile(filepath.Join(opts.dir, filepath.FromSlash(path.Join(t.Package.Dir, "CHANGELOG.md"))))
				if err != nil {
					return err
				}
				version := t.Version
				notes, err := release.NotesFromChangelog(string(data), t.Package.Name, version)
				if err != nil {
					return fmt.Errorf("%s %s: %w", t.Package.Name, version, err)
				}
				cmd.Print(notes)
				return nil
			}
			return fmt.Errorf("no package release matches tag %q", tag)
		},
	}
	cmd.Flags().StringVar(&tag, "tag", "", "release tag, for example v1.2.0 or @acme/web@v2.5.0")
	return cmd
}
