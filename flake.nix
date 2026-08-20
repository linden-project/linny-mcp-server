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

          lint = pkgs.runCommand "linny-mcp-lint"
            {
              nativeBuildInputs = [ pkgs.go pkgs.golangci-lint ];
            }
            ''
              cp -r ${self.packages.${system}.default.src} src
              chmod -R u+w src
              cd src
              export HOME="$TMPDIR"
              export GOCACHE="$TMPDIR/go-cache"
              export GOLANGCI_LINT_CACHE="$TMPDIR/golangci-cache"
              export GOFLAGS="-mod=mod"
              golangci-lint run ./...
              touch "$out"
            '';
        });

      # Fleshed out in milestone 06 (nixos-module-hardened). Declared here so the
      # output exists from the first change; the hardened systemd unit lands later.
      nixosModules.linny-mcp = import ./nix/module.nix;

      formatter = forAllSystems (system: nixpkgsFor.${system}.nixpkgs-fmt);
    };
}
