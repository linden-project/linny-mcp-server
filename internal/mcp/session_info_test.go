package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/linden-project/linny-mcp-server/internal/auth"
	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
)

// sessionOf drives session_info through the real HTTP handler, so the threading
// in server.go is what is under test rather than a copy of it in the fixture.
func sessionOf(t *testing.T, f *writeFixture, scopes ...string) sessionInfoOut {
	t.Helper()
	token, err := auth.GenerateToken()
	if err != nil {
		t.Fatal(err)
	}
	f.server.Auth = auth.NewStaticTokenAuthenticator([]auth.TokenRecord{
		{Name: "claude-test", Hash: auth.HashToken(token), Scopes: scopes},
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

	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{Name: "session_info", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("call session_info: %v", err)
	}
	var out sessionInfoOut
	decodeStructured(t, res, &out)
	return out
}

func TestSessionInfoReportsServerNotebookAndCaller(t *testing.T) {
	f := newWriteFixture(t, "read:*")
	f.server.NotebookName = "testbook"

	out := sessionOf(t, f, "read:*", "write:inbox")

	if out.Server.Name != "linny-mcp" || out.Server.Version != buildinfo.Version {
		t.Errorf("server not reported: %+v", out.Server)
	}
	if out.Notebook.Name != "testbook" {
		t.Errorf("notebook name = %q, want testbook", out.Notebook.Name)
	}
	if !out.Notebook.Quarantine || !out.Notebook.WritesEnabled {
		t.Errorf("notebook flags wrong: %+v", out.Notebook)
	}
	if out.Caller.Identity != "claude-test" {
		t.Errorf("identity = %q", out.Caller.Identity)
	}
	if strings.Join(out.Caller.Scopes, ",") != "read:*,write:inbox" {
		t.Errorf("raw scopes not echoed: %v", out.Caller.Scopes)
	}
	if !out.Caller.CanRead {
		t.Error("read:* should report can_read")
	}
}

func TestSessionInfoModifyPermissionHasThreeStates(t *testing.T) {
	cases := []struct {
		scopes []string
		modify string
		create bool
	}{
		{[]string{"read:*", "write:*"}, modifyAll, true},
		{[]string{"read:*", "write:inbox"}, modifyOwnDrafts, true},
		{[]string{"read:*"}, modifyNone, false},
	}
	for _, c := range cases {
		t.Run(c.modify, func(t *testing.T) {
			f := newWriteFixture(t, "read:*")
			out := sessionOf(t, f, c.scopes...)
			if out.Caller.CanModify != c.modify {
				t.Errorf("can_modify = %q, want %q", out.Caller.CanModify, c.modify)
			}
			if out.Caller.CanCreate != c.create {
				t.Errorf("can_create = %v, want %v", out.Caller.CanCreate, c.create)
			}
		})
	}
}

// The response describes the connection and the credential. Nothing about the
// corpus, and no host layout, may appear in it.
func TestSessionInfoDisclosesNothingAboutTheCorpus(t *testing.T) {
	f := newWriteFixture(t, "read:*")
	f.server.NotebookName = "testbook"
	f.putDoc(t, "secret_slug.md", "---\ntitle: Findable\ntags:\n  - health\n---\n\nbody\n")
	if err := os.WriteFile(filepath.Join(f.root, "content", "conflict.md"),
		[]byte("<<<<<<< HEAD\nx\n>>>>>>> y\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := sessionOf(t, f, "read:*", "write:*")
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)

	for _, forbidden := range []string{f.root, "/content", "secret_slug", "conflict.md", "health", "tags"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("response discloses %q: %s", forbidden, body)
		}
	}
	if strings.Contains(body, "can_delete") {
		t.Error("can_delete must not be reported: no tool implements delete")
	}
}

func TestSessionInfoWriteVerdict(t *testing.T) {
	clean := sessionStatusOut{}
	degraded := sessionStatusOut{Degraded: true, Reason: "committed conflict markers"}
	parse := func(scopes ...string) *authz.ScopeSet {
		ss, err := authz.Parse(scopes)
		if err != nil {
			t.Fatal(err)
		}
		return ss
	}

	cases := []struct {
		name     string
		enabled  bool
		status   sessionStatusOut
		scopes   *authz.ScopeSet
		want     bool
		wantWord string
	}{
		{"registration blocks", false, clean, parse("read:*", "write:*"), false, "read-only"},
		{"degraded blocks", true, degraded, parse("read:*", "write:*"), false, "degraded"},
		{"no write scope blocks", true, clean, parse("read:*"), false, "write scope"},
		{"write:* permits", true, clean, parse("read:*", "write:*"), true, ""},
		{"write:inbox permits creating", true, clean, parse("read:*", "write:inbox"), true, "write:*"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, reason := writeVerdict(c.enabled, c.status, c.scopes)
			if got != c.want {
				t.Fatalf("can_write_now = %v, want %v (reason %q)", got, c.want, reason)
			}
			if c.wantWord != "" && !strings.Contains(reason, c.wantWord) {
				t.Errorf("reason %q should mention %q", reason, c.wantWord)
			}
		})
	}
}

// The gate order is registration, then guard, then scope: a server with writes
// off reports that, not a scope problem, even when the scopes are also narrow.
func TestSessionInfoNamesTheFirstFailingGate(t *testing.T) {
	ss, err := authz.Parse([]string{"read:*"})
	if err != nil {
		t.Fatal(err)
	}
	_, reason := writeVerdict(false, sessionStatusOut{Degraded: true}, ss)
	if !strings.Contains(reason, "read-only") {
		t.Fatalf("registration should dominate, got %q", reason)
	}
}

func TestSessionInfoAgreesWithSyncStatusOnDegraded(t *testing.T) {
	f := newWriteFixture(t, "read:*")
	if err := os.WriteFile(filepath.Join(f.root, "content", "boom.md"),
		[]byte("<<<<<<< HEAD\nx\n>>>>>>> y\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := sessionOf(t, f, "read:*", "write:*")
	if !out.Status.Degraded {
		t.Fatal("session_info should report the degraded tree")
	}
	if out.CanWriteNow {
		t.Fatal("a degraded tree must block writing")
	}

	// sync_status, reading from the same accessor, must agree.
	g := f.server.Guard
	rd := f.w.reader()
	rd.syncStatus = func() SyncStatus {
		st := g.State()
		return SyncStatus{Degraded: !st.Clean || g.ForcedReadOnly(), Conflicted: st.Conflicted, Reason: st.Reason}
	}
	_, sync, err := rd.syncStatusTool(context.Background(), nil, emptyIn{})
	if err != nil {
		t.Fatal(err)
	}
	if sync.Degraded != out.Status.Degraded {
		t.Fatalf("sync_status says degraded=%v, session_info says %v", sync.Degraded, out.Status.Degraded)
	}
}

// Registration and the report share one predicate, so a server that registers no
// write tools cannot report writes as enabled.
func TestSessionInfoWritesEnabledMatchesRegistration(t *testing.T) {
	f := newWriteFixture(t, "read:*")
	f.server.Audit = nil // no audit log: the write tools are not registered

	if f.server.writesEnabled() {
		t.Fatal("writesEnabled must be false without an audit log")
	}
	out := sessionOf(t, f, "read:*", "write:*")
	if out.Notebook.WritesEnabled {
		t.Error("writes_enabled must follow registration")
	}
	if out.CanWriteNow {
		t.Error("can_write_now must be false when no write tool exists")
	}
}

// The case that motivated the tool: a write:inbox token reporting exactly what it
// can and cannot do, instead of the caller discovering it through a refusal.
func TestSessionInfoDiagnosesTheInboxOnlyToken(t *testing.T) {
	f := newWriteFixture(t, "read:*")
	out := sessionOf(t, f, "read:*", "write:inbox")

	if out.Caller.CanModify != modifyOwnDrafts || !out.Caller.CanCreate {
		t.Fatalf("expected draft-only modify with creating allowed: %+v", out.Caller)
	}
	if !out.CanWriteNow {
		t.Fatal("this token can create, so writing is possible")
	}
	if !strings.Contains(out.Reason, "write:*") {
		t.Fatalf("the reason should name the scope needed to modify: %q", out.Reason)
	}
}
