package cli

import (
	"errors"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/fileset"
	"github.com/thusalapi/sealtag/internal/scan"
)

// ErrFindings makes the process exit non-zero so CI fails when secrets are found.
var ErrFindings = errors.New("possible secrets found")

// ScanResult is the --json output of the scan command.
type ScanResult struct {
	FilesScanned int            `json:"files_scanned"`
	Findings     []scan.Finding `json:"findings"`
	Allowed      []scan.Finding `json:"allowed"`
}

func newScanCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Look for secrets in files that could be published",
		Long: "scan checks file names and contents for keys, tokens and credential files.\n" +
			"Inside a git repository it respects .gitignore. Known false positives can be\n" +
			"listed in " + scan.AllowFile + " with a reason.\n" +
			"It exits with status 1 when anything is found, so it can gate a CI job.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			files, err := fileset.List(opts.dir, detect.SkipDir)
			if err != nil {
				return err
			}
			all, err := scan.Files(opts.dir, files)
			if err != nil {
				return err
			}
			allow, err := scan.LoadAllow(opts.dir)
			if err != nil {
				return err
			}
			kept, allowed := scan.Filter(all, allow)
			res := ScanResult{FilesScanned: len(files), Findings: nonNil(kept), Allowed: nonNil(allowed)}

			if opts.json {
				if err := writeJSON(cmd.OutOrStdout(), res); err != nil {
					return err
				}
			} else {
				printScan(cmd, res)
			}
			if len(kept) > 0 {
				return ErrFindings
			}
			return nil
		},
	}
}

func printScan(cmd *cobra.Command, res ScanResult) {
	if len(res.Findings) == 0 {
		cmd.Printf("No secrets found in %d file(s).", res.FilesScanned)
		if n := len(res.Allowed); n > 0 {
			cmd.Printf(" %d finding(s) allowed by %s.", n, scan.AllowFile)
		}
		cmd.Println()
		return
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\tLINE\tRULE\tMATCH")
	for _, f := range res.Findings {
		line := "-"
		if f.Line > 0 {
			line = fmt.Sprint(f.Line)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.File, line, f.RuleID, orDash(f.Match))
	}
	w.Flush()
	cmd.Printf("\n%d finding(s) in %d file(s) scanned. Remove them, or add them to %s with a reason.\n",
		len(res.Findings), res.FilesScanned, scan.AllowFile)
}

func nonNil(f []scan.Finding) []scan.Finding {
	if f == nil {
		return []scan.Finding{}
	}
	return f
}
