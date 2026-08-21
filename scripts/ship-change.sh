#!/usr/bin/env bash
# Ship one OpenSpec change end-to-end: stage -> gate -> archive -> commit -> push.
# Usage: scripts/ship-change.sh <change-name> "<commit subject>"
#
# The gate is `nix flake check` (build + tests + lint + coverage). It is never
# bypassed: if it fails, nothing is archived, committed, or pushed. Commits are
# authored by jj's configured user (Pim Snel) with no self-promoting trailers.
set -euo pipefail

change="${1:?usage: ship-change.sh <change-name> \"<commit subject>\"}"
subject="${2:?usage: ship-change.sh <change-name> \"<commit subject>\"}"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

nix_flags=(--extra-experimental-features 'nix-command flakes')

echo ">> staging working tree"
git add -A

echo ">> gate: nix flake check (build + tests + lint + coverage)"
if ! nix flake check "${nix_flags[@]}"; then
  echo "!! gate failed — not archiving, committing, or pushing." >&2
  exit 1
fi

echo ">> archiving OpenSpec change: $change"
openspec archive "$change" --yes

echo ">> committing"
git add -A
jj commit -m "$subject"
jj bookmark set main -r @-

echo ">> pushing main"
jj git push --bookmark main

echo ">> shipped: $change"
