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
	"github.com/linden-project/linny-mcp-server/internal/index"
)

// putDoc writes a document with known front matter and reindexes, so the write
// tools (which only touch documents the store says are readable) can see it.
func (f *writeFixture) putDoc(t *testing.T, slug, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(f.root, "content", slug), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	f.reindex(t)
}

func (f *writeFixture) reindex(t *testing.T) {
	t.Helper()
	g, _, err := index.Build(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.w.store.Populate(g); err != nil {
		t.Fatal(err)
	}
}

// frontMatterOf returns the raw front-matter block of a document on disk.
func (f *writeFixture) frontMatterOf(t *testing.T, slug string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(f.root, "content", slug))
	if err != nil {
		t.Fatal(err)
	}
	fm, _, err := splitFrontMatter(string(b))
	if err != nil {
		t.Fatalf("splitting front matter of %s: %v", slug, err)
	}
	return fm
}

const plainDoc = `---
title: Typed Note
customer: eric
---

body stays put
`

// --- 4.1 type round-trips -------------------------------------------------

func TestSetFrontMatterPreservesTypes(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "typed.md", plainDoc)

	cases := []struct {
		key   string
		value any
		want  string
	}{
		{"priority", float64(3), "priority: 3"},
		{"ratio", 3.5, "ratio: 3.5"},
		{"starred", true, "starred: true"},
		{"cleared", nil, "cleared: null"},
		{"day", "2026-09-14", `day: "2026-09-14"`},
		{"label", "plain", "label: plain"},
	}
	for _, c := range cases {
		_, out, err := f.w.setFrontMatter(context.Background(), nil,
			setFMIn{Slug: "typed.md", Key: c.key, Value: c.value})
		if err != nil {
			t.Fatalf("set %s: %v", c.key, err)
		}
		if !out.OK {
			t.Fatalf("set %s refused: %s", c.key, out.Message)
		}
	}

	fm := f.frontMatterOf(t, "typed.md")
	for _, c := range cases {
		if !strings.Contains(fm, c.want) {
			t.Errorf("front matter missing %q:\n%s", c.want, fm)
		}
	}
	// The stringifying default branch is gone: no value is a quoted number.
	if strings.Contains(fm, "priority: '3'") || strings.Contains(fm, `priority: "3"`) {
		t.Errorf("number was written as a string:\n%s", fm)
	}
}

// --- 4.2 sequences --------------------------------------------------------

func TestSetFrontMatterWritesSequence(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "seq.md", plainDoc)

	_, out, err := f.w.setFrontMatter(context.Background(), nil,
		setFMIn{Slug: "seq.md", Key: "tags", Value: []any{"note", "idea"}})
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatalf("refused: %s", out.Message)
	}

	fm := f.frontMatterOf(t, "seq.md")
	if !strings.Contains(fm, "- note") || !strings.Contains(fm, "- idea") {
		t.Fatalf("expected a block sequence:\n%s", fm)
	}
	// The bug this change fixes: a list stringified into a single junk term.
	if strings.Contains(fm, "[note") {
		t.Fatalf("list was stringified:\n%s", fm)
	}
	for _, want := range []string{"note", "idea"} {
		if !hasMembership(out.Membership, "tags", want) {
			t.Errorf("membership missing tags:%s, got %+v", want, out.Membership)
		}
	}
}

func hasMembership(ms []index.TermMembership, tax, term string) bool {
	for _, m := range ms {
		if m.Taxonomy == tax && m.Term == term {
			return true
		}
	}
	return false
}

// --- 4.3 nested structures ------------------------------------------------

func TestSetFrontMatterRefusesNested(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "nested.md", plainDoc)
	before := f.frontMatterOf(t, "nested.md")

	for name, v := range map[string]any{
		"mapping":     map[string]any{"a": "b"},
		"nested list": []any{[]any{"a"}},
	} {
		_, out, err := f.w.setFrontMatter(context.Background(), nil,
			setFMIn{Slug: "nested.md", Key: "meta", Value: v})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if out.OK {
			t.Errorf("%s must be refused", name)
		}
	}
	if got := f.frontMatterOf(t, "nested.md"); got != before {
		t.Fatalf("document changed on a refused write:\n%s", got)
	}
}

// --- 4.4 inert taxonomy values --------------------------------------------

func TestTaxonomyValueMustProduceMembership(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "inert.md", plainDoc)

	// `tags` is a declared taxonomy: a number would index as nothing at all.
	_, out, err := f.w.setFrontMatter(context.Background(), nil,
		setFMIn{Slug: "inert.md", Key: "tags", Value: float64(3)})
	if err != nil {
		t.Fatal(err)
	}
	if out.OK {
		t.Fatal("a number on a taxonomy key must be refused")
	}
	if !strings.Contains(out.Message, "tags") {
		t.Errorf("refusal should name the taxonomy, got %q", out.Message)
	}

	// An ordinary property has no such constraint.
	_, out, err = f.w.setFrontMatter(context.Background(), nil,
		setFMIn{Slug: "inert.md", Key: "priority", Value: float64(3)})
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatalf("a number on a non-taxonomy key must be accepted: %s", out.Message)
	}
}

// --- 4.5 / 4.6 add_term ---------------------------------------------------

func TestAddTermShapes(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")

	t.Run("absent key becomes a sequence", func(t *testing.T) {
		f.putDoc(t, "a1.md", plainDoc)
		mustAddTerm(t, f, "a1.md", "tags", "note")
		fm := f.frontMatterOf(t, "a1.md")
		if !strings.Contains(fm, "tags:") || !strings.Contains(fm, "- note") {
			t.Fatalf("expected tags as a sequence:\n%s", fm)
		}
	})

	t.Run("scalar is promoted", func(t *testing.T) {
		f.putDoc(t, "a2.md", "---\ntitle: T\ntags: note\n---\n\nbody\n")
		mustAddTerm(t, f, "a2.md", "tags", "idea")
		fm := f.frontMatterOf(t, "a2.md")
		if !strings.Contains(fm, "- note") || !strings.Contains(fm, "- idea") {
			t.Fatalf("scalar not promoted to a sequence:\n%s", fm)
		}
	})

	t.Run("sequence is appended", func(t *testing.T) {
		f.putDoc(t, "a3.md", "---\ntitle: T\ntags:\n  - note\n---\n\nbody\n")
		mustAddTerm(t, f, "a3.md", "tags", "todo")
		fm := f.frontMatterOf(t, "a3.md")
		if !strings.Contains(fm, "- note") || !strings.Contains(fm, "- todo") {
			t.Fatalf("term not appended:\n%s", fm)
		}
	})

	t.Run("already a member is a no-op", func(t *testing.T) {
		f.putDoc(t, "a4.md", "---\ntitle: T\ntags:\n  - note\n---\n\nbody\n")
		before, err := os.ReadFile(filepath.Join(f.root, "content", "a4.md"))
		if err != nil {
			t.Fatal(err)
		}
		mustAddTerm(t, f, "a4.md", "tags", "note")
		after, err := os.ReadFile(filepath.Join(f.root, "content", "a4.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(before) != string(after) {
			t.Fatalf("document rewritten on a no-op:\nbefore:\n%s\nafter:\n%s", before, after)
		}
	})
}

func TestAddTermMatchesNormalizedMembership(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "norm.md", "---\ntitle: T\ntags:\n  - new-term\n---\n\nbody\n")
	before, err := os.ReadFile(filepath.Join(f.root, "content", "norm.md"))
	if err != nil {
		t.Fatal(err)
	}
	// "New Term" normalizes to "new-term": already a member.
	mustAddTerm(t, f, "norm.md", "tags", "New Term")
	after, err := os.ReadFile(filepath.Join(f.root, "content", "norm.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("duplicate term added:\n%s", after)
	}
}

func mustAddTerm(t *testing.T, f *writeFixture, slug, tax, term string) writeOut {
	t.Helper()
	_, out, err := f.w.addTerm(context.Background(), nil, termIn{Slug: slug, Taxonomy: tax, Term: term})
	if err != nil {
		t.Fatalf("add_term: %v", err)
	}
	if !out.OK {
		t.Fatalf("add_term refused: %s", out.Message)
	}
	return out
}

// --- 4.7 remove_term ------------------------------------------------------

func TestRemoveTerm(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")

	t.Run("removing the last term removes the key", func(t *testing.T) {
		f.putDoc(t, "r1.md", "---\ntitle: T\ntags:\n  - note\n---\n\nbody\n")
		_, out, err := f.w.removeTerm(context.Background(), nil,
			termIn{Slug: "r1.md", Taxonomy: "tags", Term: "note"})
		if err != nil {
			t.Fatal(err)
		}
		if !out.OK {
			t.Fatalf("refused: %s", out.Message)
		}
		fm := f.frontMatterOf(t, "r1.md")
		if strings.Contains(fm, "tags") {
			t.Fatalf("key should be gone, not left empty:\n%s", fm)
		}
	})

	t.Run("one of several", func(t *testing.T) {
		f.putDoc(t, "r2.md", "---\ntitle: T\ntags:\n  - note\n  - idea\n---\n\nbody\n")
		if _, out, err := f.w.removeTerm(context.Background(), nil,
			termIn{Slug: "r2.md", Taxonomy: "tags", Term: "note"}); err != nil || !out.OK {
			t.Fatalf("refused: %v %+v", err, out)
		}
		fm := f.frontMatterOf(t, "r2.md")
		if strings.Contains(fm, "- note") || !strings.Contains(fm, "- idea") {
			t.Fatalf("wrong term removed:\n%s", fm)
		}
	})

	t.Run("absent term is a no-op", func(t *testing.T) {
		f.putDoc(t, "r3.md", "---\ntitle: T\ntags:\n  - note\n---\n\nbody\n")
		before, err := os.ReadFile(filepath.Join(f.root, "content", "r3.md"))
		if err != nil {
			t.Fatal(err)
		}
		if _, out, err := f.w.removeTerm(context.Background(), nil,
			termIn{Slug: "r3.md", Taxonomy: "tags", Term: "nothing-here"}); err != nil || !out.OK {
			t.Fatalf("refused: %v %+v", err, out)
		}
		after, err := os.ReadFile(filepath.Join(f.root, "content", "r3.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(before) != string(after) {
			t.Fatalf("document rewritten on a no-op:\n%s", after)
		}
	})
}

// --- 4.8 coined terms -----------------------------------------------------

func TestNewTermsReported(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	// Declare one term explicitly (L2 files are keyed by the SINGULAR name).
	if err := os.WriteFile(filepath.Join(f.root, "lindenConfig", "L2-CONF-TAX-tag-TRM-note.yml"),
		[]byte("title: Note\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f.putDoc(t, "coin.md", plainDoc)

	out := mustAddTerm(t, f, "coin.md", "tags", "quantum-widget")
	if len(out.NewTerms) != 1 || out.NewTerms[0] != "quantum-widget" {
		t.Fatalf("expected the coined term reported, got %v", out.NewTerms)
	}

	out = mustAddTerm(t, f, "coin.md", "tags", "note")
	if len(out.NewTerms) != 0 {
		t.Fatalf("a declared term is not new, got %v", out.NewTerms)
	}
}

// --- 4.9 surgical edits ---------------------------------------------------

func TestTypedEditsStaySurgical(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	const doc = `---
title: Ordered
# keep me
customer: eric
type: log
subject: nix
---

body must not move
`
	f.putDoc(t, "order.md", doc)
	mustAddTerm(t, f, "order.md", "tags", "note")

	b, err := os.ReadFile(filepath.Join(f.root, "content", "order.md"))
	if err != nil {
		t.Fatal(err)
	}
	fm, body, err := splitFrontMatter(string(b))
	if err != nil {
		t.Fatal(err)
	}
	if body != "\nbody must not move\n" {
		t.Fatalf("body changed: %q", body)
	}
	if !strings.Contains(fm, "# keep me") {
		t.Errorf("comment lost:\n%s", fm)
	}
	title, customer, typ := strings.Index(fm, "title:"), strings.Index(fm, "customer:"), strings.Index(fm, "type:")
	if title > customer || customer > typ {
		t.Errorf("key order changed:\n%s", fm)
	}
}

// --- 3b.3 the degraded gate applies to every write tool -------------------

func TestEveryWriteToolRefusedWhenDegraded(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "deg.md", "---\ntitle: T\ntags:\n  - note\n---\n\nbody\n")
	before, err := os.ReadFile(filepath.Join(f.root, "content", "deg.md"))
	if err != nil {
		t.Fatal(err)
	}
	// A committed conflict marker degrades the guard.
	if err := os.WriteFile(filepath.Join(f.root, "content", "conflict.md"),
		[]byte("<<<<<<< HEAD\nx\n>>>>>>> y\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	calls := map[string]func() (writeOut, error){
		"create_doc": func() (writeOut, error) {
			_, o, e := f.w.createDoc(ctx, nil, createDocIn{Title: "Nope"})
			return o, e
		},
		"append_to_doc": func() (writeOut, error) {
			_, o, e := f.w.appendToDoc(ctx, nil, appendIn{Slug: "deg.md", Text: "more"})
			return o, e
		},
		"set_front_matter": func() (writeOut, error) {
			_, o, e := f.w.setFrontMatter(ctx, nil, setFMIn{Slug: "deg.md", Key: "customer", Value: "eric"})
			return o, e
		},
		"unset_front_matter": func() (writeOut, error) {
			_, o, e := f.w.unsetFrontMatter(ctx, nil, unsetFMIn{Slug: "deg.md", Key: "title"})
			return o, e
		},
		"archive": func() (writeOut, error) {
			_, o, e := f.w.archive(ctx, nil, archiveIn{Slug: "deg.md"})
			return o, e
		},
		"add_term": func() (writeOut, error) {
			_, o, e := f.w.addTerm(ctx, nil, termIn{Slug: "deg.md", Taxonomy: "tags", Term: "idea"})
			return o, e
		},
		"remove_term": func() (writeOut, error) {
			_, o, e := f.w.removeTerm(ctx, nil, termIn{Slug: "deg.md", Taxonomy: "tags", Term: "note"})
			return o, e
		},
	}
	for name, call := range calls {
		out, err := call()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if out.OK {
			t.Errorf("%s must be refused while the tree is degraded", name)
		}
	}
	after, err := os.ReadFile(filepath.Join(f.root, "content", "deg.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("a refused write still changed the document:\n%s", after)
	}
}

// --- 4.10 protocol level --------------------------------------------------

func TestAddTermOverMCP(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	f.putDoc(t, "wire.md", plainDoc)

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

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tl := range tools.Tools {
		names[tl.Name] = true
	}
	for _, want := range []string{"add_term", "remove_term"} {
		if !names[want] {
			t.Fatalf("tool %q not registered (have %v)", want, names)
		}
	}

	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "add_term",
		Arguments: map[string]any{"slug": "wire.md", "taxonomy": "tags", "term": "idea"},
	})
	if err != nil {
		t.Fatalf("call add_term: %v", err)
	}
	var wo writeOut
	decodeStructured(t, res, &wo)
	if !wo.OK {
		t.Fatalf("add_term failed over the wire: %s", wo.Message)
	}

	res, err = session.CallTool(ctx, &mcpsdk.CallToolParams{
		Name:      "get_doc",
		Arguments: map[string]any{"slug": "wire.md"},
	})
	if err != nil {
		t.Fatalf("call get_doc: %v", err)
	}
	var doc getDocOut
	decodeStructured(t, res, &doc)
	if !doc.Found {
		t.Fatal("document should be readable")
	}
	tags, ok := doc.Props["tags"].([]any)
	if !ok || len(tags) != 1 || tags[0] != "idea" {
		t.Fatalf("expected tags as a list carrying idea, got %#v", doc.Props["tags"])
	}
}
