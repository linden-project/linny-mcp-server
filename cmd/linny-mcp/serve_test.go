package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestServeRequiresTokensFile(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := serveCmd([]string{}, &out, &errOut); code != 2 {
		t.Fatalf("serve without --tokens-file/--config exit=%d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "tokens-file") {
		t.Fatalf("expected tokens-file hint, got %q", errOut.String())
	}
}

func TestServeRefusesPublicBind(t *testing.T) {
	var out, errOut bytes.Buffer
	// Bind refusal happens before any token load or listener bind.
	code := serveCmd([]string{"--tokens-file", "/nonexistent", "--listen", "8.8.8.8"}, &out, &errOut)
	if code != 1 {
		t.Fatalf("serve on public bind exit=%d, want 1", code)
	}
	if !strings.Contains(errOut.String(), "refusing to bind") {
		t.Fatalf("expected bind-refusal error, got %q", errOut.String())
	}
}
