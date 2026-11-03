# 0002: One interface per registry ecosystem

- Status: accepted
- Date: 2026-11-03

## Context

Each registry has different manifests, build steps and publish APIs. We want the core
logic (planning, scanning, publishing order) to stay the same for all of them.

## Decision

Every registry implements a small `Ecosystem` interface in its own package under
`internal/ecosystem/`. The core only talks to the interface.

## Consequences

- Adding a registry does not touch existing code.
- Each ecosystem can be tested alone against fixture directories.
- The interface will grow with each milestone, so keep it minimal and add methods
  only when a milestone needs them.
