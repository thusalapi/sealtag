package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/ecosystem"
)

func newDetectCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "List every publishable package in the repository",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pkgs, err := detect.Walk(opts.dir, detect.All())
			if err != nil {
				return err
			}
			if opts.json {
				if pkgs == nil {
					pkgs = []ecosystem.Package{}
				}
				return writeJSON(cmd.OutOrStdout(), pkgs)
			}
			if len(pkgs) == 0 {
				cmd.Println("No packages found.")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ECOSYSTEM\tNAME\tVERSION\tPATH\tNOTE")
			for _, p := range pkgs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", p.Ecosystem, p.Name, orDash(p.Version), p.Dir, note(p))
			}
			return w.Flush()
		},
	}
}

func note(p ecosystem.Package) string {
	switch {
	case p.Private && p.Note != "":
		return "private, " + p.Note
	case p.Private:
		return "private (will not publish)"
	}
	return p.Note
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
