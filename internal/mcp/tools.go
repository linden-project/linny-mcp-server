package mcp

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

// reader holds everything a read tool needs, bound to one caller's scope. The
// scope is pre-compiled to a SQL subquery (deny-by-default) so every query
// filters in SQL; content fields are passed through the redactor on the way out.
type reader struct {
	store     *index.Store
	red       *redact.Redactor
	scopeSQL  string
	scopeArgs []any
}

// newReader builds a reader for a scope set.
func newReader(store *index.Store, red *redact.Redactor, ss *authz.ScopeSet) *reader {
	sql, args := ss.ReadableFilenamesSQL()
	return &reader{store: store, red: red, scopeSQL: sql, scopeArgs: args}
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

	return srv
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
		Body:     body,
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
