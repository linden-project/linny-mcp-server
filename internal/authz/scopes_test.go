package authz

import (
	"strings"
	"testing"
)

func TestParseValidAndInvalid(t *testing.T) {
	ss, err := Parse([]string{"read:*", "deny:taxonomy:tags:health", "write:inbox", "read:taxonomy:customer", "delete:*"})
	if err != nil {
		t.Fatalf("valid scopes rejected: %v", err)
	}
	if len(ss.rules) != 5 {
		t.Fatalf("expected 5 rules, got %d", len(ss.rules))
	}

	for _, bad := range []string{"frobnicate:everything", "read:taxonomy:", "", "read:bogus:x"} {
		if _, err := Parse([]string{bad}); err == nil {
			t.Errorf("expected error for scope %q", bad)
		}
	}
}

func TestDenyByDefault(t *testing.T) {
	ss, err := Parse([]string{"write:inbox"}) // no read rule
	if err != nil {
		t.Fatal(err)
	}
	sql, args := ss.ReadableFilenamesSQL()
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	// allowExpr must be the constant-false "0" so nothing is selected.
	if !strings.Contains(sql, "WHERE 0 AND NOT 0") {
		t.Fatalf("deny-by-default SQL should select nothing, got: %s", sql)
	}
}

func TestReadAllSQL(t *testing.T) {
	ss, _ := Parse([]string{"read:*"})
	sql, args := ss.ReadableFilenamesSQL()
	if len(args) != 0 || !strings.Contains(sql, "(1=1) AND NOT 0") {
		t.Fatalf("read:* SQL = %q args=%v", sql, args)
	}
}

func TestArgOrderMatchesPlaceholders(t *testing.T) {
	// allow args (a, c) must precede deny args (b), matching render order
	// (allow conditions first, then deny conditions).
	ss, _ := Parse([]string{"read:taxonomy:a", "deny:taxonomy:b", "read:taxonomy:c"})
	_, args := ss.ReadableFilenamesSQL()
	want := []any{"a", "c", "b"}
	if len(args) != len(want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d] = %v, want %v (full: %v)", i, args[i], want[i], args)
		}
	}
}

func TestCanReadAny(t *testing.T) {
	cases := []struct {
		name   string
		scopes []string
		want   bool
	}{
		{"read all", []string{"read:*"}, true},
		{"read one taxonomy", []string{"read:taxonomy:work"}, true},
		{"read one term", []string{"read:taxonomy:work:acme"}, true},
		{"no scopes at all", nil, false},
		{"write only", []string{"write:*"}, false},
		{"write inbox only", []string{"write:inbox"}, false},
		{"deny only", []string{"deny:taxonomy:health"}, false},
		{"read plus deny still reads", []string{"read:*", "deny:taxonomy:health"}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ss, err := Parse(c.scopes)
			if err != nil {
				t.Fatal(err)
			}
			if got := ss.CanReadAny(); got != c.want {
				t.Fatalf("CanReadAny() = %v, want %v", got, c.want)
			}
		})
	}
}

// A deny rule is parsed as a read action; it must never be mistaken for a grant.
func TestDenyIsNotAReadGrant(t *testing.T) {
	ss, err := Parse([]string{"deny:taxonomy:health"})
	if err != nil {
		t.Fatal(err)
	}
	if ss.CanReadAny() {
		t.Fatal("a deny rule alone grants no read")
	}
	sql, _ := ss.ReadableFilenamesSQL()
	if !strings.Contains(sql, "WHERE 0 AND NOT") {
		t.Fatalf("deny-only should select nothing, got %q", sql)
	}
}
