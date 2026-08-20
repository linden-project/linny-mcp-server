package mcp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/linden-project/linny-mcp-server/internal/audit"
	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/corpus"
	"github.com/linden-project/linny-mcp-server/internal/defense"
	"github.com/linden-project/linny-mcp-server/internal/gitsafe"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

type writeFixture struct {
	w      *writer
	root   string
	audit  string // audit log path
	server *Server
}

func newWriteFixture(t *testing.T, scopes ...string) *writeFixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	// Clean corpus (no edge cases): the edge-case set includes a committed
	// conflict marker, which would keep the git-safety guard degraded.
	if err := corpus.Generate(root, corpus.Options{Seed: 5, Count: 15, EnableEdgeCases: false}); err != nil {
		t.Fatal(err)
	}
	// Declare `status` as a taxonomy so the quarantine term is indexed.
	if err := os.WriteFile(filepath.Join(root, "lindenConfig", "L1-CONF-TAX-status.yml"),
		[]byte("title: Status\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=T", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=T", "GIT_COMMITTER_EMAIL=t@t")
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-qm", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	g, _, err := index.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	store, err := index.OpenStore(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.Populate(g); err != nil {
		t.Fatal(err)
	}

	auditPath := filepath.Join(t.TempDir(), "audit.log")
	al, err := audit.Open(auditPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = al.Close() })

	srv := &Server{
		Store: store, Redactor: redact.New(), CorpusPath: root,
		Guard: gitsafe.NewGuard(root, false), Audit: al, Policy: defense.DefaultPolicy(),
	}
	ss, err := authz.Parse(scopes)
	if err != nil {
		t.Fatal(err)
	}
	return &writeFixture{w: newWriter(srv, ss, "tester"), root: root, audit: auditPath, server: srv}
}

func (f *writeFixture) auditContains(t *testing.T, tool, outcome string) bool {
	t.Helper()
	b, err := os.ReadFile(f.audit)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(b), `"tool":"`+tool+`"`) && strings.Contains(string(b), `"outcome":"`+outcome+`"`)
}

func TestCreateDocQuarantined(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:inbox")
	_, out, err := f.w.createDoc(context.Background(), nil, createDocIn{Title: "My New Note", Body: "hello world"})
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK || out.Slug != "my_new_note.md" {
		t.Fatalf("unexpected result: %+v", out)
	}
	if !out.Quarantined {
		t.Fatal("new doc must be quarantined by default")
	}
	// Membership reflects the quarantine term (status:agent-draft) after reindex.
	found := false
	for _, m := range out.Membership {
		if m.Taxonomy == "status" && m.Term == "agent-draft" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected status:agent-draft membership, got %+v", out.Membership)
	}
	// File exists on disk.
	if _, err := os.Stat(filepath.Join(f.root, "content", "my_new_note.md")); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if !f.auditContains(t, "create_doc", "ok") {
		t.Fatal("expected an ok audit entry for create_doc")
	}
}

func TestCreateDocForbiddenWithoutWriteScope(t *testing.T) {
	f := newWriteFixture(t, "read:*") // no write scope
	_, out, err := f.w.createDoc(context.Background(), nil, createDocIn{Title: "Nope"})
	if err != nil {
		t.Fatal(err)
	}
	if out.OK {
		t.Fatal("create_doc must be forbidden without write:inbox/write:*")
	}
	if _, err := os.Stat(filepath.Join(f.root, "content", "nope.md")); !os.IsNotExist(err) {
		t.Fatal("no file should be written on a forbidden create")
	}
	if !f.auditContains(t, "create_doc", "denied") {
		t.Fatal("expected a denied audit entry")
	}
}

func TestCreateDocRefusedWhenDegraded(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:inbox")
	// Introduce a committed conflict marker → guard degrades.
	if err := os.WriteFile(filepath.Join(f.root, "content", "conflict.md"),
		[]byte("<<<<<<< HEAD\nx\n>>>>>>> y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, out, err := f.w.createDoc(context.Background(), nil, createDocIn{Title: "During Conflict"})
	if err != nil {
		t.Fatal(err)
	}
	if out.OK {
		t.Fatal("writes must be refused while the tree is degraded")
	}
	if _, err := os.Stat(filepath.Join(f.root, "content", "during_conflict.md")); !os.IsNotExist(err) {
		t.Fatal("no file should be written while degraded")
	}
}

func TestSetFrontMatterUpdatesMembership(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	slug := firstReadableSlug(t, f.w.reader())

	_, out, err := f.w.setFrontMatter(context.Background(), nil, setFMIn{Slug: slug, Key: "customer", Value: "eric"})
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatalf("set_front_matter failed: %+v", out)
	}
	found := false
	for _, m := range out.Membership {
		if m.Taxonomy == "customer" && m.Term == "eric" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected customer:eric membership after set, got %+v", out.Membership)
	}
	// The raw file carries the new key.
	b, _ := os.ReadFile(filepath.Join(f.root, "content", slug))
	if !strings.Contains(string(b), "customer: eric") {
		t.Fatalf("file missing the set key:\n%s", b)
	}
}

func TestArchiveSetsFlag(t *testing.T) {
	f := newWriteFixture(t, "read:*", "write:*")
	slug := firstReadableSlug(t, f.w.reader())
	_, out, err := f.w.archive(context.Background(), nil, archiveIn{Slug: slug})
	if err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Fatalf("archive failed: %+v", out)
	}
	b, _ := os.ReadFile(filepath.Join(f.root, "content", slug))
	if !strings.Contains(string(b), "archived: true") {
		t.Fatalf("archived flag not set:\n%s", b)
	}
}

func TestModifyForbiddenWithReadOnlyScope(t *testing.T) {
	f := newWriteFixture(t, "read:*") // can read, cannot write
	slug := firstReadableSlug(t, f.w.reader())
	_, out, err := f.w.setFrontMatter(context.Background(), nil, setFMIn{Slug: slug, Key: "customer", Value: "eric"})
	if err != nil {
		t.Fatal(err)
	}
	if out.OK {
		t.Fatal("modifying an existing doc must require write:*")
	}
}

// reader builds a read-only view over the same store/scope for helpers.
func (w *writer) reader() *reader {
	return &reader{store: w.store, red: w.red, scopeSQL: w.scopeSQL, scopeArgs: w.scopeArgs, corpusPath: w.corpusPath}
}
