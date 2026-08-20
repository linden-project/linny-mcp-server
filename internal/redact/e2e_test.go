package redact_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

// TestIndexedSecretNoteScrubbed proves the end-to-end property: a note carrying
// fake credentials, once indexed and read back, has its secrets removed by the
// redactor before the bytes would leave the server.
func TestIndexedSecretNoteScrubbed(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 1, Count: 5, EnableEdgeCases: true}); err != nil {
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
	defer func() { _ = store.Close() }()
	if err := store.Populate(g); err != nil {
		t.Fatal(err)
	}

	doc, ok, err := store.GetDoc("fake_secrets.md")
	if err != nil || !ok {
		t.Fatalf("GetDoc fake_secrets.md ok=%v err=%v", ok, err)
	}

	r := redact.New()
	scrubbedBody, n := r.Redact(doc.Body)
	if n == 0 {
		t.Fatal("expected redactions in the fake-secrets note body")
	}

	// None of the planted secrets may survive.
	for _, secret := range []string{
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"ghp_0123456789abcdefghijklmnopqrstuvwxyzAB",
		"NL91ABNA0417164300",
		"BEGIN RSA PRIVATE KEY",
	} {
		if strings.Contains(scrubbedBody, secret) {
			t.Errorf("secret survived redaction: %q", secret)
		}
	}
}
