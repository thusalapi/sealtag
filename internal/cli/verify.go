package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// ErrVerifyFailed is returned when any package fails to install or load.
var ErrVerifyFailed = errors.New("verify failed")

// VerifyResult is the result for one package.
type VerifyResult struct {
	Package ecosystem.Package `json:"package"`
	Status  string            `json:"status"` // "ok", "failed" or "unsupported"
	Error   string            `json:"error,omitempty"`
}

func verifyPackage(root string, p ecosystem.Package, log io.Writer) VerifyResult {
	v, ok := detect.ByName(p.Ecosystem).(ecosystem.Verifier)
	if !ok {
		return VerifyResult{Package: p, Status: "unsupported"}
	}
	dir := filepath.Join(root, filepath.FromSlash(p.Dir))
	if err := v.Verify(context.Background(), dir, p, log); err != nil {
		return VerifyResult{Package: p, Status: "failed", Error: err.Error()}
	}
	return VerifyResult{Package: p, Status: "ok"}
}

func newVerifyCmd(opts *options) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Install each package from its built artifact in an empty environment and load it",
		Long: "verify builds each package the way the registry will receive it, installs it\n" +
			"into a fresh temporary environment with nothing from the repository, and checks\n" +
			"that it loads: npm import, Python import in a new venv, cargo package build,\n" +
			"and go build of the extracted module zip.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			pkgs, err := detect.Walk(opts.dir, detect.All())
			if err != nil {
				return err
			}
			log := io.Discard
			if verbose && !opts.json {
				log = cmd.OutOrStdout()
			}
			results := []VerifyResult{}
			failed := false
			for _, p := range pkgs {
				if p.Private {
					continue
				}
				r := verifyPackage(opts.dir, p, log)
				failed = failed || r.Status == "failed"
				results = append(results, r)
				if !opts.json {
					mark := map[string]string{"ok": "PASS", "failed": "FAIL", "unsupported": "SKIP"}[r.Status]
					cmd.Printf("[%s] %s (%s)\n", mark, p.Name, p.Ecosystem)
					if r.Error != "" {
						cmd.Printf("       %s\n", r.Error)
					}
				}
			}
			if opts.json {
				if err := writeJSON(cmd.OutOrStdout(), results); err != nil {
					return err
				}
			}
			if failed {
				if !verbose && !opts.json {
					fmt.Fprintln(cmd.OutOrStdout(), "\nRun with --verbose to see the tool output.")
				}
				return ErrVerifyFailed
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show the output of npm, uv, cargo and go")
	return cmd
}
