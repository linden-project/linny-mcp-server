package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/linden-project/linny-mcp-server/internal/auth"
	"github.com/linden-project/linny-mcp-server/internal/authz"
)

// starredFixture builds a corpus with known starred documents. `tags` is a
// declared taxonomy in the synthetic corpus, so denying one of its terms gives a
// caller that cannot read a specific starred document.
func starredFixture(t *testing.T, scopes ...string) *writeFixture {
	t.Helper()
	f := newWriteFixture(t, scopes...)
	f.putDoc(t, "star_one.md", "---\ntitle: Alpha Note\nstarred: true\ntags:\n  - work\n---\n\nbody\n")
	f.putDoc(t, "star_two.md", "---\ntitle: Beta Note\nstarred: true\ntags:\n  - health\n---\n\nbody\n")
	f.putDoc(t, "plain_one.md", "---\ntitle: Not Starred\n---\n\nbody\n")
	f.putDoc(t, "explicit_false.md", "---\ntitle: Explicitly False\nstarred: false\n---\n\nbody\n")
	return f
}

func (f *writeFixture) starred(t *testing.T) starredDocsOut {
	t.Helper()
	_, out, err := f.w.reader().starredDocs(context.Background(), nil, emptyIn{})
	if err != nil {
		t.Fatalf("starred_docs: %v", err)
	}
	return out
}

func slugsOf(out starredDocsOut) []string {
	s := make([]string, 0, len(out.Docs))
	for _, d := range out.Docs {
		s = append(s, d.Filename)
	}
	return s
}

func TestStarredDocsReturnsStarredWithTitles(t *testing.T) {
	f := starredFixture(t, "read:*")
	out := f.starred(t)

	got := map[string]string{}
	for _, d := range out.Docs {
		got[d.Filename] = d.Title
	}
	if got["star_one.md"] != "Alpha Note" || got["star_two.md"] != "Beta Note" {
		t.Fatalf("starred documents missing or untitled: %+v", out.Docs)
	}
	if _, ok := got["plain_one.md"]; ok {
		t.Error("a document with no starred key must not appear")
	}
	if _, ok := got["explicit_false.md"]; ok {
		t.Error("starred: false must not appear")
	}
	if !out.ScopeFiltered {
		t.Error("scope_filtered must be reported even for read:*")
	}
}

// The synthetic corpus stars a document of its own, so this also proves the tool
// reads real front matter rather than only the documents a test wrote.
func TestStarredDocsFindsCorpusStarredDocuments(t *testing.T) {
	f := newWriteFixture(t, "read:*")
	out := f.starred(t)
	if len(out.Docs) == 0 {
		t.Fatal("the generated corpus stars at least one document; none found")
	}
	for _, d := range out.Docs {
		if d.Title == "" {
			t.Errorf("%s returned without a title", d.Filename)
		}
	}
}

func TestStarredDocsHidesDeniedDocumentsWithoutTrace(t *testing.T) {
	full := starredFixture(t, "read:*")
	before := full.starred(t)

	// Same corpus, a caller denied the tags:health term.
	denied := starredFixture(t, "read:*", "deny:taxonomy:tags:health")
	after := denied.starred(t)

	if !contains(slugsOf(before), "star_two.md") {
		t.Fatalf("fixture should expose the health-tagged starred doc, got %v", slugsOf(before))
	}
	got := slugsOf(after)
	if contains(got, "star_two.md") {
		t.Fatalf("a denied starred document leaked: %v", got)
	}
	if !contains(got, "star_one.md") {
		t.Fatalf("a readable starred document went missing: %v", got)
	}
	// The filtered response must be shaped exactly like an unfiltered one: no
	// count, no marker, nothing that says something was removed.
	if after.ScopeFiltered != before.ScopeFiltered {
		t.Error("scope_filtered must not vary with what scope removed")
	}
}

func TestStarredDocsRedactsTitles(t *testing.T) {
	f := newWriteFixture(t, "read:*")
	f.putDoc(t, "secret_title.md", "---\ntitle: key AKIAIOSFODNN7EXAMPLE here\nstarred: true\n---\n\nbody\n")

	var title string
	for _, d := range f.starred(t).Docs {
		if d.Filename == "secret_title.md" {
			title = d.Title
		}
	}
	if title == "" {
		t.Fatal("the starred document with a secret in its title was not returned")
	}
	if strings.Contains(title, "AKIAIOSFODNN7EXAMPLE") {
		t.Fatalf("credential leaked through the title: %q", title)
	}
}

func TestStarredDocsOrderingIsStable(t *testing.T) {
	f := starredFixture(t, "read:*")
	one, two := f.starred(t), f.starred(t)
	if strings.Join(slugsOf(one), ",") != strings.Join(slugsOf(two), ",") {
		t.Fatalf("ordering is not stable: %v then %v", slugsOf(one), slugsOf(two))
	}
	// Ordered by title, so the list is non-decreasing on Title.
	for i := 1; i < len(one.Docs); i++ {
		if one.Docs[i-1].Title > one.Docs[i].Title {
			t.Fatalf("not ordered by title: %q before %q", one.Docs[i-1].Title, one.Docs[i].Title)
		}
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func TestStarredDocsDenyByDefault(t *testing.T) {
	// A token with no read rule reads nothing, so it sees no starred documents.
	f := starredFixture(t)
	ss, err := authz.Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	w := newWriter(f.server, ss, "tester")
	_, out, err := w.reader().starredDocs(context.Background(), nil, emptyIn{})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Docs) != 0 {
		t.Fatalf("deny-by-default must yield nothing, got %v", slugsOf(out))
	}
	if out.Docs == nil {
		t.Error("an empty result should be a list, not null")
	}
}

func TestStarredDocsOverMCP(t *testing.T) {
	f := starredFixture(t, "read:*")
	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	f.server.Auth = auth.NewStaticTokenAuthenticator([]auth.TokenRecord{
		{Name: "e2e", Hash: auth.HashToken(token), Scopes: []string{"read:*"}},
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

	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{Name: "starred_docs", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("call starred_docs: %v", err)
	}
	var out starredDocsOut
	decodeStructured(t, res, &out)
	if !contains(slugsOf(out), "star_one.md") || !contains(slugsOf(out), "star_two.md") {
		t.Fatalf("unexpected result over the wire: %+v", out.Docs)
	}
	if !out.ScopeFiltered {
		t.Error("scope_filtered missing over the wire")
	}
}
