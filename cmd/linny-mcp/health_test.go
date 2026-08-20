package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/gitsafe"
	"github.com/linden-project/linny-mcp-server/internal/mcp"
)

func TestHealthzReflectsConflict(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A tracked file with committed conflict markers.
	if err := os.WriteFile(filepath.Join(root, "content", "bad.md"),
		[]byte("<<<<<<< HEAD\nmine\n>>>>>>> other\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	guard := gitsafe.NewGuard(root, false)
	srv := &mcp.Server{Health: healthFromGuard(guard)}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}
	var hs mcp.HealthStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &hs); err != nil {
		t.Fatal(err)
	}
	if !hs.Degraded || !hs.Conflicted || hs.Status != "degraded" {
		t.Fatalf("expected degraded+conflicted health, got %+v", hs)
	}
	if len(hs.Conflicts) != 1 {
		t.Fatalf("expected one conflicted path, got %v", hs.Conflicts)
	}
}
