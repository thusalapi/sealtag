package gomod

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/module"

	"github.com/thusalapi/sealtag/internal/ecosystem"
)

// Proxy is the Go module proxy to notify.
var Proxy = "https://proxy.golang.org"

// Publish asks the Go module proxy for the new version. A Go module is published by
// pushing its tag, but the proxy only fetches it on first request. Requesting it here
// makes `go get` work immediately and adds it to the checksum database.
func (Ecosystem) Publish(ctx context.Context, dir string, p ecosystem.Package, opts ecosystem.PublishOptions) error {
	version := p.Version
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	escaped, err := module.EscapePath(p.Name)
	if err != nil {
		return err
	}
	u := fmt.Sprintf("%s/%s/@v/%s.info", strings.TrimRight(Proxy, "/"), escaped, version)
	if opts.DryRun {
		fmt.Fprintf(opts.Log, "would request %s\n", u)
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("module proxy returned HTTP %d for %s: is the tag pushed?", resp.StatusCode, u)
	}
	fmt.Fprintf(opts.Log, "%s %s is available on the module proxy\n", p.Name, version)
	return nil
}
