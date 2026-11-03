package ecosystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Run executes a registry tool in dir, streaming its output to log. extraEnv adds
// variables such as a short-lived token. Secrets go in the environment rather than
// the command line, because command lines are visible to other processes.
func Run(ctx context.Context, log io.Writer, dir string, extraEnv []string, name string, args ...string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s: %w", name, ErrToolMissing)
	}
	fmt.Fprintf(log, "$ %s %s\n", name, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	cmd.Stdout = log
	cmd.Stderr = log
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w", name, args[0], err)
	}
	return nil
}
