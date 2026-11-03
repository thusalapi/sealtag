# Roadmap

Each milestone ends with something that works end to end and can be shown to someone.

## Milestone 1: Detect, plan, scan (done)

- [x] Project skeleton, docs, CI
- [x] Detect packages for npm, PyPI, Cargo and Go modules, including monorepos
- [x] `plan` command: current version and next version for a semver bump
- [x] `scan` command: find secrets (cloud keys, registry tokens, private keys) in files
- [x] Respect `.gitignore` when scanning
- [x] Allowlist file for known false positives (see ADR 0003)
- [x] JSON output (`--json`) for every command so CI can consume it

## Milestone 2: Know exactly what ships (done)

- [x] npm: file list from `npm pack --dry-run --json --ignore-scripts`
- [x] PyPI: build sdist and wheel with uv, scan the bytes inside the archives
- [x] Cargo: `cargo package --list`, with generated files marked
- [x] Go: module zip rules via golang.org/x/mod/zip, limited to git-tracked files
- [x] `check` command: file list plus secret scan plus size warnings

## Milestone 3: Versions and changelogs (done)

- [x] Read Conventional Commits since each package's last tag (ADR 0006)
- [x] Decide the bump automatically (fix = patch, feat = minor, breaking = major)
- [x] Write the new version into each manifest, keeping formatting
- [x] `changelog` command prints release notes
- [x] Write release notes into each package's `CHANGELOG.md`
- [x] `prepare --commit`: release commit plus annotated tags (ADR 0007)
- [x] Read and write dynamic Python versions (hatch `[tool.hatch.version]`)

## Milestone 4: Trusted publishing

- [ ] Read the GitHub Actions OIDC token
- [ ] Exchange it with PyPI, npm and crates.io for short-lived publish tokens
- [ ] Publish to each registry
- [ ] Refuse to run with a long-lived token unless `--allow-token` is passed
- [ ] `doctor` command that explains how to configure trusted publishing on each registry

## Milestone 5: Provenance and clean-room testing

- [ ] Sign artifacts and attach provenance (Sigstore, SLSA)
- [ ] Install each built package in a clean container and run a smoke test

## Milestone 6: Launch

- [ ] Reusable GitHub Action (`uses: thusalapi/sealtag@v1`)
- [ ] Release sealtag with sealtag
- [ ] Docs site and launch blog post
