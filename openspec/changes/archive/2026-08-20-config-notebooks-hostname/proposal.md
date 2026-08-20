## Why

Two open questions were resolved by the user: the public hostname must be
**configuration, never a hardcoded constant**, and **multi-notebook support should be
designed in now** (even if v1 serves one). Getting the config *shape* right early is
the point — it is annoying to retrofit N-notebook support later. This change lands the
config model and the NixOS options so nothing downstream has to be rewritten when a
second notebook (e.g. personal vs. business) is added.

## What Changes

- Add a real `Config` model in `internal/config`: a server section
  (`listenAddress`, `port`, `tokensFile`, `logLevel`, `readOnly`, and a configurable
  `publicHostname`) plus a list of `Notebook` entries, each with `name`, `corpusPath`,
  and `stateDir`. No hostname is hardcoded anywhere.
- Load config from a JSON file (`--config`), with validation (unique non-empty
  notebook names, at least one notebook, paths present). Keep the existing single
  notebook `serve` flags as sugar that construct a one-notebook config.
- `serve` resolves its notebook(s) from the config; a `--notebook` selector picks one
  when several are configured (v1 serves the selected/first; full per-notebook routing
  is a follow-up). Each notebook gets its own git-safety `Guard`.
- Update the NixOS module: add `publicHostname`, add a `notebooks` list option
  (name/corpusPath/stateDir), and keep `corpusPath`/`stateDir` as a convenience that
  desugars to a single default notebook. `tokensFile` stays a path (secret hygiene).

## Capabilities

### New Capabilities
- `configuration`: the server + N-notebook config model, its file loader and
  validation, and the configurable public hostname.

### Modified Capabilities

## Impact

- New: `internal/config` config model + loader (extends the package that currently
  holds bind safety).
- Modified: `cmd/linny-mcp` `serve` (config resolution, `--config`, `--notebook`),
  `nix/module.nix` (options: `publicHostname`, `notebooks`).
- No new external dependency (JSON via stdlib).
