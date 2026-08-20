// Package buildinfo holds build-time metadata shared by all binaries.
//
// Version is overridden at build time via the linker:
//
//	-ldflags "-X github.com/linden-project/linny-mcp-server/internal/buildinfo.Version=1.2.3"
//
// The Nix package sets this from the flake's version attribute so that the
// `version` subcommand of every binary reports a consistent value.
package buildinfo

// Version is the build version. The default is used for `go run`/dev builds.
var Version = "0.0.0-dev"
