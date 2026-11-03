package oidc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeRegistry plays GitHub's token service, PyPI and crates.io in one server.
func fakeRegistry(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /gh-token", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "bearer request-token" {
			http.Error(w, "bad request token", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"value": "jwt-for-" + r.URL.Query().Get("audience")})
	})
	mux.HandleFunc("GET /_/oidc/audience", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"audience": "pypi"})
	})
	mux.HandleFunc("POST /_/oidc/mint-token", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["token"] != "jwt-for-pypi" {
			http.Error(w, `{"message":"invalid-publisher"}`, http.StatusUnprocessableEntity)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true, "token": "pypi-short-lived"})
	})
	mux.HandleFunc("POST /api/v1/trusted_publishing/tokens", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["jwt"] != "jwt-for-crates.io" || !strings.HasPrefix(r.Header.Get("User-Agent"), "sealtag") {
			http.Error(w, "nope", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": "cio-short-lived"})
	})
	mux.HandleFunc("DELETE /api/v1/trusted_publishing/tokens", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "cio-short-lived" {
			http.Error(w, "unknown token", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func client(srv *httptest.Server) Client {
	env := map[string]string{
		"ACTIONS_ID_TOKEN_REQUEST_URL":   srv.URL + "/gh-token?api-version=2.0",
		"ACTIONS_ID_TOKEN_REQUEST_TOKEN": "request-token",
	}
	return Client{HTTP: srv.Client(), Getenv: func(k string) string { return env[k] }}
}

func TestPyPIFlow(t *testing.T) {
	srv := fakeRegistry(t)
	c, ctx := client(srv), context.Background()
	aud, err := c.PyPIAudience(ctx, srv.URL)
	if err != nil || aud != "pypi" {
		t.Fatalf("audience %q %v", aud, err)
	}
	jwt, err := c.GitHubToken(ctx, aud)
	if err != nil || jwt != "jwt-for-pypi" {
		t.Fatalf("jwt %q %v", jwt, err)
	}
	tok, err := c.PyPIToken(ctx, srv.URL, jwt)
	if err != nil || tok != "pypi-short-lived" {
		t.Fatalf("token %q %v", tok, err)
	}
}

func TestPyPIRejectionShowsRegistryMessage(t *testing.T) {
	srv := fakeRegistry(t)
	_, err := client(srv).PyPIToken(context.Background(), srv.URL, "wrong")
	if err == nil || !strings.Contains(err.Error(), "invalid-publisher") || !strings.Contains(err.Error(), "422") {
		t.Errorf("err = %v", err)
	}
}

func TestCratesFlow(t *testing.T) {
	srv := fakeRegistry(t)
	c, ctx := client(srv), context.Background()
	jwt, _ := c.GitHubToken(ctx, "crates.io")
	tok, err := c.CratesToken(ctx, srv.URL, jwt)
	if err != nil || tok != "cio-short-lived" {
		t.Fatalf("token %q %v", tok, err)
	}
	if err := c.RevokeCratesToken(ctx, srv.URL, tok); err != nil {
		t.Errorf("revoke: %v", err)
	}
}

func TestNotInCI(t *testing.T) {
	c := Client{Getenv: func(string) string { return "" }}
	if _, err := c.GitHubToken(context.Background(), "pypi"); !errors.Is(err, ErrNotInCI) {
		t.Errorf("err = %v", err)
	}
}

func TestMask(t *testing.T) {
	var b bytes.Buffer
	Mask(&b, "s3cret")
	Mask(&b, "")
	if b.String() != "::add-mask::s3cret\n" {
		t.Errorf("got %q", b.String())
	}
}
