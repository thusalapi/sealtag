package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Verify packs the package, installs the tarball into an empty project and imports
// it with Node. Dynamic import() works for both CommonJS and ES modules.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, log io.Writer) error {
	tmp, err := os.MkdirTemp("", "sealtag-verify-npm-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	packDir := filepath.Join(tmp, "pack")
	appDir := filepath.Join(tmp, "app")
	for _, d := range []string{packDir, appDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	if err := ecosystem.Run(ctx, log, dir, nil, "npm", "pack", "--ignore-scripts", "--pack-destination", packDir); err != nil {
		return err
	}
	tgz, _ := filepath.Glob(filepath.Join(packDir, "*.tgz"))
	if len(tgz) != 1 {
		return fmt.Errorf("npm pack produced %d tarballs, want 1", len(tgz))
	}
	if err := os.WriteFile(filepath.Join(appDir, "package.json"), []byte(`{"name":"sealtag-verify","private":true}`), 0o644); err != nil {
		return err
	}
	// --ignore-scripts: the package under test must not run install scripts here.
	if err := ecosystem.Run(ctx, log, appDir, nil, "npm", "install", "--ignore-scripts", "--no-audit", "--no-fund", tgz[0]); err != nil {
		return err
	}
	name, _ := json.Marshal(p.Name)
	script := fmt.Sprintf("import(%s).then(() => console.log('imported %s')).catch(e => { console.error(e.message); process.exit(1) })", name, p.Name)
	return ecosystem.Run(ctx, log, appDir, nil, "node", "-e", script)
}
