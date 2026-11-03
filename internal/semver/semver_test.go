package semver

import "testing"

func TestParseRoundTrip(t *testing.T) {
	for _, s := range []string{"0.0.0", "1.2.3", "v1.2.3", "1.2.3-rc.1", "1.2.3-rc.1+build.5", "10.20.30"} {
		v, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q): %v", s, err)
		}
		if got := v.String(); got != s {
			t.Errorf("round trip %q gave %q", s, got)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, s := range []string{"", "1", "1.2", "1.2.3.4", "01.2.3", "1.2.x", "1.2.3-", "a.b.c"} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) should fail", s)
		}
	}
}

func TestCompare(t *testing.T) {
	// Ordered lowest to highest, taken from the example in the semver spec.
	order := []string{
		"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-beta",
		"1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0", "1.0.1", "1.1.0", "2.0.0",
	}
	for i := 0; i < len(order)-1; i++ {
		a, _ := Parse(order[i])
		b, _ := Parse(order[i+1])
		if a.Compare(b) != -1 || b.Compare(a) != 1 {
			t.Errorf("expected %s < %s", order[i], order[i+1])
		}
	}
}

func TestBump(t *testing.T) {
	cases := []struct{ in, kind, want string }{
		{"1.2.3", Patch, "1.2.4"},
		{"1.2.3", Minor, "1.3.0"},
		{"1.2.3", Major, "2.0.0"},
		{"v0.4.9", Minor, "v0.5.0"},
		{"1.3.0-rc.1", Minor, "1.3.0"},
		{"1.2.4-rc.1", Patch, "1.2.4"},
		{"2.0.0-beta", Major, "2.0.0"},
		{"1.2.3+meta", Patch, "1.2.4"},
	}
	for _, c := range cases {
		v, _ := Parse(c.in)
		got, err := v.Bump(c.kind)
		if err != nil || got.String() != c.want {
			t.Errorf("%s bump %s = %s, %v; want %s", c.in, c.kind, got, err, c.want)
		}
	}
	v, _ := Parse("1.0.0")
	if _, err := v.Bump("huge"); err == nil {
		t.Error("unknown bump kind should fail")
	}
}
