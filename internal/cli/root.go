// Package cli wires the shipwright commands together. Keep logic out of this package:
// commands should only parse flags, call internal packages and print results.
package cli

import (
	"github.com/spf13/cobra"
)

// Set at build time with -ldflags "-X github.com/thusalapi/shipwright/internal/cli.version=..."
var (
	version = "dev"
	commit  = "none"
)

type options struct {
	dir string
}

// NewRoot builds the root command with every subcommand attached.
func NewRoot() *cobra.Command {
	opts := &options{}
	root := &cobra.Command{
		Use:           "shipwright",
		Short:         "Publish packages to npm, PyPI, crates.io and Go safely",
		Long:          "shipwright detects publishable packages, plans releases and checks for leaked secrets\nbefore anything is published. Publishing with trusted publishing is coming soon.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVarP(&opts.dir, "dir", "C", ".", "repository root to work in")
	root.AddCommand(newDetectCmd(opts), newPlanCmd(opts), newScanCmd(opts), newVersionCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the shipwright version",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Printf("shipwright %s (%s)\n", version, commit)
		},
	}
}
