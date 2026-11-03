package cli

import (
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/scan"
)

// ErrFindings makes the process exit non-zero so CI fails when secrets are found.
var ErrFindings = errors.New("possible secrets found")

func newScanCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Look for secrets in files that could be published",
		Long:  "scan checks file names and contents for keys, tokens and credential files.\nIt exits with status 1 when anything is found, so it can gate a CI job.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			findings, err := scan.Dir(opts.dir, detect.SkipDir)
			if err != nil {
				return err
			}
			if len(findings) == 0 {
				cmd.Println("No secrets found.")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "FILE\tLINE\tRULE\tMATCH")
			for _, f := range findings {
				line := "-"
				if f.Line > 0 {
					line = fmt.Sprint(f.Line)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.File, line, f.Description, orDash(f.Match))
			}
			w.Flush()
			cmd.Printf("\n%d finding(s). Remove these before publishing.\n", len(findings))
			return ErrFindings
		},
	}
}
