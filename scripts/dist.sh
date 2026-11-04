#!/usr/bin/env bash
# Builds sealtag release archives for every supported platform into dist/.
# Usage: scripts/dist.sh v0.2.0
#
# Builds are reproducible: -trimpath removes local paths, and the version and
# commit are the only values injected. The release workflow attests every file
# in dist/ with signed SLSA build provenance.
set -euo pipefail

version="${1:?usage: scripts/dist.sh <version>}"
commit="$(git rev-parse --short=7 HEAD)"
# Every timestamp in the archives is the commit time, so rebuilding the same
# commit gives byte-identical files (https://reproducible-builds.org).
epoch="$(git log -1 --format=%ct)"
pkg="github.com/thusalapi/sealtag/internal/cli"
ldflags="-s -w -X ${pkg}.version=${version} -X ${pkg}.commit=${commit}"
targets="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

rm -rf dist
mkdir -p dist/build
for t in $targets; do
  os="${t%/*}"
  arch="${t#*/}"
  name="sealtag_${version#v}_${os}_${arch}"
  dir="dist/build/${name}"
  mkdir -p "$dir"
  bin="sealtag"
  [ "$os" = "windows" ] && bin="sealtag.exe"
  echo "building ${name}"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags "$ldflags" -o "${dir}/${bin}" ./cmd/sealtag
  cp LICENSE README.md "$dir/"
  find "$dir" -exec touch -h -d "@${epoch}" {} +
  if [ "$os" = "windows" ]; then
    # -X drops extra file attributes (uid, extended timestamps).
    (cd dist/build && find "$name" | LC_ALL=C sort | zip -qX "../${name}.zip" -@)
  else
    # Sorted names, fixed owner and gzip -n (no name or time in the header).
    tar --sort=name --owner=0 --group=0 --numeric-owner -C dist/build -cf - "$name" \
      | gzip -n -9 > "dist/${name}.tar.gz"
  fi
done
rm -rf dist/build
(cd dist && sha256sum * > checksums.txt)
ls -l dist
