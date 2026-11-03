# 0004: Rename the project from shipwright to sealtag

- Status: accepted
- Date: 2026-11-03

## Context

"Shipwright" is already the name of a CNCF project (a framework for building container
images on Kubernetes). Sharing a name with a known project in the same space confuses
search results and could cause trademark problems.

## Decision

Rename to **sealtag**. On 2026-11-03 the name was free on npm, PyPI, crates.io,
RubyGems and Homebrew, and no developer tool with that name showed up in a web search.

## Consequences

- Module path is now `github.com/thusalapi/sealtag`, binary is `sealtag`.
- The name describes the job: seal a release (check and sign it), then tag and publish.
- Reserve the name on npm, PyPI and crates.io before the public launch.
