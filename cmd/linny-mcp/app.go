// Command linny-mcp is the MCP server exposing a Linny notebook to AI agents.
//
// At scaffold stage it only implements the `version` subcommand and stubs for
// the subcommands later epics fill in (`serve`, `gen-token`).
package main

import (
	"fmt"
	"io"

	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
)

const usage = `linny-mcp - MCP server for a Linny notebook

Usage:
  linny-mcp <command>

Commands:
  serve       Run the MCP server (not yet implemented)
  gen-token   Generate a bearer token (not yet implemented)
  version     Print the version and exit
`

// Run executes the CLI and returns a process exit code. It is separated from
// main so it can be unit-tested without spawning a process.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[1] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "linny-mcp %s\n", buildinfo.Version)
		return 0
	case "serve":
		fmt.Fprintln(stderr, "linny-mcp: serve is not implemented yet")
		return 1
	case "gen-token":
		fmt.Fprintln(stderr, "linny-mcp: gen-token is not implemented yet")
		return 1
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "linny-mcp: unknown command %q\n\n%s", args[1], usage)
		return 2
	}
}
