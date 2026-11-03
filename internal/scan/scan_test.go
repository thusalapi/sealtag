package scan

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Fake secrets are built at runtime from pieces so this file never contains a
// string that looks like a real token. That keeps GitHub push protection and
// other scanners quiet about our own test code.
func fake(prefix string, n int) string { return prefix + strings.Repeat("A", n) }

func TestContentRules(t *testing.T) {
	cases := map[string]string{
		"aws-access-key":  "key = " + fake("AK"+"IA", 16),
		"github-token":    "token: " + fake("gh"+"p_", 36),
		"npm-token":       "//registry.npmjs.org/:_authToken=" + fake("np"+"m_", 36),
		"private-key":     "-----BEGIN " + "RSA PRIVATE KEY-----",
		"stripe-live-key": fake("sk"+"_live_", 24),
		"google-api-key":  fake("AI"+"za", 35),
	}
	for rule, line := range cases {
		fs := Content([]byte("first line\n"+line+"\n"), "f.txt")
		found := false
		for _, f := range fs {
			if f.RuleID == rule {
				found = true
				if f.Line != 2 {
					t.Errorf("%s: line %d, want 2", rule, f.Line)
				}
				if strings.Count(f.Match, "*") == 0 {
					t.Errorf("%s: match not redacted: %q", rule, f.Match)
				}
			}
		}
		if !found {
			t.Errorf("rule %s did not fire on %q", rule, line)
		}
	}
}

func TestNoFalsePositives(t *testing.T) {
	clean := `const AKIA = "short";
//registry.npmjs.org/:_authToken=${NPM_TOKEN}
password := os.Getenv("PASSWORD")
`
	if fs := Content([]byte(clean), "clean.go"); len(fs) != 0 {
		t.Errorf("expected no findings, got %+v", fs)
	}
}

func TestFileNames(t *testing.T) {
	flagged := []string{".env", ".env.production", "certs/server.pem", "id_rsa", ".pypirc"}
	for _, n := range flagged {
		if checkName(n) == nil {
			t.Errorf("%s should be flagged", n)
		}
	}
	for _, n := range []string{".env.example", "env.go", "README.md", ".npmrc"} {
		if f := checkName(n); f != nil {
			t.Errorf("%s should not be flagged: %+v", n, f)
		}
	}
}

func TestFilesSkipsBinary(t *testing.T) {
	root := t.TempDir()
	write := func(rel string, data []byte) {
		p := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	secret := []byte(fake("gh"+"p_", 40))
	write("src/config.js", secret)
	write("img.bin", append([]byte{0, 1, 2}, secret...)) // binary file
	fs, err := Files(root, []string{"src/config.js", "img.bin"})
	if err != nil {
		t.Fatal(err)
	}
	if len(fs) != 1 || fs[0].File != "src/config.js" {
		t.Fatalf("want one finding in src/config.js, got %+v", fs)
	}
}

func TestRedact(t *testing.T) {
	if got := Redact("abc"); got != "***" {
		t.Errorf("Redact short = %q", got)
	}
	if got := Redact("abcdefghij"); got != "abcdef****" {
		t.Errorf("Redact = %q", got)
	}
}
