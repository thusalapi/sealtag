# Architecture

## Overview

sealtag is a single Go binary. The CLI layer is thin. Each command calls into
packages under `internal/` that have no knowledge of the terminal, so they can be
tested directly.

```
cmd/sealtag/        entry point (main.go)
internal/cli/          cobra commands: detect, plan, scan, check, changelog, prepare, version
internal/ecosystem/    the Ecosystem interface and the shared Package type
internal/ecosystem/npm     package.json
internal/ecosystem/pypi    pyproject.toml
internal/ecosystem/cargo   Cargo.toml
internal/ecosystem/gomod   go.mod and git tags
internal/detect/       walks a repo and asks every ecosystem what it finds
internal/fileset/      lists project files (git-aware, respects .gitignore)
internal/semver/       parse and bump versions
internal/commits/      parse Conventional Commits, choose a bump
internal/gitlog/       read commits and tags for one package
internal/release/      plan each package's next version and render release notes
internal/prepare/      write versions and changelogs, optionally commit and tag
internal/textedit/     change one value in JSON, TOML or Python without reformatting
internal/testutil/     throwaway git repositories for tests
internal/scan/         secret detection rules and file scanner
testdata/              fixture repositories used by tests
```

## Key abstraction: Ecosystem

Every registry is an implementation of one interface:

```go
type Ecosystem interface {
    Name() string                          // "npm", "pypi", "cargo", "go"
    Detect(dir string) (*Package, error)   // nil if this dir has no package of this kind
}
```

Optional abilities are separate small interfaces, so ecosystems can add them one at a
time. The first is `Packer`:

```go
type Packer interface {
    ShippedFiles(dir string) ([]File, error) // exactly what a publish would upload
}
```

The second is `Versioner` (`SetVersion`). Go does not implement it, because a Go
version is only a tag. Milestone 4 adds `Publisher` the same way.

## Release flow

```
sealtag plan -> sealtag prepare --commit -> git push --follow-tags -> CI: sealtag publish
``` Adding a new
registry means adding one package that implements the interface. Nothing else changes.

## Data flow

```
repo dir -> detect.Walk -> []Package -> release.Build (tags + commits -> bump) -> table or JSON
repo dir -> fileset.List -> scan.Files -> scan.Filter (allowlist) -> table or JSON
```

## Design rules

1. No network calls in detect, plan or scan. They must work offline and in tests.
2. Every command will support `--json` so CI and other tools can use the output.
3. Fail closed: if sealtag cannot tell what will ship, it refuses to publish.
