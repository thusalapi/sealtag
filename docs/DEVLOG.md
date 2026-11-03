# Dev log

Newest entry first. One entry per work session: what was done, what was learned,
what is next.

## 2026-11-03: Rename, finish milestone 1, start milestone 2

**Done**
- Renamed the project to sealtag (ADR 0004). Shipwright is already a CNCF project.
- `--json` on every command.
- `scan` now asks git for the file list, so `.gitignore` is respected. Untracked but
  not ignored files are still scanned, because they can still be packed.
- `.sealtag-allow` allowlist with a required reason per entry.
- New `check` command: exact file list per package (npm and Go so far), secret scan
  of only those files, and warnings for large files.
- In-process CLI tests, including one that runs the real npm.

**Learned**
- `npm pack` runs lifecycle scripts unless told not to.
- Go module zips include `testdata/`.
- `modzip.CheckDir` returns paths joined with the directory, not relative ones.

**Next**
- Cargo file list (`cargo package --list`) and PyPI (build sdist and wheel, list contents).
- Then milestone 3: Conventional Commits and automatic version bumps.

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
