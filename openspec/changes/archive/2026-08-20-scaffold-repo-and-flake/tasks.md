## 1. Go module & entrypoints

- [x] 1.1 Create `go.mod` (`github.com/linden-project/linny-mcp-server`, go 1.26)
- [x] 1.2 Add `internal/buildinfo` with a `Version` var stamped via `-ldflags`
- [x] 1.3 Create `cmd/linny-mcp/main.go` stub with a `version` subcommand
- [x] 1.4 Create `cmd/lindexer/main.go` stub with a `version` subcommand
- [x] 1.5 Add a smoke test asserting `version` output for both binaries

## 2. Internal package layout

- [x] 2.1 Add `doc.go` stubs for `internal/{index,auth,authz,gitsafe,redact,mcp,config}`
- [x] 2.2 `go vet ./...` clean

## 3. Nix flake (plain nix, no flake-utils)

- [x] 3.1 `flake.nix`: `nixpkgs` input only; `systems` list + `forAllSystems` helper
- [x] 3.2 `nix/package.nix`: `buildGoModule` with `vendorHash = null`, version ldflags
- [x] 3.3 `packages.<system>.{linny-mcp,default}`
- [x] 3.4 `overlays.default` adding `linny-mcp`
- [x] 3.5 `devShells.default`: go, golangci-lint, hugo
- [x] 3.6 `checks.<system>.{gotest,lint}`
- [x] 3.7 `nixosModules.linny-mcp` stub (options declared, no implementation yet)
- [x] 3.8 `.golangci.yml` minimal linter set

## 4. Docs & hygiene

- [x] 4.1 Rewrite `README.md`: purpose + prominent secret-hygiene rule (§5.4)
- [x] 4.2 Extend `.gitignore`: `result`, `result-*`, `.direnv/`, `.linny_temp/`, state dir
- [x] 4.3 Add `docs/` placeholder note pointing to the index spec (written next epic)

## 5. Verification

- [x] 5.1 `go build ./...` and `go test ./...` green
- [x] 5.2 `go vet ./...` clean
- [x] 5.3 `nix flake check` attempted; record result (network-permitting) in the bean
