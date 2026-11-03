# Learning notes

Concepts met while building sealtag, in plain words. Add to this as you go. These
notes are also good material for blog posts and interview answers.

## Trusted publishing (OIDC)

OIDC (OpenID Connect) lets one system prove its identity to another with a signed,
short-lived token. GitHub Actions can issue a token that says "this is workflow
`release.yml` in repo `owner/name`". A registry that trusts GitHub checks the signature
and the claims, then hands back a publish token that lasts minutes. Nothing long-lived
is stored, so there is nothing to steal.

Think of it like a hotel key card that only works today, instead of a metal key that
works forever.

## Semantic versioning

`MAJOR.MINOR.PATCH`. Bump PATCH for fixes, MINOR for new features that do not break
anyone, MAJOR for breaking changes. Pre-releases look like `1.2.0-rc.1`.

## Go module versions

Go has no central upload step. A Go module version is just a git tag such as `v1.4.0`.
Modules in a subfolder use a prefixed tag such as `sdk/v1.4.0`. Major versions 2 and
above must also change the module path (`example.com/lib/v2`).

## Secret scanning

Most leaked secrets have a recognisable shape. AWS access keys start with `AKIA`,
GitHub tokens start with `ghp_`, npm tokens start with `npm_`. A scanner matches these
shapes with regular expressions. Generic secrets are harder and need entropy checks.

## What npm actually publishes

npm decides the file list from the `files` field in package.json, then `.npmignore`
(or `.gitignore` if there is no `.npmignore`), plus files it always includes
(`package.json`, `README`, `LICENSE`) and files it never includes (`.git`,
`node_modules`). Without a `files` field, everything not ignored is published,
including a stray `.env`. The only fully accurate answer comes from npm itself:
`npm pack --dry-run --json`.

`npm pack` also runs the package's `prepack` and `prepare` scripts. sealtag passes
`--ignore-scripts` so that a "dry run" can never run someone else's code.

## What a Go module zip contains

The Go proxy builds a zip from the git commit for the tag. The rules live in
`golang.org/x/mod/zip`: no `vendor/` folders, no nested modules (folders with their
own `go.mod`), file names must be valid, and the zip must stay under 500 MB.
`testdata/` *is* included, which surprises many people.

## What a Python package contains

A Python release usually has two files. The **sdist** (`.tar.gz`) is the source code
plus metadata. The **wheel** (`.whl`, really a zip) is ready to install. PyPI and
tools like uv build the wheel *from the sdist*, so anything in the sdist matters.

Which files go in is decided by the build backend named in `pyproject.toml`. For
example, hatchling includes everything that is not ignored by `.gitignore`. In a
folder without a `.gitignore`, a stray `.env` ends up in the sdist. sealtag's tests
prove this with a real build.

## Conventional Commits

A commit message format that tools can read: `type(scope)!: description`.
`fix` means a bug fix (patch), `feat` a new feature (minor), and a `!` or a
`BREAKING CHANGE:` footer means a breaking change (major). Other types such as
`docs`, `chore` and `ci` do not cause a release. Example: `feat(cli): add --json`.

Before 1.0.0, most tools (and Cargo's own compatibility rules) treat the minor
number as the "breaking" number, so a breaking change goes from 0.4.2 to 0.5.0.

## Reading git history from a program

`git log --format=...` can print any fields with custom separators. sealtag uses the
ASCII "unit separator" (0x1F) between fields and "record separator" (0x1E) between
commits, because those characters never appear in normal messages. A pathspec such
as `-- packages/a ':(exclude)packages/a/plugin'` limits the log to one package.

## To study next

- Sigstore: keyless signing with short-lived certificates
- SLSA provenance: a signed record of how and where an artifact was built
- How `npm pack` decides which files to include (`files`, `.npmignore`)
