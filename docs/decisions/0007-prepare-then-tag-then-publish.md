# 0007: Separate prepare, tag and publish

- Status: accepted
- Date: 2026-11-03

## Context

A release has three parts: change files (versions, changelog), mark the release in
git (commit and tag), and upload to registries. Doing all three in one command on a
laptop is what leads to stored registry tokens on developer machines.

## Decision

- `sealtag prepare` only edits files. It never publishes.
- `sealtag prepare --commit` commits exactly the files it changed and creates one
  annotated tag per package. It refuses to run on a dirty tree, so a release commit
  never carries unrelated work.
- Publishing (milestone 4) runs in CI when the tags are pushed, using trusted
  publishing. The laptop never holds a registry token.
- Edits are textual (`internal/textedit`) so comments and formatting in manifests
  survive. Re-encoding JSON or TOML would rewrite the whole file.

## Consequences

- The flow is: `plan` -> `prepare --commit` -> `git push --follow-tags` -> CI publishes.
- The release commit is `chore(release): ...`, which is not releasable, so the next
  `plan` correctly finds nothing new.
