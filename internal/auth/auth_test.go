package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func newAuth(t *testing.T) (*StaticTokenAuthenticator, string) {
	t.Helper()
	token, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	a := NewStaticTokenAuthenticator([]TokenRecord{
		{Name: "claude-mobile", Hash: HashToken(token), Scopes: []string{"read:*", "write:inbox"}},
	})
	return a, token
}

func TestAuthenticateValid(t *testing.T) {
	a, token := newAuth(t)
	id, err := a.Authenticate(token)
	if err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if id.Name != "claude-mobile" || len(id.Scopes) != 2 {
		t.Fatalf("unexpected identity %+v", id)
	}
}

func TestAuthenticateWrongAndEmpty(t *testing.T) {
	a, _ := newAuth(t)
	if _, err := a.Authenticate("definitely-not-the-token"); err != ErrUnauthorized {
		t.Fatalf("wrong token: got %v, want ErrUnauthorized", err)
	}
	if _, err := a.Authenticate(""); err != ErrUnauthorized {
		t.Fatalf("empty token: got %v, want ErrUnauthorized", err)
	}
}

func TestGenerateTokenStrength(t *testing.T) {
	tok, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(tok)
	if err != nil {
		t.Fatalf("token is not base64url: %v", err)
	}
	if len(raw) < 32 {
		t.Fatalf("token entropy = %d bytes, want >= 32", len(raw))
	}
}

func TestParseBearer(t *testing.T) {
	cases := map[string]struct {
		header  string
		wantTok string
		wantOK  bool
	}{
		"valid":            {"Bearer abc", "abc", true},
		"case-insensitive": {"bearer abc", "abc", true},
		"missing":          {"", "", false},
		"wrong scheme":     {"Basic abc", "", false},
		"empty token":      {"Bearer ", "", false},
	}
	for name, c := range cases {
		tok, ok := ParseBearer(c.header)
		if ok != c.wantOK || tok != c.wantTok {
			t.Errorf("%s: ParseBearer(%q) = (%q,%v), want (%q,%v)", name, c.header, tok, ok, c.wantTok, c.wantOK)
		}
	}
}

func TestTokenFileRoundTripAndNoRawToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens")
	content := "# tokens\n" +
		`{"name":"a","hash":"` + HashToken(token) + `","scopes":["read:*"]}` + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	// The raw token must never appear in the file.
	if b, _ := os.ReadFile(path); contains(string(b), token) {
		t.Fatal("token file must not contain the raw token")
	}
	recs, err := LoadTokenFile(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if _, err := NewStaticTokenAuthenticator(recs).Authenticate(token); err != nil {
		t.Fatalf("round-trip auth failed: %v", err)
	}
}

// TestMiddlewareNoInformationLeak asserts that a missing header and a wrong
// token produce byte-identical 401 responses with empty bodies.
func TestMiddlewareNoInformationLeak(t *testing.T) {
	a, _ := newAuth(t)
	h := Middleware(a, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	do := func(header string) (int, string, http.Header) {
		req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String(), rec.Result().Header
	}

	c1, b1, _ := do("")                   // missing
	c2, b2, _ := do("Bearer wrong-token") // wrong
	c3, b3, _ := do("Basic xyz")          // malformed scheme

	for i, got := range []int{c1, c2, c3} {
		if got != http.StatusUnauthorized {
			t.Fatalf("case %d: status = %d, want 401", i, got)
		}
	}
	if b1 != "" || b2 != "" || b3 != "" {
		t.Fatalf("401 bodies must be empty, got %q/%q/%q", b1, b2, b3)
	}
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
