package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/thusalapi/sealtag/internal/detect"
	"github.com/thusalapi/sealtag/internal/gitlog"
)

const workflowFile = "release.yml"

func newDoctorCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Explain how to set up trusted publishing for each package",
		Long: "doctor checks the repository setup and prints, for each package, the exact\n" +
			"values to enter on the registry's trusted publisher page.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ok := func(good bool, msg, fix string) {
				if good {
					cmd.Printf("  [ok]   %s\n", msg)
				} else {
					cmd.Printf("  [todo] %s\n         %s\n", msg, fix)
				}
			}
			repo := gitlog.GitHubRepo(opts.dir)
			wf := filepath.Join(opts.dir, ".github", "workflows", workflowFile)
			wfData, wfErr := os.ReadFile(wf)

			cmd.Println("Repository")
			ok(repo != "", "origin remote is on GitHub ("+orDash(repo)+")", "git remote add origin https://github.com/<owner>/<repo>.git")
			ok(wfErr == nil, ".github/workflows/"+workflowFile+" exists", "add a release workflow that runs `sealtag publish` on tag pushes")
			ok(wfErr == nil && strings.Contains(string(wfData), "id-token: write"), "the workflow has `permissions: id-token: write`", "without it GitHub will not issue OIDC tokens")
			found := tokenVars(os.Getenv)
			ok(len(found) == 0, "no long-lived registry tokens in this environment", "unset "+strings.Join(found, ", "))

			pkgs, err := detect.Walk(opts.dir, detect.All())
			if err != nil {
				return err
			}
			owner, name, _ := strings.Cut(orDash(repo), "/")
			for _, p := range pkgs {
				if p.Private {
					continue
				}
				cmd.Printf("\n%s (%s)\n", p.Name, p.Ecosystem)
				switch p.Ecosystem {
				case "go":
					cmd.Println("  Nothing to configure. Pushing the tag publishes a Go module.")
					continue
				case "npm":
					cmd.Printf("  Open https://www.npmjs.com/package/%s/access and add a Trusted Publisher.\n", p.Name)
					cmd.Println("  npm only allows this after the package exists, so the very first version")
					cmd.Println("  must be published by hand (npm publish). Every later version uses OIDC.")
				case "pypi":
					cmd.Println("  Open https://pypi.org/manage/account/publishing/ and add a (pending) publisher.")
					cmd.Println("  A pending publisher lets the first release come from CI too.")
				case "cargo":
					cmd.Printf("  Open https://crates.io/crates/%s/settings and add a Trusted Publisher.\n", p.Name)
					cmd.Println("  crates.io needs one manual `cargo publish` before this page exists.")
				}
				cmd.Printf("  Owner: %s\n  Repository: %s\n  Workflow: %s\n  Environment: release\n", owner, name, workflowFile)
			}
			return nil
		},
	}
}
