package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/auth"
)

func testServer(t *testing.T) (*Server, string) {
	t.Helper()
	tok, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	a := auth.NewStaticTokenAuthenticator([]auth.TokenRecord{
		{Name: "tester", Hash: auth.HashToken(tok), Scopes: []string{"read:*"}},
	})
	return &Server{Auth: a}, tok
}

func TestHealthzUnauthenticated(t *testing.T) {
	s, _ := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}
	var hs HealthStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &hs); err != nil {
		t.Fatalf("healthz body not JSON: %v", err)
	}
	if hs.Status == "" {
		t.Fatal("healthz missing status field")
	}
}

func TestMCPRequiresAuth(t *testing.T) {
	s, tok := testServer(t)
	h := s.Handler()

	// Without a token → 401.
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("/mcp without token = %d, want 401", rec.Code)
	}

	// With a valid token → 200 and identity echoed.
	req = httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("/mcp with token = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not JSON: %v", err)
	}
	if body["identity"] != "tester" {
		t.Fatalf("identity = %v, want tester", body["identity"])
	}
}
