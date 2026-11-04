# sealtag

**Publish a package to npm, PyPI, crates.io and Go from one config, with no stored tokens.**

The name: sealtag *seals* each release (checks it, signs it) before it *tags* and publishes it.

> Status: early development (milestone 1). Not ready for real releases yet.

## The problem

Publishing a library is a different ritual on every registry. Each one has its own
version file, its own build step and its own login. Most projects solve the login
part by saving a long-lived API token as a CI secret. Those tokens are what attackers
steal. In 2025 the "Shai-Hulud" worm phished npm maintainers, harvested tokens from
build machines and used them to push malware into hundreds of packages.

Registries now support **trusted publishing**: CI proves who it is with a short-lived
OIDC token, so nothing long-lived is stored. npm, PyPI and crates.io all support it,
but each one is configured differently.

## What sealtag does

1. **Detects** every publishable package in a repository (`package.json`,
   `pyproject.toml`, `Cargo.toml`, `go.mod`).
2. **Plans** the release: reads Conventional Commits since each package's last tag
   and picks the next version (fix = patch, feat = minor, breaking = major).
3. **Checks** what will ship: lists the files and scans them for secrets before
   anything leaves your machine.
4. **Verifies** the built package: installs it into an empty environment and loads
   it, so a missing file breaks the release instead of breaking your users.
5. **Publishes** from CI with trusted publishing only. It refuses to run if a
   long-lived token is present.

## What it catches

Real examples from sealtag's own tests:

| Problem | What happens without sealtag | sealtag |
|---|---|---|
| A Python project with no `.gitignore` has a local `.env` | hatchling puts `.env` into the sdist and it is uploaded to PyPI | `check` builds the sdist, scans inside it, and fails |
| npm `"files"` lists `index.js` but forgets `lib/` | Tests pass in the repo, but every install fails with "Cannot find module" | `verify` installs the tarball somewhere empty and fails |
| A Go module needs a gitignored generated file | It builds locally, but `go get` users get a compile error | `verify` builds the real module zip and fails |
| `NPM_TOKEN` is still stored in CI secrets | A stolen token can publish malware | `publish` refuses to run while it is set |

## Use it in GitHub Actions

```yaml
on:
  push:
    tags: ["v*", "*@v*", "*/v*"]

jobs:
  publish:
    runs-on: ubuntu-latest
    environment: release
    permissions:
      contents: read
      id-token: write   # trusted publishing
    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
      - uses: thusalapi/sealtag@v0
        with:
          command: publish
```

Then release from your laptop:

```bash
sealtag plan                 # what would be released
sealtag prepare --commit     # bump versions, write CHANGELOG.md, commit, tag
git push --follow-tags       # CI checks, verifies and publishes
```

## Quick start

```bash
go install github.com/thusalapi/sealtag/cmd/sealtag@latest

sealtag detect               # list publishable packages
sealtag plan                 # next version of each package, from Conventional Commits
sealtag changelog            # release notes for the pending release
sealtag prepare --commit     # write versions + CHANGELOG.md, commit, tag
sealtag verify               # install each built package somewhere empty and load it
sealtag doctor               # how to set up trusted publishing per registry
sealtag publish              # in CI: upload tagged packages with OIDC
sealtag scan                 # look for secrets in the repository
sealtag check --files        # exact files each package would publish, scanned
```

## Documentation

| Doc | Purpose |
|---|---|
| [docs/ROADMAP.md](docs/ROADMAP.md) | Milestones and what is done |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | How the code is organised |
| [docs/decisions/](docs/decisions/) | Architecture decision records (why things are the way they are) |
| [docs/DEVLOG.md](docs/DEVLOG.md) | Dated log of work sessions |
| [docs/LEARNING.md](docs/LEARNING.md) | Concepts learned while building this |

## Development

Needs Go 1.26 or newer. npm and cargo are optional and only used for the real-tool
tests and the `check` command.

```bash
make build   # builds ./bin/sealtag
make test    # runs all tests
make lint    # go vet
```

## License

MIT
