package scan

import (
	"bufio"
	"strings"
	"testing"
)

func parse(t *testing.T, s string) []AllowEntry {
	t.Helper()
	e, err := ParseAllow("test", bufio.NewScanner(strings.NewReader(s)))
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestParseAllow(t *testing.T) {
	e := parse(t, "# comment\n\ntestdata/*.pem key-file fake cert for tests\nfixtures/ * all fake\n")
	if len(e) != 2 || e[0].Reason != "fake cert for tests" || e[1].Line != 4 {
		t.Fatalf("got %+v", e)
	}
	for _, bad := range []string{"only-two fields\n", "[ key-file reason\n"} {
		if _, err := ParseAllow("t", bufio.NewScanner(strings.NewReader(bad))); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
}

func TestFilter(t *testing.T) {
	allow := parse(t, "testdata/*.pem key-file fake\nfixtures/ * fake\n")
	findings := []Finding{
		{File: "testdata/a.pem", RuleID: "key-file"},      // allowed by line 1
		{File: "testdata/a.pem", RuleID: "private-key"},   // different rule: kept
		{File: "fixtures/deep/x.env", RuleID: "env-file"}, // allowed by folder rule
		{File: "src/.env", RuleID: "env-file"},            // kept
	}
	kept, allowed := Filter(findings, allow)
	if len(kept) != 2 || len(allowed) != 2 {
		t.Fatalf("kept %+v allowed %+v", kept, allowed)
	}
	if kept[0].RuleID != "private-key" || kept[1].File != "src/.env" {
		t.Errorf("wrong findings kept: %+v", kept)
	}
}
