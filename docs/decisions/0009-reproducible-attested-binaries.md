# 0009: Reproducible, attested release binaries

- Status: accepted
- Date: 2026-11-04

## Context

`go install` needs a Go toolchain. Many CI users prefer a download. A download is
only trustworthy if people can check where it came from. Unsigned binaries also
trip antivirus false positives (seen on Windows during development).

## Decision

- `scripts/dist.sh` builds archives for Linux, macOS and Windows on amd64 and arm64,
  plus `checksums.txt`.
- Builds are reproducible: `-trimpath`, `CGO_ENABLED=0`, and every timestamp in the
  archives set to the commit time, with sorted entries and fixed owners. Rebuilding
  the same commit gives byte-identical files, so anyone can check a release.
- The release workflow signs SLSA build provenance for every file with
  `actions/attest` (keyless Sigstore signing with the workflow's OIDC identity).
  Users check a download with `gh attestation verify <file> --repo thusalapi/sealtag`.
- The GitHub release text comes from `sealtag notes`, which reads the section that
  `sealtag prepare` wrote into CHANGELOG.md.

## Consequences

- No signing keys to store or rotate: the signature is tied to the workflow run.
- Windows code signing (Authenticode) is a separate, paid step and is not done yet.
