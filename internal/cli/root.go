// Package cli wires the sealtag commands together. Keep logic out of this package:
// commands should only parse flags, call internal packages and print results.
package cli

import (
	"github.com/spf13/cobra"
)

// Set at build time with -ldflags "-X github.com/thusalapi/sealtag/internal/cli.version=..."
var (
	version = "dev"
	commit  = "none"
)

type options struct {
	dir  string
	json bool
}

// NewRoot builds the root command with every subcommand attached.
func NewRoot() *cobra.Command {
	opts := &options{}
	root := &cobra.Command{
		Use:           "sealtag",
		Short:         "Publish packages to npm, PyPI, crates.io and Go safely",
		Long:          "sealtag detects publishable packages, plans releases and checks for leaked secrets\nbefore anything is published. Publishing with trusted publishing is coming soon.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVarP(&opts.dir, "dir", "C", ".", "repository root to work in")
	root.PersistentFlags().BoolVar(&opts.json, "json", false, "print machine-readable JSON")
	root.AddCommand(newDetectCmd(opts), newPlanCmd(opts), newScanCmd(opts), newCheckCmd(opts), newChangelogCmd(opts), newPrepareCmd(opts), newPublishCmd(opts), newDoctorCmd(opts), newVersionCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the sealtag version",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Printf("sealtag %s (%s)\n", version, commit)
		},
	}
}
