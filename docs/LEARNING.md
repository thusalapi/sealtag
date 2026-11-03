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

## To study next

- Sigstore: keyless signing with short-lived certificates
- SLSA provenance: a signed record of how and where an artifact was built
- How `npm pack` decides which files to include (`files`, `.npmignore`)
