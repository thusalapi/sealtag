package ecosystem

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Sandbox runs verify steps for one folder, either on the host or inside a
// container that sees only that folder, mounted at /work.
type Sandbox struct {
	Opts  VerifyOptions
	Image string // container image, used when Opts.Container is true
	Root  string // host folder that holds everything the steps need
	Env   []string
}

// Engine returns the container engine to use: docker, or podman when docker is
// not installed.
func Engine() (string, error) {
	for _, e := range []string{"docker", "podman"} {
		if p, err := exec.LookPath(e); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("docker or podman: %w", ErrToolMissing)
}

// Path returns how a file under Root is named inside the sandbox.
func (s Sandbox) Path(rel string) string {
	if s.Opts.Container {
		return "/work/" + filepath.ToSlash(rel)
	}
	return filepath.Join(s.Root, filepath.FromSlash(rel))
}

// Run executes a command with rel (a folder under Root) as the working directory.
func (s Sandbox) Run(ctx context.Context, rel, name string, args ...string) error {
	if !s.Opts.Container {
		return Run(ctx, s.Opts.Log, filepath.Join(s.Root, filepath.FromSlash(rel)), s.Env, name, args...)
	}
	engine, err := Engine()
	if err != nil {
		return err
	}
	cargs := s.ContainerArgs(rel, name, args...)
	return Run(ctx, s.Opts.Log, s.Root, nil, filepath.Base(engine), cargs...)
}

// ContainerArgs returns the arguments Run passes to the container engine.
// HOME points at the container's own /tmp, so npm, pip, go and cargo start with
// empty caches. Commands run as the current user so files written into the
// mounted folder are not owned by root on Linux.
func (s Sandbox) ContainerArgs(rel, name string, args ...string) []string {
	out := []string{"run", "--rm", "-v", s.Root + ":/work", "-w", s.Path(rel), "-e", "HOME=/tmp"}
	if runtime.GOOS != "windows" {
		out = append(out, "--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()))
	}
	for _, e := range s.Env {
		out = append(out, "-e", e)
	}
	return append(append(out, s.Image, name), args...)
}
