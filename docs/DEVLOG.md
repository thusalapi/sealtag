# Dev log

Newest entry first. One entry per work session: what was done, what was learned,
what is next.

## 2026-11-04: Clean-room verify (milestone 5 starts)

**Done**
- `Verifier` for all four ecosystems and a `verify` command. `publish` runs it first.
- Tests reproduce real breakages: an npm `"files"` list missing `lib/`, a Python
  package importing a missing module, a Go module needing a gitignored file.
- Tool errors now include the last lines of stderr.

**Learned**
- `verify` immediately found a bug in our own fixture: `@acme/web` declared
  `"main": "index.js"` without shipping it.
- `modzip.Unzip` creates read-only files, like the module cache.

**Next**
- Push to GitHub (waiting on login) and run the first release.

## 2026-11-03: Trusted publishing (milestone 4, offline part)

**Done**
- `internal/oidc`: GitHub OIDC token request, PyPI and crates.io token exchange,
  crates.io token revocation, log masking. Tested against an `httptest` fake.
- `Publisher` for npm, PyPI, Cargo and Go (ADR 0008).
- `publish` command: finds release tags on HEAD, checks tags match manifests, runs
  `check` on every package first, refuses long-lived tokens, supports `--dry-run`.
- `doctor` command and `.github/workflows/release.yml`.

**Learned**
- PyPI uploads go to a different host (`upload.pypi.org`) from the index.
- Go module paths with capital letters are escaped as `!` plus lower case in proxy URLs.

**Next**
- Push to GitHub and run the first real release (sealtag v0.1.0) through Actions.

## 2026-11-03: prepare command, milestone 3 done

**Done**
- `internal/textedit`: change one value in JSON, TOML or a Python file without
  reformatting anything else.
- `Versioner` for npm (also package-lock.json), Cargo (including workspace-inherited
  versions) and PyPI (`[project]`, Poetry, or a hatch `__version__` file).
- Dynamic Python versions are now read, so `plan` knows their current version.
- `prepare` writes versions and changelogs. `prepare --commit` commits only those
  files and creates annotated tags, and refuses a dirty tree (ADR 0007).
- Shared `testutil.Repo` helper for git-based tests.

**Learned**
- `git commit -- .` commits every tracked change under the path, not just what is
  staged. The release commit must use no pathspec.

**Next**
- Milestone 4: trusted publishing from GitHub Actions. Needs the repo on GitHub.

## 2026-11-03: Automatic version bumps from commits

**Done**
- `internal/commits` parses Conventional Commits and picks the bump.
- `internal/gitlog` reads commits and tags per package, excluding nested packages.
- `internal/release` plans each package: last tag, releasable commits, next version,
  next tag (ADR 0006).
- `plan` now defaults to `--bump auto`. New `changelog` command prints release notes.
- Tests build throwaway git repositories to check monorepo scoping and tag handling.

**Learned**
- Go tags inside a folder keep their "v" after the prefix is stripped. Easy to lose.
- sealtag's own early commits were not conventional, so `sealtag plan` on itself
  finds nothing to release. From now on this repo uses Conventional Commits.

**Next**
- Write versions into manifests and release notes into `CHANGELOG.md`
  (a `version` or `prepare` command).
- Then milestone 4: trusted publishing.

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
