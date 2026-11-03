package gitlog

import "testing"

func TestParseGitHubRemote(t *testing.T) {
	cases := map[string]string{
		"https://github.com/octo/lib.git":   "octo/lib",
		"https://github.com/octo/lib":       "octo/lib",
		"git@github.com:octo/lib.git":       "octo/lib",
		"ssh://git@github.com/octo/lib.git": "octo/lib",
		"https://gitlab.com/octo/lib.git":   "",
		"https://github.com/octo/lib/extra": "",
	}
	for in, want := range cases {
		if got := ParseGitHubRemote(in); got != want {
			t.Errorf("ParseGitHubRemote(%q) = %q, want %q", in, got, want)
		}
	}
}
