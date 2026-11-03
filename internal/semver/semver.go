// Package semver parses, compares and bumps semantic versions (https://semver.org).
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a parsed semantic version. VPrefix records a leading "v" so it can be
// written back the same way (Go tags use "v1.2.3", npm uses "1.2.3").
type Version struct {
	Major, Minor, Patch uint64
	Pre                 string // "rc.1" in 1.2.3-rc.1
	Build               string // "abc" in 1.2.3+abc
	VPrefix             bool
}

// Parse reads a version such as "1.2.3", "v1.2.3" or "1.2.3-rc.1+build.5".
func Parse(s string) (Version, error) {
	var v Version
	orig := s
	if strings.HasPrefix(s, "v") {
		v.VPrefix = true
		s = s[1:]
	}
	s, v.Build, _ = strings.Cut(s, "+")
	s, v.Pre, _ = strings.Cut(s, "-")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: want MAJOR.MINOR.PATCH", orig)
	}
	nums := [3]*uint64{&v.Major, &v.Minor, &v.Patch}
	for i, p := range parts {
		if p == "" || (len(p) > 1 && p[0] == '0') {
			return Version{}, fmt.Errorf("invalid version %q: bad number %q", orig, p)
		}
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("invalid version %q: bad number %q", orig, p)
		}
		*nums[i] = n
	}
	if strings.Contains(orig, "-") && v.Pre == "" {
		return Version{}, fmt.Errorf("invalid version %q: empty pre-release", orig)
	}
	return v, nil
}

// String formats the version, keeping the original "v" prefix.
func (v Version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		s += "-" + v.Pre
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	if v.VPrefix {
		s = "v" + s
	}
	return s
}

// Compare returns -1, 0 or 1. Build metadata is ignored, as the spec requires.
func (v Version) Compare(o Version) int {
	for _, d := range [3][2]uint64{{v.Major, o.Major}, {v.Minor, o.Minor}, {v.Patch, o.Patch}} {
		if d[0] != d[1] {
			if d[0] < d[1] {
				return -1
			}
			return 1
		}
	}
	return comparePre(v.Pre, o.Pre)
}

// A release is higher than any of its pre-releases: 1.0.0 > 1.0.0-rc.1.
func comparePre(a, b string) int {
	switch {
	case a == b:
		return 0
	case a == "":
		return 1
	case b == "":
		return -1
	}
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		an, aErr := strconv.ParseUint(as[i], 10, 64)
		bn, bErr := strconv.ParseUint(bs[i], 10, 64)
		switch {
		case aErr == nil && bErr == nil:
			if an != bn {
				if an < bn {
					return -1
				}
				return 1
			}
		case aErr == nil: // numeric identifiers sort before alphanumeric ones
			return -1
		case bErr == nil:
			return 1
		default:
			if c := strings.Compare(as[i], bs[i]); c != 0 {
				return c
			}
		}
	}
	switch {
	case len(as) < len(bs):
		return -1
	case len(as) > len(bs):
		return 1
	}
	return 0
}

// Bump kinds accepted by Bump.
const (
	Major = "major"
	Minor = "minor"
	Patch = "patch"
)

// Bump returns the next version. Pre-release and build data are dropped.
// Releasing a pre-release (1.2.0-rc.1) with the matching bump gives 1.2.0.
func (v Version) Bump(kind string) (Version, error) {
	n := Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch, VPrefix: v.VPrefix}
	isPre := v.Pre != ""
	switch kind {
	case Major:
		if !(isPre && v.Minor == 0 && v.Patch == 0) {
			n.Major, n.Minor, n.Patch = v.Major+1, 0, 0
		}
	case Minor:
		if !(isPre && v.Patch == 0) {
			n.Minor, n.Patch = v.Minor+1, 0
		}
	case Patch:
		if !isPre {
			n.Patch = v.Patch + 1
		}
	default:
		return Version{}, fmt.Errorf("unknown bump %q: use major, minor or patch", kind)
	}
	return n, nil
}
