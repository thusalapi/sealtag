package commits

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		subject, body string
		want          Commit
	}{
		{"feat(parser): add arrays", "", Commit{Type: "feat", Scope: "parser", Description: "add arrays"}},
		{"fix: handle empty input", "", Commit{Type: "fix", Description: "handle empty input"}},
		{"feat!: drop Node 16", "", Commit{Type: "feat", Breaking: true, Description: "drop Node 16"}},
		{"refactor(core)!: new API", "", Commit{Type: "refactor", Scope: "core", Breaking: true, Description: "new API"}},
		{"fix: x", "details\n\nBREAKING CHANGE: config renamed", Commit{Type: "fix", Breaking: true, Description: "x"}},
		{"Feat: capitalised type", "", Commit{Type: "feat", Description: "capitalised type"}},
		{"Update README", "", Commit{Description: "Update README"}},
		{"fix:missing space", "", Commit{Description: "fix:missing space"}},
	}
	for _, c := range cases {
		got := Parse("h", c.subject, c.body)
		if got.Type != c.want.Type || got.Scope != c.want.Scope || got.Breaking != c.want.Breaking || got.Description != c.want.Description {
			t.Errorf("Parse(%q) = %+v, want %+v", c.subject, got, c.want)
		}
	}
}

func TestBump(t *testing.T) {
	p := func(s string) Commit { return Parse("h", s, "") }
	cases := []struct {
		subjects []string
		want     string
	}{
		{nil, None},
		{[]string{"chore: deps", "docs: typo"}, None},
		{[]string{"chore: deps", "fix: bug"}, Patch},
		{[]string{"fix: bug", "feat: thing", "perf: faster"}, Minor},
		{[]string{"fix: bug", "feat!: break"}, Major},
	}
	for _, c := range cases {
		var cs []Commit
		for _, s := range c.subjects {
			cs = append(cs, p(s))
		}
		if got := Bump(cs); got != c.want {
			t.Errorf("Bump(%v) = %q, want %q", c.subjects, got, c.want)
		}
	}
	if n := len(Releasable([]Commit{p("chore: x"), p("feat: y"), p("fix: z")})); n != 2 {
		t.Errorf("Releasable kept %d, want 2", n)
	}
}
