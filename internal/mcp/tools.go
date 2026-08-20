package mcp

import (
	"context"
	"path/filepath"
	"sort"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
	"github.com/linden-project/linny-mcp-server/internal/defense"
	"github.com/linden-project/linny-mcp-server/internal/gitsafe"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

// contentDir is the corpus subdirectory holding records. It matches the Hugo
// template / synthetic generator; multi-content-dir support is a later refinement.
const contentDir = "content"

// reader holds everything a read tool needs, bound to one caller's scope. The
// scope is pre-compiled to a SQL subquery (deny-by-default) so every query
// filters in SQL; content fields are passed through the redactor on the way out.
type reader struct {
	store      *index.Store
	red        *redact.Redactor
	scopeSQL   string
	scopeArgs  []any
	corpusPath string            // for git history tools; may be empty
	syncStatus func() SyncStatus // operational status; when set, sync_status is registered
}

// SyncStatus is the operational git-safety state returned by the sync_status tool.
type SyncStatus struct {
	Degraded   bool     `json:"degraded"`
	Conflicted bool     `json:"conflicted"`
	Conflicts  []string `json:"conflicts,omitempty"`
	InProgress string   `json:"in_progress,omitempty"`
	Detached   bool     `json:"detached"`
	ReadOnly   bool     `json:"read_only"`
	Reason     string   `json:"reason,omitempty"`
}

// newReader builds a reader for a scope set.
func newReader(store *index.Store, red *redact.Redactor, ss *authz.ScopeSet, corpusPath string) *reader {
	sql, args := ss.ReadableFilenamesSQL()
	return &reader{store: store, red: red, scopeSQL: sql, scopeArgs: args, corpusPath: corpusPath}
}

// readable reports whether the caller's scope permits reading slug.
func (rd *reader) readable(slug string) (bool, error) {
	_, ok, err := rd.store.GetDocScoped(slug, rd.scopeSQL, rd.scopeArgs)
	return ok, err
}

// buildToolServer returns an MCP server exposing the read tools bound to rd.
func buildToolServer(rd *reader) *mcpsdk.Server {
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "linny-mcp",
		Version: buildinfo.Version,
	}, nil)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "search",
		Description: "Full-text search the notebook; returns ranked snippets with titles.",
	}, rd.search)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "get_doc",
		Description: "Fetch a single document by its slug/filename (title, front matter, body).",
	}, rd.getDoc)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "list_taxonomies",
		Description: "List the taxonomies that have at least one readable document.",
	}, rd.listTaxonomies)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "terms",
		Description: "List the terms of a taxonomy that have at least one readable document.",
	}, rd.terms)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "docs_by_term",
		Description: "List the documents tagged with a given taxonomy term.",
	}, rd.docsByTerm)

	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "history",
		Description: "Git commit history for a document (newest first).",
	}, rd.history)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "diff",
		Description: "Diff a document between a git ref and the working tree.",
	}, rd.diff)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "changed_since",
		Description: "List documents changed since a date or revision (readable ones only).",
	}, rd.changedSince)

	if rd.corpusPath != "" && rd.store != nil {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        "verify_index",
			Description: "Check whether the served index is consistent with the corpus on disk (drift/staleness).",
		}, rd.verifyIndex)
	}

	if rd.syncStatus != nil {
		mcpsdk.AddTool(srv, &mcpsdk.Tool{
			Name:        "sync_status",
			Description: "Report the notebook's git-safety state (degraded, conflicted paths, in-progress op).",
		}, rd.syncStatusTool)
	}

	return srv
}

func (rd *reader) syncStatusTool(_ context.Context, _ *mcpsdk.CallToolRequest, _ emptyIn) (*mcpsdk.CallToolResult, SyncStatus, error) {
	return nil, rd.syncStatus(), nil
}

// --- tool input/output types (schemas are inferred from these) ---

type searchIn struct {
	Query string `json:"query" jsonschema:"the full-text query"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of hits (default 20)"`
}
type searchHit struct {
	Filename string  `json:"filename"`
	Title    string  `json:"title"`
	Snippet  string  `json:"snippet"`
	Score    float64 `json:"score"`
}
type searchOut struct {
	Hits []searchHit `json:"hits"`
}

type getDocIn struct {
	Slug string `json:"slug" jsonschema:"the document slug/filename, e.g. my_note.md"`
}
type getDocOut struct {
	Found    bool           `json:"found"`
	Filename string         `json:"filename,omitempty"`
	Title    string         `json:"title,omitempty"`
	Props    map[string]any `json:"props,omitempty"`
	Body     string         `json:"body,omitempty"`
}

type emptyIn struct{}

type taxonomiesOut struct {
	Taxonomies []string `json:"taxonomies"`
}

type termsIn struct {
	Taxonomy string `json:"taxonomy" jsonschema:"the taxonomy name, e.g. customer"`
}
type termsOut struct {
	Terms []string `json:"terms"`
}

type docsByTermIn struct {
	Taxonomy string `json:"taxonomy" jsonschema:"the taxonomy name"`
	Term     string `json:"term" jsonschema:"the term within the taxonomy"`
}
type docsByTermOut struct {
	Docs []string `json:"docs"`
}

// --- handlers ---

func (rd *reader) search(_ context.Context, _ *mcpsdk.CallToolRequest, in searchIn) (*mcpsdk.CallToolResult, searchOut, error) {
	hits, err := rd.store.SearchScoped(in.Query, in.Limit, rd.scopeSQL, rd.scopeArgs)
	if err != nil {
		return nil, searchOut{}, err
	}
	out := searchOut{Hits: make([]searchHit, 0, len(hits))}
	for _, h := range hits {
		title, _ := rd.red.Redact(h.Title)
		snippet, _ := rd.red.Redact(h.Snippet)
		out.Hits = append(out.Hits, searchHit{
			Filename: h.Filename,
			Title:    title,
			Snippet:  snippet,
			Score:    h.Score,
		})
	}
	return nil, out, nil
}

func (rd *reader) getDoc(_ context.Context, _ *mcpsdk.CallToolRequest, in getDocIn) (*mcpsdk.CallToolResult, getDocOut, error) {
	doc, ok, err := rd.store.GetDocScoped(in.Slug, rd.scopeSQL, rd.scopeArgs)
	if err != nil {
		return nil, getDocOut{}, err
	}
	if !ok {
		return nil, getDocOut{Found: false}, nil // denied == missing
	}
	title, _ := rd.red.Redact(doc.Title)
	body, _ := rd.red.Redact(doc.Body)
	props, _ := rd.red.RedactValue(doc.Props).(map[string]any)
	return nil, getDocOut{
		Found:    true,
		Filename: doc.Filename,
		Title:    title,
		Props:    props,
		// Wrap the (already redacted) body in data delimiters: the corpus is
		// untrusted input, and this signals "data, not instructions".
		Body: defense.Delimit(body),
	}, nil
}

func (rd *reader) listTaxonomies(_ context.Context, _ *mcpsdk.CallToolRequest, _ emptyIn) (*mcpsdk.CallToolResult, taxonomiesOut, error) {
	tax, err := rd.store.ListTaxonomiesScoped(rd.scopeSQL, rd.scopeArgs)
	if err != nil {
		return nil, taxonomiesOut{}, err
	}
	return nil, taxonomiesOut{Taxonomies: tax}, nil
}

func (rd *reader) terms(_ context.Context, _ *mcpsdk.CallToolRequest, in termsIn) (*mcpsdk.CallToolResult, termsOut, error) {
	terms, err := rd.store.TermsForTaxonomyScoped(in.Taxonomy, rd.scopeSQL, rd.scopeArgs)
	if err != nil {
		return nil, termsOut{}, err
	}
	return nil, termsOut{Terms: terms}, nil
}

func (rd *reader) docsByTerm(_ context.Context, _ *mcpsdk.CallToolRequest, in docsByTermIn) (*mcpsdk.CallToolResult, docsByTermOut, error) {
	docs, err := rd.store.DocsByTermScoped(in.Taxonomy, in.Term, rd.scopeSQL, rd.scopeArgs)
	if err != nil {
		return nil, docsByTermOut{}, err
	}
	return nil, docsByTermOut{Docs: docs}, nil
}

// --- history tool types ---

type historyIn struct {
	Slug  string `json:"slug" jsonschema:"the document slug/filename"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of commits (default 50)"`
}
type historyOut struct {
	Found   bool             `json:"found"`
	Commits []gitsafe.Commit `json:"commits,omitempty"`
}

type diffIn struct {
	Slug string `json:"slug" jsonschema:"the document slug/filename"`
	Ref  string `json:"ref" jsonschema:"a git ref to diff against (e.g. HEAD~1); must not begin with '-'"`
}
type diffOut struct {
	Found bool   `json:"found"`
	Diff  string `json:"diff,omitempty"`
}

type changedSinceIn struct {
	Since string `json:"since" jsonschema:"a date or revision (e.g. 2024-01-01); must not begin with '-'"`
}
type changedSinceOut struct {
	Docs []string `json:"docs"`
}

// --- history handlers ---

func (rd *reader) history(_ context.Context, _ *mcpsdk.CallToolRequest, in historyIn) (*mcpsdk.CallToolResult, historyOut, error) {
	ok, err := rd.readable(in.Slug)
	if err != nil {
		return nil, historyOut{}, err
	}
	if !ok {
		return nil, historyOut{Found: false}, nil // denied == missing
	}
	commits, err := gitsafe.History(rd.corpusPath, filepath.Join(contentDir, in.Slug), in.Limit)
	if err != nil {
		return nil, historyOut{}, err
	}
	for i := range commits {
		commits[i].Subject, _ = rd.red.Redact(commits[i].Subject)
	}
	return nil, historyOut{Found: true, Commits: commits}, nil
}

func (rd *reader) diff(_ context.Context, _ *mcpsdk.CallToolRequest, in diffIn) (*mcpsdk.CallToolResult, diffOut, error) {
	ok, err := rd.readable(in.Slug)
	if err != nil {
		return nil, diffOut{}, err
	}
	if !ok {
		return nil, diffOut{Found: false}, nil
	}
	raw, err := gitsafe.Diff(rd.corpusPath, filepath.Join(contentDir, in.Slug), in.Ref)
	if err != nil {
		return nil, diffOut{}, err
	}
	redacted, _ := rd.red.Redact(raw)
	return nil, diffOut{Found: true, Diff: redacted}, nil
}

func (rd *reader) changedSince(_ context.Context, _ *mcpsdk.CallToolRequest, in changedSinceIn) (*mcpsdk.CallToolResult, changedSinceOut, error) {
	paths, err := gitsafe.ChangedSince(rd.corpusPath, in.Since)
	if err != nil {
		return nil, changedSinceOut{}, err
	}
	out := changedSinceOut{Docs: []string{}}
	seen := map[string]bool{}
	prefix := contentDir + "/"
	for _, p := range paths {
		if !strings.HasPrefix(p, prefix) || !strings.HasSuffix(p, ".md") {
			continue
		}
		slug := p[len(prefix):]
		if seen[slug] {
			continue
		}
		seen[slug] = true
		ok, err := rd.readable(slug)
		if err != nil {
			return nil, changedSinceOut{}, err
		}
		if ok {
			out.Docs = append(out.Docs, slug)
		}
	}
	return nil, out, nil
}

// --- operational: verify_index ---

type verifyIndexOut struct {
	InSync           bool     `json:"in_sync"`
	CorpusDocs       int      `json:"corpus_docs"`
	StoreDocs        int      `json:"store_docs"`
	MissingFromStore []string `json:"missing_from_store"`
	StaleInStore     []string `json:"stale_in_store"`
	Conflicted       []string `json:"conflicted,omitempty"`
}

func (rd *reader) verifyIndex(_ context.Context, _ *mcpsdk.CallToolRequest, _ emptyIn) (*mcpsdk.CallToolResult, verifyIndexOut, error) {
	g, report, err := index.Build(rd.corpusPath)
	if err != nil {
		return nil, verifyIndexOut{}, err
	}
	corpusSet := map[string]bool{}
	for name := range g.Records {
		corpusSet[name] = true
	}
	stored, err := rd.store.AllDocFilenames()
	if err != nil {
		return nil, verifyIndexOut{}, err
	}
	storeSet := map[string]bool{}
	for _, f := range stored {
		storeSet[f] = true
	}

	out := verifyIndexOut{
		CorpusDocs:       len(corpusSet),
		StoreDocs:        len(storeSet),
		MissingFromStore: []string{},
		StaleInStore:     []string{},
	}
	for name := range corpusSet {
		if !storeSet[name] {
			out.MissingFromStore = append(out.MissingFromStore, name)
		}
	}
	for _, f := range stored {
		if !corpusSet[f] {
			out.StaleInStore = append(out.StaleInStore, f)
		}
	}
	sort.Strings(out.MissingFromStore)
	sort.Strings(out.StaleInStore)
	for _, c := range report.Conflicted {
		out.Conflicted = append(out.Conflicted, c.File)
	}
	out.InSync = len(out.MissingFromStore) == 0 && len(out.StaleInStore) == 0
	return nil, out, nil
}
