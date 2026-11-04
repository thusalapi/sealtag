package ecosystem

import (
	"bytes"
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
//
// On failure the error includes the last lines of stderr, so the cause is visible
// even when log is io.Discard.
func Run(ctx context.Context, log io.Writer, dir string, extraEnv []string, name string, args ...string) error {
	path, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s: %w", name, ErrToolMissing)
	}
	fmt.Fprintf(log, "$ %s %s\n", name, strings.Join(args, " "))
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	var stderr bytes.Buffer
	cmd.Stdout = log
	cmd.Stderr = io.MultiWriter(log, &stderr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s failed: %w%s", name, args[0], err, lastLines(stderr.String(), 6))
	}
	return nil
}

func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return ""
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return "\n         " + strings.Join(lines, "\n         ")
}
