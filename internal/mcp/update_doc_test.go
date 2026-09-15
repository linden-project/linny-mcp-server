package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/linden-project/linny-mcp-server/internal/auth"
	"github.com/linden-project/linny-mcp-server/internal/defense"
)

const updDoc = `---
title: Update Me
# a comment
customer: eric
---

first paragraph
second paragraph
third paragraph
`

// secretDoc carries something the egress redactor rewrites, so get_doc's body
// differs from what is on disk.
const secretDoc = `---
title: Has A Secret
---

the key is AKIAIOSFODNN7EXAMPLE and it matters
trailing line
`

func (f *writeFixture) readDoc(t *testing.T, slug string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(f.root, "content", slug))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func (f *writeFixture) hashOf(t *testing.T, slug string) string {
	t.Helper()
	_, out, err := f.w.reader().getDoc(context.Background(), nil, getDocIn{Slug: slug})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Found {
		t.Fatalf("%s not readable", slug)
	}
	return out.ContentHash
}

func (f *writeFixture) update(t *testing.T, in updateDocIn) writeOut {
	t.Helper()
	_, out, err := f.w.updateDoc(context.Background(), nil, in)
	if err != nil {
		t.Fatalf("update_doc: %v", err)
	}
	return out
}

// --- 3.1 anchored replace -------------------------------------------------

func TestUpdateDocAnchoredReplace(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u1.md", updDoc)

	out := f.update(t, updateDocIn{Slug: "u1.md", Old: "second paragraph", New: "SECOND, rewritten"})
	if !out.OK {
		t.Fatalf("refused: %s", out.Message)
	}
	got := f.readDoc(t, "u1.md")
	want := strings.Replace(updDoc, "second paragraph", "SECOND, rewritten", 1)
	if got != want {
		t.Fatalf("document is not a single-anchor replacement:\n%q", got)
	}
}

// --- 3.2 anchored refusals ------------------------------------------------

func TestUpdateDocAnchoredRefusals(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u2.md", "---\ntitle: T\n---\n\nrepeat\nrepeat\nunique\n")
	before := f.readDoc(t, "u2.md")

	t.Run("no match", func(t *testing.T) {
		out := f.update(t, updateDocIn{Slug: "u2.md", Old: "absent", New: "x"})
		if out.OK {
			t.Fatal("a missing anchor must be refused")
		}
		if !strings.Contains(out.Message, "not found") {
			t.Errorf("unhelpful message: %q", out.Message)
		}
	})

	t.Run("ambiguous", func(t *testing.T) {
		out := f.update(t, updateDocIn{Slug: "u2.md", Old: "repeat", New: "x"})
		if out.OK {
			t.Fatal("an ambiguous anchor must be refused")
		}
		if !strings.Contains(out.Message, "2 matches") {
			t.Errorf("message should name the match count, got %q", out.Message)
		}
	})

	if got := f.readDoc(t, "u2.md"); got != before {
		t.Fatalf("a refused anchor still changed the document:\n%q", got)
	}
}

// --- 3.3 anchors cannot carry redaction into the corpus -------------------

func TestUpdateDocAnchorFromRedactedTextFailsClosed(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u3.md", secretDoc)

	// What the agent sees is redacted, so an anchor taken from it holds the
	// placeholder rather than the stored text.
	_, doc, err := f.w.reader().getDoc(context.Background(), nil, getDocIn{Slug: "u3.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !doc.Redacted {
		t.Fatal("fixture should be redacted on read")
	}
	var anchor string
	for _, line := range strings.Split(doc.Body, "\n") {
		if strings.Contains(line, "the key is") {
			anchor = line
		}
	}
	if anchor == "" || strings.Contains(anchor, "AKIAIOSFODNN7EXAMPLE") {
		t.Fatalf("expected a redacted anchor line, got %q", anchor)
	}

	out := f.update(t, updateDocIn{Slug: "u3.md", Old: anchor, New: "replaced"})
	if out.OK {
		t.Fatal("an anchor built from redacted text must not match")
	}
	got := f.readDoc(t, "u3.md")
	if got != secretDoc {
		t.Fatalf("document changed:\n%q", got)
	}
	if strings.Contains(got, "REDACTED") {
		t.Fatal("a redaction placeholder was written into the corpus")
	}
}

// --- 3.4 / 3.5 whole-body -------------------------------------------------

func TestUpdateDocWholeBody(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u4.md", updDoc)
	hash := f.hashOf(t, "u4.md")

	out := f.update(t, updateDocIn{Slug: "u4.md", Body: "entirely new prose\n", BaseHash: hash})
	if !out.OK {
		t.Fatalf("refused: %s", out.Message)
	}
	got := f.readDoc(t, "u4.md")
	if !strings.Contains(got, "entirely new prose") || strings.Contains(got, "second paragraph") {
		t.Fatalf("body not replaced:\n%s", got)
	}
}

func TestUpdateDocWholeBodyRefusals(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u5.md", updDoc)
	f.putDoc(t, "u5secret.md", secretDoc)
	hash := f.hashOf(t, "u5.md")

	cases := []struct {
		name string
		in   updateDocIn
		want string
	}{
		{"missing base_hash", updateDocIn{Slug: "u5.md", Body: "x\n"}, "base_hash"},
		{"stale base_hash", updateDocIn{Slug: "u5.md", Body: "x\n", BaseHash: strings.Repeat("0", 64)}, "stale"},
		{"redacted document", updateDocIn{Slug: "u5secret.md", Body: "x\n", BaseHash: f.hashOf(t, "u5secret.md")}, "redacted"},
		{"echoed fence", updateDocIn{Slug: "u5.md", Body: defense.Delimit("x"), BaseHash: hash}, "delimiter"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := f.readDoc(t, c.in.Slug)
			out := f.update(t, c.in)
			if out.OK {
				t.Fatalf("%s must be refused", c.name)
			}
			if !strings.Contains(out.Message, c.want) {
				t.Errorf("message %q should mention %q", out.Message, c.want)
			}
			if got := f.readDoc(t, c.in.Slug); got != before {
				t.Fatalf("document changed on a refused write:\n%q", got)
			}
		})
	}
}

func TestUpdateDocRejectsBothOrNeitherMode(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u6.md", updDoc)
	for _, in := range []updateDocIn{
		{Slug: "u6.md", Old: "first paragraph", New: "x", Body: "y\n"},
		{Slug: "u6.md"},
	} {
		if out := f.update(t, in); out.OK {
			t.Errorf("must be refused: %+v", in)
		}
	}
}

// --- 3.6 front matter is untouchable --------------------------------------

func TestUpdateDocPreservesFrontMatter(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u7.md", updDoc)
	wantFM := f.frontMatterOf(t, "u7.md")
	hash := f.hashOf(t, "u7.md")

	// A body that itself looks like front matter must stay body.
	body := "---\ntitle: Impostor\n---\n\nreal body\n"
	if out := f.update(t, updateDocIn{Slug: "u7.md", Body: body, BaseHash: hash}); !out.OK {
		t.Fatalf("refused: %s", out.Message)
	}

	if got := f.frontMatterOf(t, "u7.md"); got != wantFM {
		t.Fatalf("front matter changed:\nwant %q\ngot  %q", wantFM, got)
	}
	if !strings.Contains(wantFM, "# a comment") {
		t.Fatal("fixture should carry a comment to protect")
	}
	_, _, err := splitFrontMatter(f.readDoc(t, "u7.md"))
	if err != nil {
		t.Fatalf("document no longer parses: %v", err)
	}
}

// --- 3.7 empty-body guard -------------------------------------------------

func TestUpdateDocEmptyBodyGuard(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u8.md", updDoc)
	hash := f.hashOf(t, "u8.md")

	if out := f.update(t, updateDocIn{Slug: "u8.md", Old: "first paragraph\nsecond paragraph\nthird paragraph\n", New: "  \n"}); out.OK {
		t.Fatal("emptying a document must need allow_empty")
	}
	if out := f.update(t, updateDocIn{Slug: "u8.md", Body: "\n", BaseHash: hash, AllowEmpty: true}); !out.OK {
		t.Fatalf("allow_empty should permit it: %s", out.Message)
	}
	if got := f.frontMatterOf(t, "u8.md"); !strings.Contains(got, "title: Update Me") {
		t.Fatalf("front matter lost on a deliberate wipe:\n%s", got)
	}
}

// --- 3.8 pipeline refusals ------------------------------------------------

func TestUpdateDocRefusedWhenDegraded(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "u9.md", updDoc)
	before := f.readDoc(t, "u9.md")
	if err := os.WriteFile(filepath.Join(f.root, "content", "conflict.md"),
		[]byte("<<<<<<< HEAD\nx\n>>>>>>> y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out := f.update(t, updateDocIn{Slug: "u9.md", Old: "first paragraph", New: "x"}); out.OK {
		t.Fatal("must be refused while degraded")
	}
	if got := f.readDoc(t, "u9.md"); got != before {
		t.Fatal("document changed while degraded")
	}
}

func TestUpdateDocScopeRules(t *testing.T) {
	t.Run("write:inbox cannot touch a normal doc", func(t *testing.T) {
		f := newWriteFixture(t, "read:*", "write:inbox")
		f.putDoc(t, "ua.md", updDoc)
		if out := f.update(t, updateDocIn{Slug: "ua.md", Old: "first paragraph", New: "x"}); out.OK {
			t.Fatal("modifying a non-draft must require write:*")
		}
	})

	t.Run("unreadable target is not-found", func(t *testing.T) {
		f := newWriteFixture(t, "read:*", "write:*")
		out := f.update(t, updateDocIn{Slug: "no_such_doc.md", Old: "a", New: "b"})
		if out.OK || !strings.Contains(out.Message, "not found") {
			t.Fatalf("expected not found, got %+v", out)
		}
	})
}

// --- 3.9 membership + audit diff ------------------------------------------

func TestUpdateDocAuditsADiff(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "ub.md", updDoc)

	out := f.update(t, updateDocIn{Slug: "ub.md", Old: "second paragraph", New: "changed line"})
	if !out.OK {
		t.Fatalf("refused: %s", out.Message)
	}
	if !hasMembership(out.Membership, "customer", "eric") {
		t.Errorf("membership not returned: %+v", out.Membership)
	}
	if !f.auditContains(t, "update_doc", "ok") {
		t.Fatal("expected an ok audit entry")
	}
	b, err := os.ReadFile(f.audit)
	if err != nil {
		t.Fatal(err)
	}
	log := string(b)
	if !strings.Contains(log, "@@") || !strings.Contains(log, "+changed line") {
		t.Fatalf("audit payload is not a diff: %s", log)
	}
	// The log records the change, not a second copy of the document: the diff
	// covers the body only, so front matter never reaches it.
	if strings.Contains(log, "title: Update Me") {
		t.Fatalf("audit stored the front matter: %s", log)
	}
	if strings.Contains(log, `"diff":"`+updDoc) {
		t.Fatalf("audit stored the full document: %s", log)
	}
}

// --- 3.10 protocol level --------------------------------------------------

func TestUpdateDocOverMCP(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "wire2.md", updDoc)

	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	f.server.Auth = auth.NewStaticTokenAuthenticator([]auth.TokenRecord{
		{Name: "e2e", Hash: auth.HashToken(token), Scopes: []string{"read:*", "write:*"}},
	})
	ts := httptest.NewServer(f.server.Handler())
	defer ts.Close()

	ctx := context.Background()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "0"}, nil)
	session, err := client.Connect(ctx, &mcpsdk.StreamableClientTransport{
		Endpoint:   ts.URL + "/mcp",
		HTTPClient: &http.Client{Transport: bearerRT{http.DefaultTransport, token}},
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() { _ = session.Close() }()

	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "get_doc", Arguments: map[string]any{"slug": "wire2.md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var doc getDocOut
	decodeStructured(t, res, &doc)
	if doc.ContentHash == "" {
		t.Fatal("get_doc must return a content_hash")
	}
	if doc.Redacted {
		t.Fatal("fixture is not redacted")
	}

	res, err = session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "update_doc",
		Arguments: map[string]any{
			"slug": "wire2.md", "body": "rewritten over the wire\n", "base_hash": doc.ContentHash,
		},
	})
	if err != nil {
		t.Fatalf("call update_doc: %v", err)
	}
	var wo writeOut
	decodeStructured(t, res, &wo)
	if !wo.OK {
		t.Fatalf("update_doc failed over the wire: %s", wo.Message)
	}

	res, err = session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name: "get_doc", Arguments: map[string]any{"slug": "wire2.md"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var after getDocOut
	decodeStructured(t, res, &after)
	if !strings.Contains(after.Body, "rewritten over the wire") {
		t.Fatalf("read-back does not show the update: %q", after.Body)
	}
	if after.ContentHash == doc.ContentHash {
		t.Fatal("content_hash should change after a write")
	}
}

// --- the diff helper ------------------------------------------------------

func TestUnifiedDiff(t *testing.T) {
	if got := unifiedDiff("same\n", "same\n"); got != "" {
		t.Errorf("identical text should produce no diff, got %q", got)
	}
	got := unifiedDiff("a\nb\nc\n", "a\nB\nc\n")
	if !strings.Contains(got, "@@") || !strings.Contains(got, "-b") || !strings.Contains(got, "+B") {
		t.Errorf("unexpected diff:\n%s", got)
	}
	// Context keeps unchanged neighbours, so a small edit is legible.
	if !strings.Contains(got, " a") || !strings.Contains(got, " c") {
		t.Errorf("diff should carry context:\n%s", got)
	}
	// A pure addition and a pure deletion both render.
	if d := unifiedDiff("", "new\n"); !strings.Contains(d, "+new") {
		t.Errorf("addition not rendered: %q", d)
	}
	if d := unifiedDiff("gone\n", ""); !strings.Contains(d, "-gone") {
		t.Errorf("deletion not rendered: %q", d)
	}
}
