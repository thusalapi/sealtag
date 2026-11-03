# 0005: Build Python packages to learn what they contain

- Status: accepted
- Date: 2026-11-03

## Context

npm and cargo can report their file list without building anything. Python cannot.
The build backend (hatchling, setuptools, flit, pdm, maturin) decides what goes into
the sdist and the wheel, and each one has its own include rules. Guessing those rules
would be wrong often, and a wrong "PASS" is worse than no answer.

## Decision

`check` builds the sdist and the wheel into a temporary folder (with `uv build`, or
`python -m build` when uv is missing), reads both archives, and scans the bytes
inside them. Nothing is uploaded and the temporary folder is deleted.

## Consequences

- The result is exact: it is what PyPI would receive.
- Building runs the project's build backend, which is code. Unlike `npm pack`, there
  is no "ignore scripts" switch. This is acceptable because sealtag runs on a project
  the user is about to publish anyway, but it must never be run on untrusted code.
- It is slower (seconds, and the first build may download the backend).
- Scanning archive bytes rather than disk files also catches files that the build
  generates, such as `PKG-INFO`.
