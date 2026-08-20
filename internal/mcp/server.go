package mcp

import (
	"encoding/json"
	"net/http"

	"github.com/linden-project/linny-mcp-server/internal/auth"
)

// HealthStatus is the body of /healthz. Degraded and the sync fields are
// placeholders filled in by the git-safety change.
type HealthStatus struct {
	Status     string   `json:"status"`
	Degraded   bool     `json:"degraded"`
	Conflicted bool     `json:"conflicted"`
	Conflicts  []string `json:"conflicts,omitempty"`
	LastSync   string   `json:"last_sync,omitempty"`
	Ahead      int      `json:"ahead"`
	Behind     int      `json:"behind"`
}

// Server is the HTTP skeleton. The MCP protocol + tools mount on /mcp in the
// follow-up change; for now /mcp is an authenticated placeholder.
type Server struct {
	Auth   auth.Authenticator
	Health func() HealthStatus
}

// Handler returns the composed HTTP handler: /healthz is unauthenticated;
// everything else sits behind bearer auth.
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

	// Authenticated placeholder for the MCP endpoint.
	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, _ := auth.FromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "authenticated",
			"identity": id.Name,
			"scopes":   id.Scopes,
			"note":     "MCP tool surface is wired in a follow-up change",
		})
	})
	mux.Handle("/mcp", auth.Middleware(s.Auth, mcpHandler))

	return mux
}
