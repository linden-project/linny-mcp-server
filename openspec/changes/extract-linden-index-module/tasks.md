# Tasks

## 1. Prepare the dependency

- [x] 1.1 Confirm `linden-index` has a tagged release containing `index` and
      `hugoref`.
- [x] 1.2 Add `github.com/linden-project/linden-index` to `go.mod` and run
      `go mod tidy`.

## 2. Rewire the server

- [x] 2.1 Update imports in `internal/mcp/server.go`, `tools.go`, `write.go`, and
      their tests from the internal package to `linden-index/index`.
- [x] 2.2 Update `internal/redact/e2e_test.go` and any other test that builds an
      index.
- [x] 2.3 Add a compile-time check or test that the server still emits and reads
      the same index files.

## 3. Delete the moved code

- [x] 3.1 Remove `cmd/lindexer`.
- [x] 3.2 Remove `internal/index` and `internal/hugoref`.
- [x] 3.3 Remove tests and fixtures that only served the moved packages.
- [ ] 3.4 Remove the specs owned by `linden-index` from `openspec/specs` when the
      change is archived.

## 4. Packaging and versioning

- [x] 4.1 Set `subPackages` in `nix/package.nix` to `cmd/linny-mcp` only.
- [x] 4.2 Drop `hugo` from the dev shell and drop the Hugo round-trip check.
- [x] 4.3 Keep the coverage gates and update the core package list.
- [x] 4.4 Confirm `VERSION` still stamps only `linny-mcp`.
- [x] 4.5 Update `RELEASING.md` to say `lindexer` is released from
      `linden-index`.

## 5. Docs

- [x] 5.1 Update `README.md`, `docs/future.md`, and any install instructions that
      mention `lindexer`.
- [x] 5.2 Note the `linden-index` dependency in `openspec/project.md`.
- [x] 5.3 Tell downstream `mipnix` consumers where `lindexer` now comes from.

## 6. Verify

- [x] 6.1 `go build ./...`
- [x] 6.2 `go test ./...` with the coverage gate
- [x] 6.3 `golangci-lint run`
- [x] 6.4 `nix flake check`

## Notes

- `linden-index` `v0.2.0` is tagged and pushed (commit `bea9f938`); pin that
  version when adding the dependency.
