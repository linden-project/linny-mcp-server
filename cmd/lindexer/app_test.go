package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
)

func TestVersionSubcommand(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run([]string{"lindexer", "version"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr: %q)", code, errOut.String())
	}
	if got := out.String(); !strings.Contains(got, "lindexer") || !strings.Contains(got, buildinfo.Version) {
		t.Fatalf("version output = %q, want name and version %q", got, buildinfo.Version)
	}
}
