# 0008: Publish with trusted publishing only

- Status: accepted
- Date: 2026-11-03

## Context

Stolen long-lived registry tokens caused the 2025 Shai-Hulud npm worm. Every major
registry now accepts short-lived tokens exchanged for a CI identity (OIDC).

## Decision

- `sealtag publish` runs in CI on tag pushes and uses OIDC for every registry:
  - **npm**: `npm publish` (npm 11.5.1+ performs the exchange and adds provenance).
  - **PyPI**: sealtag asks PyPI for the audience, gets a GitHub token, calls
    `/_/oidc/mint-token`, and uploads with `uv publish`.
  - **crates.io**: sealtag calls `/api/v1/trusted_publishing/tokens`, runs
    `cargo publish`, then revokes the token straight away.
  - **Go**: no upload exists. sealtag asks the module proxy for the new version so it
    is available at once.
- Short-lived tokens are passed through environment variables, never on the command
  line, and masked in the GitHub Actions log.
- If a long-lived token variable is set (`NPM_TOKEN`, `CARGO_REGISTRY_TOKEN`, ...),
  `publish` refuses to run unless `--allow-token` is given.
- Every tagged package is checked (`sealtag check`) before any upload starts, so a
  failure cannot leave a release half published.
- Tags must match manifest versions, or publishing stops.

## Consequences

- npm and crates.io need one manual first publish before a trusted publisher can be
  configured. `sealtag doctor` explains this per package.
- The OIDC exchange code is sealtag's own (`internal/oidc`), tested against a fake
  registry server, so the flow is visible and easy to study.
