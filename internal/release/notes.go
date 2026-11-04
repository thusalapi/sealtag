package release

import (
	"errors"
	"strings"
)

// ErrNoNotes means the changelog has no section for the requested release.
var ErrNoNotes = errors.New("no changelog section for this release")

// NotesFromChangelog returns the body of the "## <name> <version> (...)" section
// written by `sealtag prepare`, without its heading. It is used for the text of a
// GitHub release, so the notes are written once and reused.
func NotesFromChangelog(changelog, name, version string) (string, error) {
	heading := "## " + name + " " + version
	lines := strings.Split(strings.ReplaceAll(changelog, "\r\n", "\n"), "\n")
	start := -1
	for i, l := range lines {
		if l == heading || strings.HasPrefix(l, heading+" (") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", ErrNoNotes
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines[start:end], "\n")) + "\n", nil
}
