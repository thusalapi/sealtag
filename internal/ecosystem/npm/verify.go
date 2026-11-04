package npm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Image is the container used by `verify --container`.
const Image = "node:24-slim"

// Verify packs the package on the host, then installs the tarball into an empty
// project and imports it with Node, on the host or in a container. Dynamic import()
// works for both CommonJS and ES modules.
func (Ecosystem) Verify(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.VerifyOptions) error {
	tmp, err := os.MkdirTemp("", "sealtag-verify-npm-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for _, d := range []string{"pack", "app"} {
		if err := os.MkdirAll(filepath.Join(tmp, d), 0o755); err != nil {
			return err
		}
	}
	if err := ecosystem.Run(ctx, opts.Log, dir, nil, "npm", "pack", "--ignore-scripts", "--pack-destination", filepath.Join(tmp, "pack")); err != nil {
		return err
	}
	tgz, _ := filepath.Glob(filepath.Join(tmp, "pack", "*.tgz"))
	if len(tgz) != 1 {
		return fmt.Errorf("npm pack produced %d tarballs, want 1", len(tgz))
	}
	if err := os.WriteFile(filepath.Join(tmp, "app", "package.json"), []byte(`{"name":"sealtag-verify","private":true}`), 0o644); err != nil {
		return err
	}
	sb := ecosystem.Sandbox{Opts: opts, Image: Image, Root: tmp}
	// --ignore-scripts: the package under test must not run install scripts here.
	if err := sb.Run(ctx, "app", "npm", "install", "--ignore-scripts", "--no-audit", "--no-fund", sb.Path("pack/"+filepath.Base(tgz[0]))); err != nil {
		return err
	}
	name, _ := json.Marshal(p.Name)
	script := fmt.Sprintf("import(%s).then(() => console.log('imported %s')).catch(e => { console.error(e.message); process.exit(1) })", name, p.Name)
	return sb.Run(ctx, "app", "node", "-e", script)
}
