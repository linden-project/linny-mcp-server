// Command linny-mcp is the MCP server exposing a Linny notebook to AI agents.
//
// Implemented: `version`, `gen-token`, and a `serve` skeleton (bind safety,
// bearer auth, /healthz). The MCP tool surface mounts on /mcp in a follow-up.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/linden-project/linny-mcp-server/internal/auth"
	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
	"github.com/linden-project/linny-mcp-server/internal/config"
	"github.com/linden-project/linny-mcp-server/internal/gitsafe"
	"github.com/linden-project/linny-mcp-server/internal/mcp"
)

const usage = `linny-mcp - MCP server for a Linny notebook

Usage:
  linny-mcp <command> [flags]

Commands:
  serve       Run the MCP server (bind safety + bearer auth + /healthz)
  gen-token   Generate a bearer token and its token-file record
  version     Print the version and exit
`

// Run executes the CLI and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[1] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "linny-mcp %s\n", buildinfo.Version)
		return 0
	case "gen-token":
		return genTokenCmd(args[2:], stdout, stderr)
	case "serve":
		return serveCmd(args[2:], stdout, stderr)
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "linny-mcp: unknown command %q\n\n%s", args[1], usage)
		return 2
	}
}

func genTokenCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gen-token", flag.ContinueOnError)
	fs.SetOutput(stderr)
	name := fs.String("name", "", "human-readable token name (e.g. claude-mobile)")
	scopes := fs.String("scopes", "read:*", "comma-separated scopes")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *name == "" {
		fmt.Fprintln(stderr, "linny-mcp: gen-token requires --name")
		return 2
	}

	token, err := auth.GenerateToken()
	if err != nil {
		fmt.Fprintf(stderr, "linny-mcp: generating token: %v\n", err)
		return 1
	}
	rec := auth.TokenRecord{
		Name:   *name,
		Hash:   auth.HashToken(token),
		Scopes: splitScopes(*scopes),
	}
	line, _ := json.Marshal(rec)

	fmt.Fprintf(stdout, "token: %s\n", token)
	fmt.Fprintln(stdout, "# ^ copy this now; it is NOT stored. Append the record below to your tokens file:")
	fmt.Fprintf(stdout, "%s\n", line)
	return 0
}

func splitScopes(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func serveCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "", "path to a JSON config file (overrides the single-notebook flags below)")
	notebook := fs.String("notebook", "", "notebook to serve when several are configured (default: first)")
	corpus := fs.String("corpus", ".", "corpus root (single-notebook sugar; ignored with --config)")
	stateDir := fs.String("state-dir", "", "disposable index/state directory")
	listen := fs.String("listen", "127.0.0.1", "listen address (loopback/mesh only, unless override)")
	port := fs.Int("port", 8765, "listen port")
	tokensFile := fs.String("tokens-file", "", "path to the bearer token file")
	logLevel := fs.String("log-level", "info", "log level")
	readOnly := fs.Bool("read-only", false, "force read-only mode")
	override := fs.Bool("i-know-what-im-doing", false, "allow binding a public address")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	var cfg config.Config
	var err error
	if *configPath != "" {
		if cfg, err = config.Load(*configPath); err != nil {
			fmt.Fprintf(stderr, "linny-mcp: %v\n", err)
			return 1
		}
	} else {
		if *tokensFile == "" {
			fmt.Fprintln(stderr, "linny-mcp: serve requires --tokens-file (a PATH, never a token value) or --config")
			return 2
		}
		if cfg, err = config.FromFlags(*listen, *port, *tokensFile, *logLevel, *readOnly, *corpus, *stateDir); err != nil {
			fmt.Fprintf(stderr, "linny-mcp: %v\n", err)
			return 1
		}
	}

	nb, err := cfg.Resolve(*notebook)
	if err != nil {
		fmt.Fprintf(stderr, "linny-mcp: %v\n", err)
		return 1
	}
	if cfg.TokensFile == "" {
		fmt.Fprintln(stderr, "linny-mcp: config has no tokensFile (a PATH, never a token value)")
		return 2
	}
	if err := config.CheckBindAddress(cfg.ListenAddress, *override); err != nil {
		fmt.Fprintf(stderr, "linny-mcp: %v\n", err)
		return 1
	}
	records, err := auth.LoadTokenFile(cfg.TokensFile)
	if err != nil {
		fmt.Fprintf(stderr, "linny-mcp: loading tokens: %v\n", err)
		return 1
	}

	guard := gitsafe.NewGuard(nb.CorpusPath, cfg.ReadOnly)
	srv := &mcp.Server{
		Auth:   auth.NewStaticTokenAuthenticator(records),
		Health: healthFromGuard(guard),
	}
	addr := net.JoinHostPort(cfg.ListenAddress, fmt.Sprintf("%d", cfg.Port))
	if st := guard.State(); !st.Clean {
		fmt.Fprintf(stderr, "linny-mcp: WARNING starting in degraded read-only mode: %s\n", st.Reason)
	}
	hostNote := ""
	if cfg.PublicHostname != "" {
		hostNote = fmt.Sprintf(", public=%s", cfg.PublicHostname)
	}
	fmt.Fprintf(stdout, "linny-mcp %s serving notebook %q on %s (%d token(s), read-only=%t%s)\n",
		buildinfo.Version, nb.Name, addr, len(records), cfg.ReadOnly, hostNote)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil { //nolint:gosec // TLS terminates upstream
		fmt.Fprintf(stderr, "linny-mcp: server error: %v\n", err)
		return 1
	}
	return 0
}

// healthFromGuard adapts a git-safety guard to the health-status provider the
// HTTP server expects.
func healthFromGuard(g *gitsafe.Guard) func() mcp.HealthStatus {
	return func() mcp.HealthStatus {
		st := g.State()
		degraded := !st.Clean || g.ForcedReadOnly()
		status := "ok"
		if degraded {
			status = "degraded"
		}
		return mcp.HealthStatus{
			Status:     status,
			Degraded:   degraded,
			Conflicted: st.Conflicted,
			Conflicts:  st.ConflictedPaths,
		}
	}
}
