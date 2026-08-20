package mcp

import (
	"encoding/json"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/linden-project/linny-mcp-server/internal/auth"
	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

// HealthStatus is the body of /healthz.
type HealthStatus struct {
	Status     string   `json:"status"`
	Degraded   bool     `json:"degraded"`
	Conflicted bool     `json:"conflicted"`
	Conflicts  []string `json:"conflicts,omitempty"`
	LastSync   string   `json:"last_sync,omitempty"`
	Ahead      int      `json:"ahead"`
	Behind     int      `json:"behind"`
}

// Server is the HTTP surface: an unauthenticated /healthz and the authenticated
// MCP endpoint /mcp. When Store is set, /mcp serves the MCP read tools; when it
// is nil, /mcp is an authenticated placeholder (used by health-only setups).
type Server struct {
	Auth       auth.Authenticator
	Health     func() HealthStatus
	Store      *index.Store
	Redactor   *redact.Redactor
	CorpusPath string // notebook working tree, for the git history tools
}

// Handler returns the composed HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		hs := HealthStatus{Status: "ok"}
		if s.Health != nil {
			hs = s.Health()
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(hs)
	})

	mux.Handle("/mcp", auth.Middleware(s.Auth, s.mcpHandler()))
	return mux
}

// mcpHandler serves the MCP read tools when a store is attached, otherwise an
// authenticated placeholder.
func (s *Server) mcpHandler() http.Handler {
	if s.Store == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, _ := auth.FromContext(r.Context())
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":   "authenticated",
				"identity": id.Name,
				"scopes":   id.Scopes,
				"note":     "no notebook store attached; read tools unavailable",
			})
		})
	}

	red := s.Redactor
	if red == nil {
		red = redact.New()
	}
	// A fresh MCP server is built per request, bound to the caller's scopes read
	// from the authenticated request context.
	getServer := func(r *http.Request) *mcpsdk.Server {
		id, ok := auth.FromContext(r.Context())
		if !ok {
			return nil // never reached: middleware guarantees identity
		}
		ss, err := authz.Parse(id.Scopes)
		if err != nil {
			return nil // invalid scopes -> 400 from the transport
		}
		return buildToolServer(newReader(s.Store, red, ss, s.CorpusPath))
	}
	return mcpsdk.NewStreamableHTTPHandler(getServer, nil)
}
