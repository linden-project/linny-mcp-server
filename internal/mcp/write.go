package mcp

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
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
}

// writeOut is the common result: the resulting term membership so the agent sees
// what its write actually did.
type writeOut struct {
	OK          bool                   `json:"ok"`
	Slug        string                 `json:"slug"`
	Quarantined bool                   `json:"quarantined"`
	Membership  []index.TermMembership `json:"membership"`
	Message     string                 `json:"message,omitempty"`
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
	Value any    `json:"value" jsonschema:"scalar value (string, bool, or number)"`
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
	return w.finish("create_doc", slug, content)
}

func (w *writer) appendToDoc(_ context.Context, _ *mcpsdk.CallToolRequest, in appendIn) (*mcpsdk.CallToolResult, writeOut, error) {
	raw, hash, front, ok, err := w.loadForEdit(in.Slug)
	if err != nil || !ok {
		return w.notFoundOrErr("append_to_doc", in.Slug, err)
	}
	if err := w.ensureModify(front); err != nil {
		return w.deny("append_to_doc", in.Slug, err.Error())
	}
	newContent := strings.TrimRight(raw, "\n") + "\n\n" + in.Text + "\n"
	if err := gitsafe.WriteIfUnchanged(w.docPath(in.Slug), []byte(newContent), hash, 0o644); err != nil {
		return nil, writeOut{}, err
	}
	return w.finish("append_to_doc", in.Slug, newContent)
}

func (w *writer) setFrontMatter(_ context.Context, _ *mcpsdk.CallToolRequest, in setFMIn) (*mcpsdk.CallToolResult, writeOut, error) {
	return w.editFM("set_front_matter", in.Slug, func(m *yaml.Node) error {
		setMappingScalar(m, strings.ToLower(in.Key), in.Value)
		return nil
	})
}

func (w *writer) unsetFrontMatter(_ context.Context, _ *mcpsdk.CallToolRequest, in unsetFMIn) (*mcpsdk.CallToolResult, writeOut, error) {
	return w.editFM("unset_front_matter", in.Slug, func(m *yaml.Node) error {
		unsetMappingKey(m, strings.ToLower(in.Key))
		return nil
	})
}

func (w *writer) archive(_ context.Context, _ *mcpsdk.CallToolRequest, in archiveIn) (*mcpsdk.CallToolResult, writeOut, error) {
	return w.editFM("archive", in.Slug, func(m *yaml.Node) error {
		setMappingScalar(m, "archived", true)
		return nil
	})
}

// editFM is the shared surgical front-matter edit pipeline.
func (w *writer) editFM(tool, slug string, edit func(*yaml.Node) error) (*mcpsdk.CallToolResult, writeOut, error) {
	raw, hash, front, ok, err := w.loadForEdit(slug)
	if err != nil || !ok {
		return w.notFoundOrErr(tool, slug, err)
	}
	if err := w.ensureModify(front); err != nil {
		return w.deny(tool, slug, err.Error())
	}
	fmText, body, err := splitFrontMatter(raw)
	if err != nil {
		return nil, writeOut{}, err
	}
	newFM, err := editFrontMatterText(fmText, edit)
	if err != nil {
		return nil, writeOut{}, err
	}
	newContent := "---\n" + newFM + "---\n" + body
	if err := gitsafe.WriteIfUnchanged(w.docPath(slug), []byte(newContent), hash, 0o644); err != nil {
		return nil, writeOut{}, err
	}
	return w.finish(tool, slug, newContent)
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

// finish reindexes, records the audit entry, and returns the resulting membership.
func (w *writer) finish(tool, slug, content string) (*mcpsdk.CallToolResult, writeOut, error) {
	if err := w.reindex(); err != nil {
		return nil, writeOut{}, err
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
	w.log(tool, slug, content, "ok")
	return nil, writeOut{OK: true, Slug: slug, Quarantined: quarantined, Membership: membership}, nil
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

func setMappingScalar(m *yaml.Node, key string, value any) {
	v := scalarNode(value)
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = v
			return
		}
	}
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, v)
}

func unsetMappingKey(m *yaml.Node, key string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return
		}
	}
}

func scalarNode(value any) *yaml.Node {
	switch v := value.(type) {
	case bool:
		if v {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"}
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"}
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprintf("%v", v)}
	}
}
