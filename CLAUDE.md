# Notes for AI assistants working on this repo

- Language: Go. Keep dependencies few. Current: cobra (CLI), BurntSushi/toml.
- Layout and the Ecosystem interface are described in docs/ARCHITECTURE.md.
- Before finishing a work session, update:
  1. docs/ROADMAP.md checkboxes
  2. docs/DEVLOG.md with a new dated entry at the top
  3. docs/decisions/ with a new ADR for any significant choice
  4. docs/LEARNING.md with any new concept explained in plain words
- Run `go vet ./...` and `go test ./...` before committing.
- No network calls in detect, plan or scan code paths.
- Writing style for docs: plain words, short sentences, no em dashes.
