#!/usr/bin/env bash
# Promote the CHANGELOG's [Unreleased] section to a dated version.
# Usage: scripts/release.sh <version>   (e.g. 0.2.0)
#
# Rewrites "## [Unreleased]" to "## [<version>] - <date>" and leaves a fresh empty
# [Unreleased] above it. Tagging is intentionally left as an explicit follow-up.
set -euo pipefail

version="${1:?usage: release.sh <version>}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

date="$(date +%Y-%m-%d)"

awk -v v="$version" -v d="$date" '
  !seen && /^## \[Unreleased\]/ {
    print "## [Unreleased]"
    print ""
    print "## [" v "] - " d
    seen = 1
    next
  }
  { print }
' CHANGELOG.md > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md

echo ">> promoted [Unreleased] -> [$version] - $date"
echo ">> review CHANGELOG.md, bump flake.nix version, then tag, e.g.: git tag v$version"
