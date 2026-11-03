# Dev log

Newest entry first. One entry per work session: what was done, what was learned,
what is next.

## 2026-11-03: PyPI file listing, milestone 2 done

**Done**
- `check` now builds Python packages with uv and scans the files inside the sdist
  and the wheel (ADR 0005). All four ecosystems pass `check` with real tools on
  Windows.
- Files read from archives are scanned in memory, so generated files like
  `PKG-INFO` are covered too.

**Learned**
- hatchling put a stray `.env` into an sdist when there was no `.gitignore`. This is
  now a test and a good demo for the README.

**Next**
- Milestone 3: parse Conventional Commits and choose the bump automatically.

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

- Cargo file list via `cargo package --list`. Tested against real cargo on Windows.
- Installed Go 1.27.3 on Windows (checksum verified) so sealtag builds and tests
  natively there, not only in Linux.

**Learned**
- `npm pack` runs lifecycle scripts unless told not to.
- `cargo package --list` names files that are not on disk (`Cargo.toml.orig`,
  `.cargo_vcs_info.json`, the workspace `Cargo.lock`). Testing with the real tool
  found this. A mock would have hidden it.
- Go module zips include `testdata/`.
- `modzip.CheckDir` returns paths joined with the directory, not relative ones.

**Next**
- PyPI file list (build sdist and wheel, list contents).
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
