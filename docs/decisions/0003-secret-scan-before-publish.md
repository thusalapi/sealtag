# 0003: Scan for secrets before any publish

- Status: accepted
- Date: 2026-11-03

## Context

Packages sometimes ship `.env` files, keys or tokens by accident. Once a version is
published it is mirrored and cached, so removing it later does not undo the leak.

## Decision

shipwright scans every file that would ship and blocks the release on any finding,
unless the finding is listed in an allowlist file with a reason.

## Consequences

- Some false positives. The allowlist keeps them manageable.
- Rules are simple regular expressions plus file-name checks for now. Entropy checks
  may come later.
