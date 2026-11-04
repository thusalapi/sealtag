package npm

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

func TestVerifyRealNpm(t *testing.T) {
	if _, err := exec.LookPath("npm"); err != nil || testing.Short() {
		t.Skip("needs npm")
	}
	dir := t.TempDir()
	// "files" forgets lib/, which index.js needs: works locally, breaks once published.
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"sealtag-demo","version":"1.0.0","main":"index.js","files":["index.js"]}`), 0o644)
	os.WriteFile(filepath.Join(dir, "index.js"), []byte("module.exports = require('./lib/impl')\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, "lib"), 0o755)
	os.WriteFile(filepath.Join(dir, "lib", "impl.js"), []byte("module.exports = 42\n"), 0o644)
	p := ecosystem.Package{Ecosystem: "npm", Name: "sealtag-demo"}

	if err := (Ecosystem{}).Verify(context.Background(), dir, p, ecosystem.VerifyOptions{Log: io.Discard}); err == nil {
		t.Fatal("verify should fail: lib/ is not in \"files\"")
	}
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"sealtag-demo","version":"1.0.0","main":"index.js","files":["index.js","lib"]}`), 0o644)
	if err := (Ecosystem{}).Verify(context.Background(), dir, p, ecosystem.VerifyOptions{Log: io.Discard}); err != nil {
		t.Errorf("verify should pass: %v", err)
	}
}
