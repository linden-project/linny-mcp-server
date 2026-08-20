package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
)

func TestVersionSubcommand(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run([]string{"linny-mcp", "version"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, errOut.String())
	}
	got := out.String()
	if !strings.Contains(got, "linny-mcp") || !strings.Contains(got, buildinfo.Version) {
		t.Fatalf("version output = %q, want name and version %q", got, buildinfo.Version)
	}
}

func TestNoArgsIsUsageError(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run([]string{"linny-mcp"}, &out, &errOut); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "Usage:") {
		t.Fatalf("expected usage on stderr, got %q", errOut.String())
	}
}
