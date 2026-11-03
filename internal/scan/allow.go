package scan

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

// AllowFile is the name of the allowlist file read from the repository root.
const AllowFile = ".sealtag-allow"

// AllowEntry suppresses findings for one rule in matching files.
//
// File format, one entry per line:
//
//	<path pattern>  <rule id or *>  <reason>
//	testdata/certs/*.pem  key-file  fake certificate used by tests
//
// A pattern ending in "/" matches everything below that folder. A reason is required,
// so every exception explains itself in code review.
type AllowEntry struct {
	Pattern string
	RuleID  string
	Reason  string
	Line    int
}

// LoadAllow reads root/.sealtag-allow. A missing file is not an error.
func LoadAllow(root string) ([]AllowEntry, error) {
	f, err := os.Open(root + "/" + AllowFile)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseAllow(f.Name(), bufio.NewScanner(f))
}

// ParseAllow parses allowlist lines. name is used in error messages.
func ParseAllow(name string, sc *bufio.Scanner) ([]AllowEntry, error) {
	var out []AllowEntry
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("%s:%d: want \"<pattern> <rule> <reason>\"", name, n)
		}
		if _, err := path.Match(fields[0], ""); err != nil {
			return nil, fmt.Errorf("%s:%d: bad pattern %q: %w", name, n, fields[0], err)
		}
		out = append(out, AllowEntry{Pattern: fields[0], RuleID: fields[1], Reason: strings.Join(fields[2:], " "), Line: n})
	}
	return out, sc.Err()
}

func (e AllowEntry) matches(f Finding) bool {
	if e.RuleID != "*" && e.RuleID != f.RuleID {
		return false
	}
	if strings.HasSuffix(e.Pattern, "/") {
		return strings.HasPrefix(f.File, e.Pattern)
	}
	ok, _ := path.Match(e.Pattern, f.File)
	return ok
}

// Filter splits findings into those still reported and those the allowlist covers.
func Filter(findings []Finding, allow []AllowEntry) (kept, allowed []Finding) {
outer:
	for _, f := range findings {
		for _, e := range allow {
			if e.matches(f) {
				allowed = append(allowed, f)
				continue outer
			}
		}
		kept = append(kept, f)
	}
	return kept, allowed
}
