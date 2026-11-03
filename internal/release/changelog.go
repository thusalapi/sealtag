package release

import (
	"fmt"
	"strings"

	"github.com/thusalapi/sealtag/internal/commits"
)

// Changelog renders the Markdown section for one release, in the style of
// https://keepachangelog.com. date is formatted by the caller (YYYY-MM-DD).
func Changelog(r Release, date string) string {
	if r.Next == "" {
		return ""
	}
	var breaking, feats, fixes []commits.Commit
	for _, c := range r.Commits {
		switch {
		case c.Breaking:
			breaking = append(breaking, c)
		case c.Type == "feat":
			feats = append(feats, c)
		default: // fix and perf
			fixes = append(fixes, c)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## %s %s (%s)\n", r.Package.Name, r.Next, date)
	section(&b, "Breaking changes", breaking)
	section(&b, "Features", feats)
	section(&b, "Fixes", fixes)
	return b.String()
}

func section(b *strings.Builder, title string, cs []commits.Commit) {
	if len(cs) == 0 {
		return
	}
	fmt.Fprintf(b, "\n### %s\n\n", title)
	for _, c := range cs {
		line := c.Description
		if c.Scope != "" {
			line = "**" + c.Scope + ":** " + line
		}
		fmt.Fprintf(b, "- %s (%s)\n", line, short(c.Hash))
	}
}

func short(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
