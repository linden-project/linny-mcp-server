package mcp

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"

	"github.com/linden-project/linny-mcp-server/internal/audit"
	"github.com/linden-project/linny-mcp-server/internal/authz"
	"github.com/linden-project/linny-mcp-server/internal/defense"
	"github.com/linden-project/linny-mcp-server/internal/gitsafe"
	"github.com/linden-project/linny-mcp-server/internal/index"
	"github.com/linden-project/linny-mcp-server/internal/redact"
)

// writer holds everything the write tools need for one caller. Writes are gated
// by the git-safety guard (degraded read-only), land in quarantine by default,
// use optimistic concurrency, reindex, and are recorded to the audit log.
type writer struct {
	store      *index.Store
	red        *redact.Redactor
	scope      *authz.ScopeSet
	scopeSQL   string
	scopeArgs  []any
	corpusPath string
	guard      *gitsafe.Guard
	audit      *audit.Log
	policy     defense.Policy
	identity   string

	// meta is the corpus's taxonomy declaration, loaded on first use: only the
	// front-matter tools consult it, so create/append pay nothing for it.
	meta *index.NotebookMeta
}

// notebookMeta loads (once per request) what the corpus declares about its
// taxonomy.
func (w *writer) notebookMeta() index.NotebookMeta {
	if w.meta == nil {
		m := index.LoadNotebookMeta(w.corpusPath)
		w.meta = &m
	}
	return *w.meta
}

func newWriter(s *Server, ss *authz.ScopeSet, identity string) *writer {
	sql, args := ss.ReadableFilenamesSQL()
	return &writer{
		store: s.Store, red: s.Redactor, scope: ss, scopeSQL: sql, scopeArgs: args,
		corpusPath: s.CorpusPath, guard: s.Guard, audit: s.Audit, policy: s.Policy,
		identity: identity,
	}
}

// registerWriteTools adds the write tools to an MCP server.
func registerWriteTools(srv *mcpsdk.Server, w *writer) {
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "create_doc",
		Description: "Create a new document. Lands in the quarantine taxonomy by default.",
	}, w.createDoc)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "append_to_doc",
		Description: "Append text to an existing document's body.",
	}, w.appendToDoc)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "set_front_matter",
		Description: "Set a front-matter key on a document (order-preserving).",
	}, w.setFrontMatter)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "unset_front_matter",
		Description: "Remove a front-matter key from a document.",
	}, w.unsetFrontMatter)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        "archive",
		Description: "Archive a document (front-matter state transition, sets archived: true).",
	}, w.archive)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name: "update_doc",
		Description: "Change an existing document's body. Prefer the anchored form: " +
			"`old` must occur exactly once and is replaced by `new`. Whole-body " +
			"replacement via `body` needs a `base_hash` from get_doc and is refused " +
			"for a document whose read body was redacted. Front matter is never touched.",
	}, w.updateDoc)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name: "add_term",
		Description: "Add a term to a document's taxonomy key (idempotent). " +
			"Creates the key as a list when absent and promotes a single value to a list.",
	}, w.addTerm)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name: "remove_term",
		Description: "Remove a term from a document's taxonomy key (idempotent). " +
			"Removing the last term removes the key.",
	}, w.removeTerm)
}

// writeOut is the common result: the resulting term membership so the agent sees
// what its write actually did.
type writeOut struct {
	OK          bool                   `json:"ok"`
	Slug        string                 `json:"slug"`
	Quarantined bool                   `json:"quarantined"`
	Membership  []index.TermMembership `json:"membership"`
	// NewTerms lists terms this write coined: written, but with no declared term
	// config. Coining a term is allowed, it just should not happen by accident.
	NewTerms []string `json:"new_terms,omitempty"`
	Message  string   `json:"message,omitempty"`
}

type createDocIn struct {
	Title       string         `json:"title" jsonschema:"the document title"`
	FrontMatter map[string]any `json:"front_matter,omitempty" jsonschema:"additional front-matter keys"`
	Body        string         `json:"body,omitempty" jsonschema:"the markdown body"`
}
type appendIn struct {
	Slug string `json:"slug" jsonschema:"the document slug/filename"`
	Text string `json:"text" jsonschema:"text to append to the body"`
}
type setFMIn struct {
	Slug  string `json:"slug"`
	Key   string `json:"key"`
	Value any    `json:"value" jsonschema:"a string, bool, number, null, or a list of those"`
}
type updateDocIn struct {
	Slug string `json:"slug" jsonschema:"the document slug/filename"`
	Old  string `json:"old,omitempty" jsonschema:"anchored mode: text to replace, which must occur exactly once in the body"`
	New  string `json:"new,omitempty" jsonschema:"anchored mode: the replacement for old"`
	Body string `json:"body,omitempty" jsonschema:"whole-body mode: the complete replacement body; requires base_hash and an unredacted document"`
	// BaseHash is get_doc's content_hash. Optional when anchored (the
	// exactly-one-match rule is itself an integrity check), required to replace
	// a whole body.
	BaseHash   string `json:"base_hash,omitempty" jsonschema:"content_hash from get_doc"`
	AllowEmpty bool   `json:"allow_empty,omitempty" jsonschema:"permit the write to leave the body empty"`
}

type termIn struct {
	Slug     string `json:"slug" jsonschema:"the document slug/filename"`
	Taxonomy string `json:"taxonomy" jsonschema:"the taxonomy front-matter key"`
	Term     string `json:"term" jsonschema:"the term to add or remove"`
}
type unsetFMIn struct {
	Slug string `json:"slug"`
	Key  string `json:"key"`
}
type archiveIn struct {
	Slug string `json:"slug"`
}

func (w *writer) createDoc(_ context.Context, _ *mcpsdk.CallToolRequest, in createDocIn) (*mcpsdk.CallToolResult, writeOut, error) {
	if !w.scope.CanWriteInbox() {
		return w.deny("create_doc", "", "requires write:inbox or write:*")
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, writeOut{}, fmt.Errorf("create_doc: title is required")
	}
	slug := slugify(in.Title) + ".md"

	front := map[string]any{}
	for k, v := range in.FrontMatter {
		front[strings.ToLower(k)] = v
	}
	if _, ok := front["title"]; !ok {
		front["title"] = in.Title
	}
	w.policy.ApplyQuarantine(front) // quarantine by default

	if err := w.guard.EnsureWritable(); err != nil {
		return w.deny("create_doc", slug, err.Error())
	}
	content := renderDoc(front, in.Body)
	path := w.docPath(slug)
	// Optimistic create: must not already exist (expected hash "").
	if err := gitsafe.WriteIfUnchanged(path, []byte(content), "", 0o644); err != nil {
		return nil, writeOut{}, err
	}
	return w.finish("create_doc", slug, content, nil, true)
}

func (w *writer) appendToDoc(_ context.Context, _ *mcpsdk.CallToolRequest, in appendIn) (*mcpsdk.CallToolResult, writeOut, error) {
	raw, hash, front, ok, err := w.loadForEdit(in.Slug)
	if err != nil || !ok {
		return w.notFoundOrErr("append_to_doc", in.Slug, err)
	}
	if err := w.ensureModify(front); err != nil {
		return w.deny("append_to_doc", in.Slug, err.Error())
	}
	if err := w.guard.EnsureWritable(); err != nil {
		return w.deny("append_to_doc", in.Slug, err.Error())
	}
	newContent := strings.TrimRight(raw, "\n") + "\n\n" + in.Text + "\n"
	if err := gitsafe.WriteIfUnchanged(w.docPath(in.Slug), []byte(newContent), hash, 0o644); err != nil {
		return nil, writeOut{}, err
	}
	return w.finish("append_to_doc", in.Slug, newContent, nil, true)
}

func (w *writer) setFrontMatter(_ context.Context, _ *mcpsdk.CallToolRequest, in setFMIn) (*mcpsdk.CallToolResult, writeOut, error) {
	key := strings.ToLower(in.Key)
	node, err := valueNode(in.Value)
	if err != nil {
		return w.deny("set_front_matter", in.Slug, err.Error())
	}
	// On a declared taxonomy key the indexer reads only strings and lists of
	// strings; anything else writes cleanly and then produces no membership at
	// all. Refuse it rather than let the write look like it worked.
	terms, isTermShape := index.TermValues(in.Value)
	if w.notebookMeta().Taxonomies[key] && !isTermShape {
		return w.deny("set_front_matter", in.Slug, fmt.Sprintf(
			"%q is a taxonomy: its value must be a string or a list of strings, "+
				"otherwise the document gains no term membership", key))
	}
	coined := w.coinedTerms(key, terms)
	return w.editFM("set_front_matter", in.Slug, func(m *yaml.Node) (bool, []string, error) {
		return setMappingValue(m, key, node), coined, nil
	})
}

func (w *writer) unsetFrontMatter(_ context.Context, _ *mcpsdk.CallToolRequest, in unsetFMIn) (*mcpsdk.CallToolResult, writeOut, error) {
	return w.editFM("unset_front_matter", in.Slug, func(m *yaml.Node) (bool, []string, error) {
		return unsetMappingKey(m, strings.ToLower(in.Key)), nil, nil
	})
}

func (w *writer) archive(_ context.Context, _ *mcpsdk.CallToolRequest, in archiveIn) (*mcpsdk.CallToolResult, writeOut, error) {
	return w.editFM("archive", in.Slug, func(m *yaml.Node) (bool, []string, error) {
		return setMappingValue(m, "archived", boolNode(true)), nil, nil
	})
}

// updateDoc changes a document's body. The invariant it enforces is that an
// agent may only replace text it has genuinely seen: get_doc returns a redacted,
// delimiter-wrapped body, so the text a caller holds is not necessarily what is
// on disk. Anchored edits prove it by matching real file content and fail closed
// when they cannot; whole-body replacement has to prove it up front.
func (w *writer) updateDoc(_ context.Context, _ *mcpsdk.CallToolRequest, in updateDocIn) (*mcpsdk.CallToolResult, writeOut, error) {
	anchored, whole := in.Old != "", in.Body != ""
	if anchored == whole {
		return w.deny("update_doc", in.Slug,
			"supply either old/new (anchored, preferred) or body (whole-body), not both and not neither")
	}

	raw, hash, front, ok, err := w.loadForEdit(in.Slug)
	if err != nil || !ok {
		return w.notFoundOrErr("update_doc", in.Slug, err)
	}
	if err := w.ensureModify(front); err != nil {
		return w.deny("update_doc", in.Slug, err.Error())
	}
	if err := w.guard.EnsureWritable(); err != nil {
		return w.deny("update_doc", in.Slug, err.Error())
	}
	// Splitting and reassembling around the same front-matter text is what keeps
	// the block byte-identical: a body rewrite can never touch classification.
	fmText, body, err := splitFrontMatter(raw)
	if err != nil {
		return nil, writeOut{}, err
	}
	if in.BaseHash != "" && in.BaseHash != hash {
		return w.deny("update_doc", in.Slug,
			"stale write: the document changed since base_hash was taken; re-read it and retry")
	}

	var newBody string
	if anchored {
		newBody, err = spliceAnchored(body, in.Old, in.New)
	} else {
		err = w.allowWholeBody(in, body)
		newBody = in.Body
	}
	if err != nil {
		return w.deny("update_doc", in.Slug, err.Error())
	}
	if strings.TrimSpace(newBody) == "" && !in.AllowEmpty {
		return w.deny("update_doc", in.Slug,
			"the resulting body would be empty; pass allow_empty to do that deliberately")
	}

	newContent := "---\n" + fmText + "---\n" + newBody
	if err := gitsafe.WriteIfUnchanged(w.docPath(in.Slug), []byte(newContent), hash, 0o644); err != nil {
		return nil, writeOut{}, err
	}
	return w.finishAudit("update_doc", in.Slug, newContent, unifiedDiff(body, newBody), nil, true)
}

// spliceAnchored replaces the single occurrence of old. Neither refusal changes
// the document: an anchor that crosses redacted text cannot match on-disk
// content, so the dangerous case degrades to a refusal rather than corruption.
func spliceAnchored(body, old, replacement string) (string, error) {
	switch n := strings.Count(body, old); n {
	case 1:
		return strings.Replace(body, old, replacement, 1), nil
	case 0:
		return "", fmt.Errorf("anchor not found in the document body; re-read the document and " +
			"take the anchor from its current text (an anchor covering redacted content never matches)")
	default:
		return "", fmt.Errorf("anchor is ambiguous: %d matches; include more surrounding "+
			"context so it identifies exactly one place", n)
	}
}

// allowWholeBody reports whether a whole-body replacement is provably safe: the
// caller read this exact file, nothing was hidden from them, and they are not
// echoing the read response's framing back into the corpus.
func (w *writer) allowWholeBody(in updateDocIn, body string) error {
	if in.BaseHash == "" {
		return fmt.Errorf("whole-body replacement requires base_hash (get_doc's content_hash); " +
			"or use old/new, which does not")
	}
	if _, n := w.red.Redact(body); n > 0 {
		return fmt.Errorf("whole-body replacement is unavailable for this document: part of it is " +
			"redacted on read, so the text you hold is not what is stored; use old/new instead")
	}
	if strings.Contains(in.Body, defense.BodyBegin) || strings.Contains(in.Body, defense.BodyEnd) {
		return fmt.Errorf("the submitted body carries the data-delimiter fence from a read " +
			"response; send the body text only")
	}
	return nil
}

func (w *writer) addTerm(_ context.Context, _ *mcpsdk.CallToolRequest, in termIn) (*mcpsdk.CallToolResult, writeOut, error) {
	key, term, err := normalizeTermArgs(in)
	if err != nil {
		return nil, writeOut{}, err
	}
	coined := w.coinedTerms(key, []string{term})
	return w.editFM("add_term", in.Slug, func(m *yaml.Node) (bool, []string, error) {
		changed := addTermToMapping(m, key, term)
		if !changed {
			return false, nil, nil // already a member: nothing coined
		}
		return true, coined, nil
	})
}

func (w *writer) removeTerm(_ context.Context, _ *mcpsdk.CallToolRequest, in termIn) (*mcpsdk.CallToolResult, writeOut, error) {
	key, term, err := normalizeTermArgs(in)
	if err != nil {
		return nil, writeOut{}, err
	}
	return w.editFM("remove_term", in.Slug, func(m *yaml.Node) (bool, []string, error) {
		return removeTermFromMapping(m, key, term), nil, nil
	})
}

// normalizeTermArgs validates and lowercases the taxonomy key the way the
// indexer reads front-matter keys.
func normalizeTermArgs(in termIn) (key, term string, err error) {
	key = strings.ToLower(strings.TrimSpace(in.Taxonomy))
	term = strings.TrimSpace(in.Term)
	if key == "" || term == "" {
		return "", "", fmt.Errorf("taxonomy and term are required")
	}
	return key, term, nil
}

// coinedTerms returns the terms that are about to be written to a taxonomy key
// without a declared term config. A key that is not a taxonomy coins nothing.
func (w *writer) coinedTerms(key string, terms []string) []string {
	meta := w.notebookMeta()
	if !meta.Taxonomies[key] {
		return nil
	}
	declared := meta.DeclaredTerms[key]
	var out []string
	for _, t := range terms {
		if t != "" && !declared[index.NormalizeTerm(t)] {
			out = append(out, t)
		}
	}
	return out
}

// fmEdit applies one surgical front-matter edit. It reports whether it changed
// anything — an edit that is already satisfied writes nothing at all — and which
// terms it coined along the way.
type fmEdit func(m *yaml.Node) (changed bool, coined []string, err error)

// editFM is the shared surgical front-matter edit pipeline.
func (w *writer) editFM(tool, slug string, edit fmEdit) (*mcpsdk.CallToolResult, writeOut, error) {
	raw, hash, front, ok, err := w.loadForEdit(slug)
	if err != nil || !ok {
		return w.notFoundOrErr(tool, slug, err)
	}
	if err := w.ensureModify(front); err != nil {
		return w.deny(tool, slug, err.Error())
	}
	if err := w.guard.EnsureWritable(); err != nil {
		return w.deny(tool, slug, err.Error())
	}
	fmText, body, err := splitFrontMatter(raw)
	if err != nil {
		return nil, writeOut{}, err
	}
	var changed bool
	var coined []string
	newFM, err := editFrontMatterText(fmText, func(m *yaml.Node) error {
		var eerr error
		changed, coined, eerr = edit(m)
		return eerr
	})
	if err != nil {
		return nil, writeOut{}, err
	}
	if !changed {
		// The document already says what the caller asked for. Leave the file
		// untouched rather than re-encoding it, and skip the reindex with it.
		return w.finish(tool, slug, raw, nil, false)
	}
	newContent := "---\n" + newFM + "---\n" + body
	if err := gitsafe.WriteIfUnchanged(w.docPath(slug), []byte(newContent), hash, 0o644); err != nil {
		return nil, writeOut{}, err
	}
	return w.finish(tool, slug, newContent, coined, true)
}

// ensureModify checks write permission for modifying an existing document:
// write:* always allows; write:inbox allows only quarantined (agent-draft) docs.
func (w *writer) ensureModify(front map[string]any) error {
	if w.scope.CanWriteAll() {
		return nil
	}
	if w.scope.CanWriteInbox() && w.policy.IsQuarantined(front) {
		return nil
	}
	return fmt.Errorf("requires write:* (or write:inbox for a quarantined draft)")
}

// loadForEdit returns the raw file, its content hash, and parsed front matter,
// but only if the caller may read the document (denied == not-found).
func (w *writer) loadForEdit(slug string) (raw, hash string, front map[string]any, ok bool, err error) {
	if _, readable, gerr := w.store.GetDocScoped(slug, w.scopeSQL, w.scopeArgs); gerr != nil || !readable {
		return "", "", nil, false, gerr
	}
	path := w.docPath(slug)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", "", nil, false, err
	}
	h, err := gitsafe.HashFile(path)
	if err != nil {
		return "", "", nil, false, err
	}
	fmText, _, err := splitFrontMatter(string(b))
	front = map[string]any{}
	if err == nil {
		_ = yaml.Unmarshal([]byte(fmText), &front)
	}
	return string(b), h, front, true, nil
}

// finish reindexes, records the audit entry, and returns the resulting
// membership. When wrote is false the corpus is untouched, so there is nothing
// to reindex and nothing to record as a diff.
func (w *writer) finish(tool, slug, content string, coined []string, wrote bool) (*mcpsdk.CallToolResult, writeOut, error) {
	diff := ""
	if wrote {
		diff = content
	}
	return w.finishAudit(tool, slug, content, diff, coined, wrote)
}

// finishAudit is finish with an explicit audit payload, so a tool that can
// record what it changed does not store a second copy of the document.
func (w *writer) finishAudit(tool, slug, content, diff string, coined []string, wrote bool) (*mcpsdk.CallToolResult, writeOut, error) {
	if wrote {
		if err := w.reindex(); err != nil {
			return nil, writeOut{}, err
		}
	}
	membership, err := w.store.TermsOfDoc(slug)
	if err != nil {
		return nil, writeOut{}, err
	}
	quarantined := false
	if fmText, _, serr := splitFrontMatter(content); serr == nil {
		var front map[string]any
		if yaml.Unmarshal([]byte(fmText), &front) == nil {
			quarantined = w.policy.IsQuarantined(front)
		}
	}
	w.log(tool, slug, diff, "ok")
	return nil, writeOut{
		OK: true, Slug: slug, Quarantined: quarantined,
		Membership: membership, NewTerms: coined,
	}, nil
}

func (w *writer) reindex() error {
	g, _, err := index.Build(w.corpusPath)
	if err != nil {
		return err
	}
	return w.store.Populate(g)
}

func (w *writer) docPath(slug string) string {
	return filepath.Join(w.corpusPath, contentDir, slug)
}

// deny records a denied/refused write and returns a soft result (OK=false).
func (w *writer) deny(tool, slug, msg string) (*mcpsdk.CallToolResult, writeOut, error) {
	w.log(tool, slug, "", "denied")
	return nil, writeOut{OK: false, Slug: slug, Message: msg}, nil
}

func (w *writer) notFoundOrErr(tool, slug string, err error) (*mcpsdk.CallToolResult, writeOut, error) {
	if err != nil {
		w.log(tool, slug, "", "error")
		return nil, writeOut{}, err
	}
	w.log(tool, slug, "", "denied")
	return nil, writeOut{OK: false, Slug: slug, Message: "not found"}, nil
}

func (w *writer) log(tool, slug, diff, outcome string) {
	if w.audit == nil {
		return
	}
	_ = w.audit.Append(audit.Entry{Identity: w.identity, Tool: tool, Slug: slug, Diff: diff, Outcome: outcome})
}

// --- rendering / front-matter helpers ---

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	return strings.NewReplacer(" ", "_", "/", "_", ":", "_").Replace(s)
}

// renderDoc renders a new document. yaml.Marshal sorts map keys, giving
// deterministic output for freshly created docs.
func renderDoc(front map[string]any, body string) string {
	b, _ := yaml.Marshal(front)
	return "---\n" + string(b) + "---\n" + body
}

// splitFrontMatter separates a leading ----fenced YAML block from the body.
func splitFrontMatter(content string) (fm, body string, err error) {
	if !strings.HasPrefix(content, "---\n") {
		return "", content, fmt.Errorf("no front matter")
	}
	rest := content[4:]
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		// closing fence may be at EOF without trailing newline
		if strings.HasSuffix(rest, "\n---") {
			return rest[:len(rest)-4], "", nil
		}
		return "", content, fmt.Errorf("unterminated front matter")
	}
	return rest[:idx+1], rest[idx+len("\n---\n"):], nil
}

// editFrontMatterText applies edit to the front matter's mapping node and
// re-encodes it, preserving key order and comments.
func editFrontMatterText(fmText string, edit func(*yaml.Node) error) (string, error) {
	var root yaml.Node
	if strings.TrimSpace(fmText) == "" {
		root = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	} else if err := yaml.Unmarshal([]byte(fmText), &root); err != nil {
		return "", err
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return "", fmt.Errorf("front matter is not a mapping")
	}
	if err := edit(root.Content[0]); err != nil {
		return "", err
	}
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(root.Content[0]); err != nil {
		return "", err
	}
	_ = enc.Close()
	return b.String(), nil
}

// setMappingValue sets key to v, reporting whether that changed anything. An
// identical scalar is left alone, which is what makes `archive` idempotent.
func setMappingValue(m *yaml.Node, key string, v *yaml.Node) bool {
	if i := findMappingValue(m, key); i >= 0 {
		if sameScalar(m.Content[i], v) {
			return false
		}
		m.Content[i] = v
		return true
	}
	m.Content = append(m.Content, strNode(key), v)
	return true
}

// unsetMappingKey removes key, reporting whether it was there.
func unsetMappingKey(m *yaml.Node, key string) bool {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return true
		}
	}
	return false
}

// findMappingValue returns the index of key's VALUE node, or -1.
func findMappingValue(m *yaml.Node, key string) int {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return i + 1
		}
	}
	return -1
}

func sameScalar(a, b *yaml.Node) bool {
	return a.Kind == yaml.ScalarNode && b.Kind == yaml.ScalarNode &&
		a.Tag == b.Tag && a.Value == b.Value
}

func strNode(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
}

func boolNode(b bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(b)}
}

// valueNode renders a caller-supplied value as a YAML node, preserving the type
// the caller sent rather than stringifying it. Types are taken from the caller
// and never inferred from a string's contents: a string that looks like a date
// stays a string. Linny front matter is flat, so mappings and nested lists are
// refused rather than silently flattened.
func valueNode(v any) (*yaml.Node, error) {
	switch t := v.(type) {
	case nil:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
	case bool:
		return boolNode(t), nil
	case string:
		return strNode(t), nil
	case float64:
		return numberNode(t), nil
	case float32:
		return numberNode(float64(t)), nil
	case int:
		return intNode(int64(t)), nil
	case int64:
		return intNode(t), nil
	case []any:
		seq := &yaml.Node{Kind: yaml.SequenceNode} // zero Style == block style
		for _, e := range t {
			n, err := valueNode(e)
			if err != nil {
				return nil, err
			}
			if n.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("list elements must be scalars; nested lists are not front matter")
			}
			seq.Content = append(seq.Content, n)
		}
		return seq, nil
	default:
		return nil, fmt.Errorf("unsupported front-matter value of type %T; "+
			"Linny front matter holds scalars and lists of scalars", v)
	}
}

// numberNode resolves JSON's single number type: an integral value within int64
// range is an integer, anything else a float.
func numberNode(f float64) *yaml.Node {
	if !math.IsInf(f, 0) && !math.IsNaN(f) && f == math.Trunc(f) &&
		f >= math.MinInt64 && f <= math.MaxInt64 {
		return intNode(int64(f))
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(f, 'g', -1, 64)}
}

func intNode(i int64) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(i, 10)}
}

// addTermToMapping adds term to key's value, reporting whether it changed
// anything. An absent key becomes a sequence; a single existing value is
// promoted to one. Membership is compared the way the indexer will read it, so
// adding "Acme" to a document already carrying "acme" is a no-op.
func addTermToMapping(m *yaml.Node, key, term string) bool {
	want := index.NormalizeTerm(term)
	i := findMappingValue(m, key)
	if i < 0 {
		m.Content = append(m.Content, strNode(key), seqNode(term))
		return true
	}
	switch v := m.Content[i]; v.Kind {
	case yaml.SequenceNode:
		for _, e := range v.Content {
			if index.NormalizeTerm(e.Value) == want {
				return false
			}
		}
		v.Content = append(v.Content, strNode(term))
		return true
	case yaml.ScalarNode:
		// An empty or null value holds no term: replace it outright.
		if v.Tag == "!!null" || v.Value == "" {
			m.Content[i] = seqNode(term)
			return true
		}
		if index.NormalizeTerm(v.Value) == want {
			return false
		}
		m.Content[i] = seqNode(v.Value, term)
		return true
	default:
		return false
	}
}

// removeTermFromMapping removes term from key's value, reporting whether it
// changed anything. Removing the last term removes the key rather than leaving
// an empty sequence behind.
func removeTermFromMapping(m *yaml.Node, key, term string) bool {
	want := index.NormalizeTerm(term)
	i := findMappingValue(m, key)
	if i < 0 {
		return false
	}
	switch v := m.Content[i]; v.Kind {
	case yaml.ScalarNode:
		if index.NormalizeTerm(v.Value) != want {
			return false
		}
		return unsetMappingKey(m, key)
	case yaml.SequenceNode:
		kept := make([]*yaml.Node, 0, len(v.Content))
		for _, e := range v.Content {
			if index.NormalizeTerm(e.Value) != want {
				kept = append(kept, e)
			}
		}
		if len(kept) == len(v.Content) {
			return false
		}
		if len(kept) == 0 {
			return unsetMappingKey(m, key)
		}
		v.Content = kept
		return true
	default:
		return false
	}
}

func seqNode(terms ...string) *yaml.Node {
	n := &yaml.Node{Kind: yaml.SequenceNode}
	for _, t := range terms {
		n.Content = append(n.Content, strNode(t))
	}
	return n
}
