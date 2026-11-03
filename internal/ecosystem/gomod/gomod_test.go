package gomod

import "testing"

func TestModulePath(t *testing.T) {
	cases := map[string]string{
		"module example.com/a\n\ngo 1.22\n":     "example.com/a",
		"// comment\nmodule \"example.com/b\"\n": "example.com/b",
		"go 1.22\n":                              "",
		"modulex example.com/c\n":                "",
	}
	for in, want := range cases {
		if got := ModulePath([]byte(in)); got != want {
			t.Errorf("ModulePath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHighestTag(t *testing.T) {
	tags := []string{"v1.0.0", "v1.10.0", "v1.9.3", "v2.0.0-rc.1", "sdk/v3.0.0", "not-a-version"}
	// Pre-releases still count: v2.0.0-rc.1 is higher than v1.10.0.
	if got, ok := HighestTag(tags, ""); !ok || got != "v2.0.0-rc.1" {
		t.Errorf("root: got %q, %v", got, ok)
	}
	if got, ok := HighestTag(tags, "sdk/"); !ok || got != "v3.0.0" {
		t.Errorf("sdk: got %q, %v", got, ok)
	}
	if _, ok := HighestTag(tags, "other/"); ok {
		t.Error("other/: expected no tag")
	}
}
