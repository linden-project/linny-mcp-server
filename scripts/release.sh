#!/usr/bin/env bash
# Cut a release: preflight gate -> gum bump chooser -> bump VERSION + roll the
# CHANGELOG -> jj commit/push -> git tag (via the colocated git) -> gh release.
#
# Usage: scripts/release.sh            (interactive: gum choose major/minor/patch)
#        scripts/release.sh patch      (non-interactive bump level)
#
# A release is just a git tag: mipnix consumes this repo as a flake input built
# from source, so there is nothing to compile-and-upload. Requires local `gh`
# auth for the GitHub release. Never bypasses the preflight gate.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

nix_flags=(--extra-experimental-features 'nix-command flakes')

# --- preflight: clean working copy on an up-to-date main, and a green gate ---
echo ">> preflight: working copy must be clean"
if ! jj status 2>/dev/null | grep -q "The working copy has no changes"; then
  echo "!! working copy is not clean — commit or abandon changes first." >&2
  exit 1
fi

echo ">> preflight: main must be in sync with origin"
jj git fetch >/dev/null 2>&1 || true
local_main="$(git rev-parse main 2>/dev/null || true)"
remote_main="$(git rev-parse origin/main 2>/dev/null || true)"
if [ -n "$remote_main" ] && [ "$local_main" != "$remote_main" ]; then
  echo "!! local main ($local_main) != origin/main ($remote_main) — sync first." >&2
  exit 1
fi

echo ">> preflight: nix flake check (build + tests + lint + coverage)"
nix flake check "${nix_flags[@]}"

# --- choose the bump level and compute the next version from VERSION ---
cur="$(cat VERSION)"
if [[ ! "$cur" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "!! VERSION ('$cur') is not X.Y.Z — release math needs a plain semver." >&2
  exit 1
fi

level="${1:-}"
if [ -z "$level" ]; then
  level="$(gum choose major minor patch)"
fi

IFS=. read -r major minor patch <<<"$cur"
case "$level" in
  major) next="$((major + 1)).0.0" ;;
  minor) next="${major}.$((minor + 1)).0" ;;
  patch) next="${major}.${minor}.$((patch + 1))" ;;
  *) echo "!! bump level must be major|minor|patch (got '$level')" >&2; exit 1 ;;
esac

echo ">> $cur -> $next ($level)"
if ! gum confirm "Cut release v$next?"; then
  echo ">> aborted."
  exit 1
fi

# --- bump VERSION and roll the CHANGELOG's [Unreleased] into a dated section ---
printf '%s\n' "$next" > VERSION
date="$(date +%Y-%m-%d)"
awk -v v="$next" -v d="$date" '
  !seen && /^## \[Unreleased\]/ {
    print "## [Unreleased]"; print ""; print "## [" v "] - " d
    seen = 1; next
  }
  { print }
' CHANGELOG.md > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md

# --- commit + push with jj, then tag the bump commit via the colocated git ---
echo ">> committing the version bump"
git add -A
jj commit -m "release: v$next"
jj bookmark set main -r @-
jj git push --bookmark main

rev="$(git rev-parse main)"
echo ">> tagging v$next at $rev (via git; jj git push does not push tags)"
git tag "v$next" "$rev"
git push origin "v$next"

# --- GitHub release with the new CHANGELOG section as the notes ---
notes="$(awk -v v="$next" '
  $0 ~ "^## \\[" v "\\]" { grab = 1; next }
  grab && /^## \[/ { exit }
  grab { print }
' CHANGELOG.md)"
echo ">> creating GitHub release v$next"
printf '%s\n' "$notes" | gh release create "v$next" --title "v$next" --notes-file -

echo ">> released v$next"
