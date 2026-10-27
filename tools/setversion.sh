#!/bin/sh
# setversion.sh VERSION BUILD - writes the release version (e.g. 1.0.4)
# and build number (Android's versionCode) into giode.toml.
set -eu
version=$1
build=$2
sed -i.bak \
  -e "s/^version = .*/version = \"$version\"/" \
  -e "s/^build = .*/build = $build/" \
  -e "s/settings\.Version=[^\" ]*/settings.Version=$version/" \
  giode.toml
rm -f giode.toml.bak
grep -E '^(version|build|ldflags) =' giode.toml
