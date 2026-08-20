package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/auth"
)

func TestGenTokenProducesMatchingRecord(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run([]string{"linny-mcp", "gen-token", "--name", "claude-mobile", "--scopes", "read:*,write:inbox"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("gen-token exit=%d stderr=%q", code, errOut.String())
	}

	var token string
	var recLine string
	for _, line := range strings.Split(out.String(), "\n") {
		switch {
		case strings.HasPrefix(line, "token: "):
			token = strings.TrimPrefix(line, "token: ")
		case strings.HasPrefix(line, "{"):
			recLine = line
		}
	}
	if token == "" || recLine == "" {
		t.Fatalf("expected a token line and a record line, got:\n%s", out.String())
	}

	var rec auth.TokenRecord
	if err := json.Unmarshal([]byte(recLine), &rec); err != nil {
		t.Fatalf("record line not JSON: %v", err)
	}
	if rec.Name != "claude-mobile" {
		t.Errorf("name = %q", rec.Name)
	}
	if rec.Hash != auth.HashToken(token) {
		t.Error("record hash does not match the printed token")
	}
	// The authenticator built from the record must accept the printed token.
	if _, err := auth.NewStaticTokenAuthenticator([]auth.TokenRecord{rec}).Authenticate(token); err != nil {
		t.Errorf("token rejected by its own record: %v", err)
	}
}

func TestGenTokenRequiresName(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run([]string{"linny-mcp", "gen-token"}, &out, &errOut); code == 0 {
		t.Fatal("gen-token without --name should fail")
	}
}
