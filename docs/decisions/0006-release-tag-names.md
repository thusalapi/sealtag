# 0006: Release tag names

- Status: accepted
- Date: 2026-11-03

## Context

sealtag finds the last release of each package by its git tag, and reads only the
commits after it. In a monorepo every package needs its own tag series.

## Decision

| Package | Tag | Example |
|---|---|---|
| Go module at the repo root | `v<version>` | `v1.4.0` |
| Go module in a folder | `<dir>/v<version>` | `sdk/go/v1.4.0` |
| Any other package, single-package repo | `v<version>` | `v1.4.0` |
| Any other package, monorepo | `<name>@v<version>` | `@acme/web@v2.5.0` |

Go's rule is fixed by the Go toolchain. `<name>@v<version>` is close to the style
used by changesets and Lerna, so it is familiar to JavaScript developers.

The version in the manifest stays the source of truth. If the manifest and the last
tag disagree, `plan` shows a note instead of guessing.

## Consequences

- Turning a single-package repo into a monorepo changes the tag style for non-Go
  packages. A future `--tag-prefix` option can pin the old style.
- Before 1.0.0, a breaking change bumps the minor version. Going to 1.0.0 needs an
  explicit `--bump major`.
