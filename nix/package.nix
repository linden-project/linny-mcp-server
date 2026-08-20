{ lib, buildGoModule, version ? "0.0.0-dev" }:

buildGoModule {
  pname = "linny-mcp";
  inherit version;

  src = lib.cleanSourceWith {
    src = ../.;
    # Keep the source closure small and deterministic: only the Go build inputs.
    filter = path: type:
      let base = baseNameOf path; in
      !(lib.elem base [ ".git" ".jj" ".direnv" "result" ".beans" "openspec" "docs" "testdata-gen" ]);
  };

  # No external Go modules yet. The first change that adds a dependency MUST
  # replace this with the real vendor hash — a null hash fails loudly, which is
  # the intended fail-fast.
  vendorHash = null;

  subPackages = [ "cmd/linny-mcp" "cmd/lindexer" ];

  ldflags = [
    "-s"
    "-w"
    "-X github.com/linden-project/linny-mcp-server/internal/buildinfo.Version=${version}"
  ];

  # Run `go test ./...` as part of the package build so the package and the
  # `gotest` check share one derivation.
  doCheck = true;

  meta = {
    description = "MCP server exposing a Linny notebook (markdown + front matter) to AI agents";
    homepage = "https://github.com/linden-project/linny-mcp-server";
    license = lib.licenses.mit;
    mainProgram = "linny-mcp";
    platforms = [ "x86_64-linux" "aarch64-linux" ];
  };
}
