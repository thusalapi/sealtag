# 0001: Write shipwright in Go

- Status: accepted
- Date: 2026-11-03

## Context

shipwright is a CLI that runs on developer laptops and in CI on Linux, macOS and
Windows. Users should not need a language runtime installed to use it.

## Decision

Use Go.

## Consequences

- One static binary per platform, easy to download in CI.
- Strong standard library for files, HTTP, JSON and crypto.
- The Go supply-chain ecosystem (Sigstore, goreleaser, SLSA tooling) is mostly written
  in Go, so we can reuse libraries and read their source.
- Rust would give a smaller binary but a slower learning curve for this project.
