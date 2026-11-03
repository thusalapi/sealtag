// Package oidc implements trusted publishing: it gets a short-lived identity token
// from the CI provider and exchanges it with a registry for a short-lived upload
// token. No long-lived secret is ever stored.
//
//	GitHub Actions --(OIDC JWT, audience "pypi")--> PyPI --(token valid ~15 min)--> upload
package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// ErrNotInCI means no OIDC provider is available. In GitHub Actions this usually
// means the workflow is missing `permissions: id-token: write`.
var ErrNotInCI = errors.New("no OIDC provider found: in GitHub Actions add `permissions: id-token: write` to the job")

// UserAgent identifies sealtag to registries, as their API policies ask.
const UserAgent = "sealtag (https://github.com/thusalapi/sealtag)"

// Client talks to the CI provider and registries. The zero value uses
// http.DefaultClient and the real environment.
type Client struct {
	HTTP   *http.Client
	Getenv func(string) string
}

func (c Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c Client) getenv(k string) string {
	if c.Getenv != nil {
		return c.Getenv(k)
	}
	return os.Getenv(k)
}

// InGitHubActions reports whether a GitHub OIDC token can be requested.
func (c Client) InGitHubActions() bool {
	return c.getenv("ACTIONS_ID_TOKEN_REQUEST_URL") != "" && c.getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN") != ""
}

// GitHubToken requests a GitHub Actions OIDC token (a signed JWT) for audience.
// The registry checks the signature against GitHub's public keys and reads claims
// such as the repository and workflow file to decide whether to trust it.
func (c Client) GitHubToken(ctx context.Context, audience string) (string, error) {
	if !c.InGitHubActions() {
		return "", ErrNotInCI
	}
	u, err := url.Parse(c.getenv("ACTIONS_ID_TOKEN_REQUEST_URL"))
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("audience", audience)
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "bearer "+c.getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN"))
	var resp struct {
		Value string `json:"value"`
	}
	if err := c.do(req, &resp); err != nil {
		return "", fmt.Errorf("requesting GitHub OIDC token: %w", err)
	}
	if resp.Value == "" {
		return "", errors.New("GitHub returned an empty OIDC token")
	}
	return resp.Value, nil
}

// PyPIToken exchanges an OIDC token for a PyPI upload token.
// base is "https://pypi.org" or "https://test.pypi.org".
func (c Client) PyPIToken(ctx context.Context, base, jwt string) (string, error) {
	body, _ := json.Marshal(map[string]string{"token": jwt})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/_/oidc/mint-token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp struct {
		Token string `json:"token"`
	}
	if err := c.do(req, &resp); err != nil {
		return "", fmt.Errorf("PyPI token exchange: %w", err)
	}
	return resp.Token, nil
}

// PyPIAudience asks the index which audience its OIDC tokens must have.
func (c Client) PyPIAudience(ctx context.Context, base string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(base, "/")+"/_/oidc/audience", nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		Audience string `json:"audience"`
	}
	if err := c.do(req, &resp); err != nil {
		return "", fmt.Errorf("PyPI audience: %w", err)
	}
	return resp.Audience, nil
}

// CratesToken exchanges an OIDC token (audience "crates.io") for a crates.io token.
// The token should be revoked with RevokeCratesToken after publishing.
func (c Client) CratesToken(ctx context.Context, base, jwt string) (string, error) {
	body, _ := json.Marshal(map[string]string{"jwt": jwt})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/api/v1/trusted_publishing/tokens", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp struct {
		Token string `json:"token"`
	}
	if err := c.do(req, &resp); err != nil {
		return "", fmt.Errorf("crates.io token exchange: %w", err)
	}
	return resp.Token, nil
}

// RevokeCratesToken invalidates a token from CratesToken before it expires.
func (c Client) RevokeCratesToken(ctx context.Context, base, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, strings.TrimRight(base, "/")+"/api/v1/trusted_publishing/tokens", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	return c.do(req, nil)
}

func (c Client) do(req *http.Request, out any) error {
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Registry error bodies explain what is misconfigured, for example a
		// workflow name that does not match the trusted publisher settings.
		return fmt.Errorf("%s %s: HTTP %d: %s", req.Method, req.URL.Path, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// Mask tells GitHub Actions to hide a secret in all later log output.
func Mask(w io.Writer, secret string) {
	if secret != "" {
		fmt.Fprintf(w, "::add-mask::%s\n", secret)
	}
}
