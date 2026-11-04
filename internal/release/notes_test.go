package release

import (
	"errors"
	"testing"
)

const sample = "# Changelog\r\n\r\nIntro.\r\n\r\n## lib v1.2.0 (2026-11-04)\r\n\r\n### Features\r\n\r\n- new thing (abc1234)\r\n\r\n## lib v1.1.0 (2026-10-01)\r\n\r\n### Fixes\r\n\r\n- old fix (def5678)\r\n"

func TestNotesFromChangelog(t *testing.T) {
	got, err := NotesFromChangelog(sample, "lib", "v1.2.0")
	if err != nil || got != "### Features\n\n- new thing (abc1234)\n" {
		t.Errorf("v1.2.0: %q %v", got, err)
	}
	got, err = NotesFromChangelog(sample, "lib", "v1.1.0")
	if err != nil || got != "### Fixes\n\n- old fix (def5678)\n" {
		t.Errorf("last section: %q %v", got, err)
	}
	if _, err := NotesFromChangelog(sample, "lib", "v1.2"); !errors.Is(err, ErrNoNotes) {
		t.Errorf("prefix of a version must not match: %v", err)
	}
}
