{
  description = "linny-mcp — MCP server exposing a Linny notebook to AI agents over MCP";

  # Only nixpkgs. No flake-utils (deliberate): systems are enumerated explicitly
  # and mapped with a small local helper.
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      version = "0.1.0-alpha";

      # Supported systems, enumerated explicitly (plain nix, no flake-utils).
      systems = [ "x86_64-linux" "aarch64-linux" ];

      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);

      nixpkgsFor = forAllSystems (system: import nixpkgs {
        inherit system;
        overlays = [ self.overlays.default ];
      });
    in
    {
      overlays.default = final: _prev: {
        linny-mcp = final.callPackage ./nix/package.nix { inherit version; };
      };

      packages = forAllSystems (system:
        let pkgs = nixpkgsFor.${system}; in {
          linny-mcp = pkgs.linny-mcp;
          default = pkgs.linny-mcp;
        });

      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system}; in {
          default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.gopls
              pkgs.golangci-lint
              pkgs.hugo # needed by the indexer `verify` path (diff vs Hugo)
            ];
          };
        });

      checks = forAllSystems (system:
        let pkgs = nixpkgsFor.${system}; in {
          # `gotest` reuses the package derivation, which runs `go test ./...`
          # in its checkPhase (doCheck = true in nix/package.nix).
          gotest = pkgs.linny-mcp;

          # `lint` overrides the package derivation so golangci-lint runs with the
          # vendored module set already present (offline, no proxy fetch).
          lint = pkgs.linny-mcp.overrideAttrs (old: {
            pname = "linny-mcp-lint";
            nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ pkgs.golangci-lint ];
            doCheck = false;
            buildPhase = ''
              runHook preBuild
              export GOLANGCI_LINT_CACHE="$TMPDIR/golangci-cache"
              golangci-lint run ./...
              runHook postBuild
            '';
            installPhase = ''
              mkdir -p "$out"
              touch "$out/lint-ok"
            '';
          });

          # `coverage` runs the whole suite with cross-package coverage and fails
          # below the 70% floor. git is on PATH so the git-backed tests run.
          coverage = pkgs.linny-mcp.overrideAttrs (old: {
            pname = "linny-mcp-coverage";
            nativeBuildInputs = (old.nativeBuildInputs or [ ]) ++ [ pkgs.git ];
            doCheck = false;
            buildPhase = ''
              runHook preBuild
              export HOME="$TMPDIR"
              go test ./... -coverpkg=./... -covermode=atomic -coverprofile=cover.out
              total=$(go tool cover -func=cover.out | awk '/^total:/ { gsub("%","",$3); print $3 }')
              echo "total statement coverage: $total% (floor 70%)"
              awk -v t="$total" 'BEGIN { if (t+0 < 70.0) { printf "FAIL: coverage %s%% is below the 70%% gate\n", t; exit 1 } }'
              runHook postBuild
            '';
            installPhase = ''
              mkdir -p "$out"
              cp cover.out "$out/cover.out"
            '';
          });
        });

      # Fleshed out in milestone 06 (nixos-module-hardened). Declared here so the
      # output exists from the first change; the hardened systemd unit lands later.
      nixosModules.linny-mcp = import ./nix/module.nix;

      formatter = forAllSystems (system: nixpkgsFor.${system}.nixpkgs-fmt);
    };
}
