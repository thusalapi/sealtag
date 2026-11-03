# Roadmap

Each milestone ends with something that works end to end and can be shown to someone.

## Milestone 1: Detect, plan, scan (in progress)

- [x] Project skeleton, docs, CI
- [x] Detect packages for npm, PyPI, Cargo and Go modules, including monorepos
- [x] `plan` command: current version and next version for a semver bump
- [x] `scan` command: find secrets (cloud keys, registry tokens, private keys) in files
- [ ] Respect `.gitignore` when scanning
- [ ] Allowlist file for known false positives (see ADR 0003)
- [ ] JSON output (`--json`) for every command so CI can consume it

## Milestone 2: Know exactly what ships

- [ ] npm: file list from `npm pack --dry-run --json`
- [ ] PyPI: build sdist and wheel, list their contents
- [ ] Cargo: `cargo package --list`
- [ ] Go: module zip rules (what `go mod` would include)
- [ ] `check` command: file list plus secret scan plus size warnings

## Milestone 3: Versions and changelogs

- [ ] Read Conventional Commits since the last tag
- [ ] Decide the bump automatically (fix = patch, feat = minor, breaking = major)
- [ ] Write the new version into each manifest
- [ ] Generate `CHANGELOG.md`

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
