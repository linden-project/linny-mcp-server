package redact

import (
	"strings"
	"testing"
)

func TestDetectorsRemoveSecrets(t *testing.T) {
	r := New()
	cases := []struct {
		name     string
		in       string
		mustGone string // substring that must NOT survive
		kind     string // placeholder kind expected
	}{
		{"aws-access-key", "id AKIAIOSFODNN7EXAMPLE here", "AKIAIOSFODNN7EXAMPLE", "aws-access-key"},
		{"github", "token: ghp_0123456789abcdefghijklmnopqrstuvwxyzAB", "ghp_0123456789abcdefghijklmnopqrstuvwxyzAB", "github-token"},
		{"jwt", "t=eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N", "dozjgNryP4J3jVmNHl0w5N", "jwt"},
		{"iban", "acct NL91ABNA0417164300 done", "NL91ABNA0417164300", "iban"},
		{"aws-secret", `aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "aws-secret-key"},
	}
	for _, c := range cases {
		got, n := r.Redact(c.in)
		if strings.Contains(got, c.mustGone) {
			t.Errorf("%s: secret survived: %q", c.name, got)
		}
		if !strings.Contains(got, "[REDACTED:"+c.kind+"]") {
			t.Errorf("%s: expected placeholder %s, got %q", c.name, c.kind, got)
		}
		if n < 1 {
			t.Errorf("%s: expected >=1 redaction, got %d", c.name, n)
		}
	}
}

func TestPEMPrivateKeyRemoved(t *testing.T) {
	in := "before\n-----BEGIN RSA PRIVATE KEY-----\nMIIBOgIBAAJBAKj34GkxFhD90vcNLYLInEXAMPLE\n-----END RSA PRIVATE KEY-----\nafter\n"
	got, n := New().Redact(in)
	if strings.Contains(got, "MIIBOgIBAAJBAKj34") || strings.Contains(got, "PRIVATE KEY") {
		t.Fatalf("private key material survived: %q", got)
	}
	if n != 1 || !strings.Contains(got, "[REDACTED:private-key]") {
		t.Fatalf("expected one private-key redaction, got n=%d %q", n, got)
	}
	if !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Fatalf("surrounding text lost: %q", got)
	}
}

func TestAssignmentKeepsKeyRedactsValue(t *testing.T) {
	got, _ := New().Redact("password: hunter2secret")
	if !strings.HasPrefix(got, "password:") {
		t.Fatalf("key should be kept: %q", got)
	}
	if strings.Contains(got, "hunter2secret") {
		t.Fatalf("value should be redacted: %q", got)
	}
}

func TestRedactValueDeepWalk(t *testing.T) {
	r := New()
	v := map[string]any{
		"title": "Innocent Note",
		"notes": []any{"nothing here", "key AKIAIOSFODNN7EXAMPLE leaked"},
		"nested": map[string]any{
			"token": "ghp_0123456789abcdefghijklmnopqrstuvwxyzAB",
		},
	}
	out := r.RedactValue(v).(map[string]any)
	if out["title"] != "Innocent Note" {
		t.Errorf("innocent title changed: %v", out["title"])
	}
	notes := out["notes"].([]any)
	if strings.Contains(notes[1].(string), "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("nested slice secret survived: %v", notes)
	}
	nested := out["nested"].(map[string]any)
	if strings.Contains(nested["token"].(string), "ghp_") {
		t.Errorf("nested map secret survived: %v", nested)
	}
}

func TestNoFalsePositiveOnProse(t *testing.T) {
	in := "The quick brown fox wrote notes about the project and its taxonomy today."
	got, n := New().Redact(in)
	if got != in || n != 0 {
		t.Fatalf("ordinary prose should be untouched, got n=%d %q", n, got)
	}
}

func TestCountReported(t *testing.T) {
	in := "AKIAIOSFODNN7EXAMPLE and token: ghp_0123456789abcdefghijklmnopqrstuvwxyzAB"
	got, n := New().Redact(in)
	if n < 2 {
		t.Fatalf("expected >=2 redactions, got %d", n)
	}
	for _, secret := range []string{"AKIAIOSFODNN7EXAMPLE", "ghp_0123456789"} {
		if strings.Contains(got, secret) {
			t.Fatalf("secret %q survived: %q", secret, got)
		}
	}
}
