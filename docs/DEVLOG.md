# Dev log

Newest entry first. One entry per work session: what was done, what was learned,
what is next.

## 2026-11-03: Project start

**Done**
- Chose the idea after researching developer pain points (see the project ideas doc).
- Created the repository, docs and decision records.
- Scaffolded the Go CLI with `detect`, `plan`, `scan` and `version` commands.
- Implemented detection for npm, PyPI, Cargo and Go modules with fixture tests.
- Implemented semver bumping and a first secret scanner.

**Learned**
- Go modules are published by git tags, so the "Go registry" step is a tag push.

**Next**
- Respect `.gitignore` in `scan`.
- Add `--json` output.
- Start milestone 2 with the npm file list.
