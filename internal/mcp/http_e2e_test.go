package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/linden-project/linny-mcp-server/internal/auth"
	"github.com/linden-project/linny-mcp-server/internal/corpus"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

// bearerRT injects a bearer token on every request.
type bearerRT struct {
	base  http.RoundTripper
	token string
}

func (b bearerRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	if b.token != "" {
		r.Header.Set("Authorization", "Bearer "+b.token)
	}
	return b.base.RoundTrip(r)
}

// newServerWithStore builds a populated store + Server and a token with the
// given scopes.
func newServerWithStore(t *testing.T, scopes ...string) (*Server, string) {
	t.Helper()
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 5, Count: 40, EnableEdgeCases: true}); err != nil {
		t.Fatal(err)
	}
	g, _, err := index.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	store, err := index.OpenStore(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Populate(g); err != nil {
		t.Fatal(err)
	}

	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	a := auth.NewStaticTokenAuthenticator([]auth.TokenRecord{
		{Name: "e2e", Hash: auth.HashToken(token), Scopes: scopes},
	})
	return &Server{Auth: a, Store: store, Redactor: redact.New()}, token
}

func TestMCPEndToEndSearch(t *testing.T) {
	srv, token := newServerWithStore(t, "read:*", "deny:taxonomy:tags:health")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	ctx := context.Background()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "0"}, nil)
	transport := &mcpsdk.StreamableClientTransport{
		Endpoint:   ts.URL + "/mcp",
		HTTPClient: &http.Client{Transport: bearerRT{http.DefaultTransport, token}},
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	// Tools are listed.
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	names := map[string]bool{}
	for _, tl := range tools.Tools {
		names[tl.Name] = true
	}
	for _, want := range []string{"search", "get_doc", "list_taxonomies", "terms", "docs_by_term"} {
		if !names[want] {
			t.Errorf("tool %q not registered (have %v)", want, names)
		}
	}

	// get_doc on the health-denied doc reports not-found (scope enforced end-to-end).
	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "get_doc",
		Arguments: map[string]any{"slug": "work_and_health.md"},
	})
	if err != nil {
		t.Fatalf("call get_doc: %v", err)
	}
	var doc getDocOut
	decodeStructured(t, res, &doc)
	if doc.Found {
		t.Fatal("work_and_health.md must be not-found when health is denied")
	}

	// get_doc on the fake-secrets note returns a redacted body.
	res, err = session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "get_doc",
		Arguments: map[string]any{"slug": "fake_secrets.md"},
	})
	if err != nil {
		t.Fatalf("call get_doc secrets: %v", err)
	}
	var secretDoc getDocOut
	decodeStructured(t, res, &secretDoc)
	if !secretDoc.Found {
		t.Fatal("fake_secrets.md should be readable")
	}
	if strings.Contains(secretDoc.Body, "AKIAIOSFODNN7EXAMPLE") || strings.Contains(secretDoc.Body, "BEGIN RSA PRIVATE KEY") {
		t.Fatalf("secret leaked over MCP: %q", secretDoc.Body)
	}
}

func TestMCPUnauthenticatedRejected(t *testing.T) {
	srv, _ := newServerWithStore(t, "read:*")
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// No Authorization header -> 401 before any MCP processing.
	resp, err := http.Post(ts.URL+"/mcp", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauth /mcp = %d, want 401", resp.StatusCode)
	}
}

// decodeStructured pulls the structured tool output into v.
func decodeStructured(t *testing.T, res *mcpsdk.CallToolResult, v any) {
	t.Helper()
	if res.IsError {
		t.Fatalf("tool returned error: %+v", res.Content)
	}
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("decode structured content: %v", err)
	}
}

// TestMCPEndToEndCreateDoc drives a write over the real MCP protocol: create a
// doc, then read it back, confirming reindex + quarantine + delimited body.
func TestMCPEndToEndCreateDoc(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:inbox") // skips if git is unavailable
	// Attach auth to the fixture's populated store/guard/audit server.
	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	f.server.Auth = auth.NewStaticTokenAuthenticator([]auth.TokenRecord{
		{Name: "e2e", Hash: auth.HashToken(token), Scopes: []string{"read:*", "write:inbox"}},
	})

	ts := httptest.NewServer(f.server.Handler())
	defer ts.Close()

	ctx := context.Background()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "0"}, nil)
	transport := &mcpsdk.StreamableClientTransport{
		Endpoint:   ts.URL + "/mcp",
		HTTPClient: &http.Client{Transport: bearerRT{http.DefaultTransport, token}},
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "create_doc",
		Arguments: map[string]any{"title": "Agent Idea", "body": "a drafted thought"},
	})
	if err != nil {
		t.Fatalf("create_doc: %v", err)
	}
	var created writeOut
	decodeStructured(t, res, &created)
	if !created.OK || !created.Quarantined || created.Slug != "agent_idea.md" {
		t.Fatalf("unexpected create result: %+v", created)
	}

	// Read it back over MCP.
	res, err = session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "get_doc",
		Arguments: map[string]any{"slug": "agent_idea.md"},
	})
	if err != nil {
		t.Fatalf("get_doc: %v", err)
	}
	var got getDocOut
	decodeStructured(t, res, &got)
	if !got.Found || !strings.Contains(got.Body, "a drafted thought") {
		t.Fatalf("created doc not readable back: %+v", got)
	}
}
