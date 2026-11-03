// Package scan looks for secrets in files before they are published.
package scan

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Rule matches one kind of secret by its content.
type Rule struct {
	ID          string
	Description string
	Pattern     *regexp.Regexp
}

// Rules are the content patterns checked in every text file. Most real tokens have a
// fixed prefix, which keeps false positives low.
var Rules = []Rule{
	{"aws-access-key", "AWS access key ID", regexp.MustCompile(`\b(AKIA|ASIA)[0-9A-Z]{16}\b`)},
	{"github-token", "GitHub token", regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{36,}\b`)},
	{"github-fine-grained", "GitHub fine-grained token", regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{60,}\b`)},
	{"npm-token", "npm access token", regexp.MustCompile(`\bnpm_[A-Za-z0-9]{36}\b`)},
	{"pypi-token", "PyPI API token", regexp.MustCompile(`\bpypi-AgEIcHlwaS5vcmc[A-Za-z0-9_-]{50,}`)},
	{"crates-token", "crates.io API token", regexp.MustCompile(`\bcio[A-Za-z0-9]{32}\b`)},
	{"private-key", "Private key", regexp.MustCompile(`-----BEGIN ((RSA|EC|DSA|OPENSSH|PGP|ENCRYPTED) )?PRIVATE KEY( BLOCK)?-----`)},
	{"slack-token", "Slack token", regexp.MustCompile(`\bxox[abposr]-[A-Za-z0-9-]{10,}`)},
	{"stripe-live-key", "Stripe live secret key", regexp.MustCompile(`\b[sr]k_live_[A-Za-z0-9]{24,}\b`)},
	{"google-api-key", "Google API key", regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`)},
	// A literal token in .npmrc. "${NPM_TOKEN}" style references are fine.
	{"npmrc-auth", "Hard-coded registry auth in npm config", regexp.MustCompile(`_auth(Token)?\s*=\s*[^\s$"']{8,}`)},
}

// Finding is one suspected secret.
type Finding struct {
	File        string `json:"file"`           // relative to the root, forward slashes
	Line        int    `json:"line,omitempty"` // 0 when the finding is about the file name
	RuleID      string `json:"rule"`
	Description string `json:"description"`
	Match       string `json:"match,omitempty"` // redacted
}

const maxFileSize = 2 << 20 // 2 MiB. Larger files are almost never hand-written secrets.

// Files scans the given files. Paths are relative to root and use forward slashes.
func Files(root string, files []string) ([]Finding, error) {
	var out []Finding
	for _, rel := range files {
		fs, err := File(filepath.Join(root, filepath.FromSlash(rel)), rel)
		if err != nil {
			return nil, err
		}
		out = append(out, fs...)
	}
	return out, nil
}

// File scans one file. name is the path used in findings.
func File(path, name string) ([]Finding, error) {
	var out []Finding
	if f := checkName(name); f != nil {
		out = append(out, *f)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxFileSize {
		return out, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if isBinary(data) {
		return out, nil
	}
	return append(out, Content(data, name)...), nil
}

// Content applies every rule to data, line by line.
func Content(data []byte, name string) []Finding {
	var out []Finding
	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), maxFileSize)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		for _, r := range Rules {
			for _, m := range r.Pattern.FindAllString(line, -1) {
				out = append(out, Finding{File: name, Line: n, RuleID: r.ID, Description: r.Description, Match: Redact(m)})
			}
		}
	}
	return out
}

// checkName flags files whose name alone means they should never be published.
func checkName(name string) *Finding {
	base := strings.ToLower(filepath.Base(name))
	ext := filepath.Ext(base)
	f := &Finding{File: name}
	switch {
	case base == ".env" || (strings.HasPrefix(base, ".env.") && !isTemplate(base)):
		f.RuleID, f.Description = "env-file", "Environment file (often holds secrets)"
	case base == "id_rsa" || base == "id_ed25519" || base == "id_ecdsa" || base == "id_dsa":
		f.RuleID, f.Description = "ssh-key-file", "SSH private key file"
	case ext == ".pem" || ext == ".key" || ext == ".p12" || ext == ".pfx":
		f.RuleID, f.Description = "key-file", "Key or certificate bundle"
	case base == ".pypirc" || base == "credentials.json":
		f.RuleID, f.Description = "credentials-file", "Credentials or registry config file"
	default:
		return nil
	}
	return f
}

func isTemplate(base string) bool {
	for _, s := range []string{".example", ".sample", ".template", ".dist"} {
		if strings.HasSuffix(base, s) {
			return true
		}
	}
	return false
}

// isBinary uses the same trick as git: a NUL byte in the first 8000 bytes.
func isBinary(data []byte) bool {
	if len(data) > 8000 {
		data = data[:8000]
	}
	return bytes.IndexByte(data, 0) >= 0
}

// Redact keeps a short prefix so the user can find the secret, and hides the rest.
func Redact(s string) string {
	const keep = 6
	if len(s) <= keep {
		return strings.Repeat("*", len(s))
	}
	return s[:keep] + strings.Repeat("*", min(len(s)-keep, 20))
}
