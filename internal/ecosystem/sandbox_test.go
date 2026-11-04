package ecosystem

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSandboxPaths(t *testing.T) {
	host := Sandbox{Root: filepath.FromSlash("/tmp/x")}
	if got := host.Path("dist/a.whl"); got != filepath.FromSlash("/tmp/x/dist/a.whl") {
		t.Errorf("host path = %q", got)
	}
	c := Sandbox{Root: "/tmp/x", Opts: VerifyOptions{Container: true}}
	if got := c.Path("dist/a.whl"); got != "/work/dist/a.whl" {
		t.Errorf("container path = %q", got)
	}
}

func TestContainerArgs(t *testing.T) {
	s := Sandbox{Root: "/tmp/x", Image: "node:24-slim", Env: []string{"A=1"}, Opts: VerifyOptions{Container: true}}
	got := strings.Join(s.ContainerArgs("app", "node", "-e", "1"), " ")
	for _, want := range []string{"run --rm -v /tmp/x:/work -w /work/app", "-e HOME=/tmp", "-e A=1", "node:24-slim node -e 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
	if runtime.GOOS != "windows" && !strings.Contains(got, "--user ") {
		t.Errorf("should run as the current user: %q", got)
	}
}
