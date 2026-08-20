package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/linden-project/linny-mcp-server/internal/buildinfo"
)

// Home-level index file basenames (see docs/linden-index-spec.md §8).
const (
	fileTaxonomies        = "_index_taxonomies.json"
	fileDocsStarred       = "_index_docs_starred.json"
	fileDocsWithProps     = "_index_docs_with_props.json"
	fileDocsWithTitle     = "_index_docs_with_title.json"
	fileDocsTasksCount    = "_index_docs_tasks_count.json"
	fileIndexerInfo       = "_indexer_info.json"
	fileTaxonomiesStarred = "_index_taxonomies_starred.json"
	fileTermsStarred      = "_index_terms_starred.json"
)

// Emit writes the full index tree under indexRoot per the index-format spec.
func Emit(g *Graph, indexRoot string) error {
	if err := os.MkdirAll(indexRoot, 0o755); err != nil {
		return err
	}

	// Only taxonomies that actually have members are listed / get files, matching
	// Hugo, which only materializes taxonomies with terms.
	occurring := make([]string, 0, len(g.Taxonomies))
	for _, tax := range g.Taxonomies {
		if len(g.Members[tax]) > 0 {
			occurring = append(occurring, tax)
		}
	}
	sort.Strings(occurring)

	writes := []struct {
		name string
		v    any
	}{
		{fileTaxonomies, occurring},
		{fileDocsStarred, g.StarredDocs},
		{fileDocsWithProps, g.docsWithProps()},
		{fileDocsWithTitle, g.docsWithTitle()},
		{fileDocsTasksCount, g.docsTasksCount()},
		{fileIndexerInfo, g.indexerInfo(indexRoot)},
		{fileTaxonomiesStarred, nonNil(g.StarredTaxonomies)},
		{fileTermsStarred, nonNilTerms(g.StarredTerms)},
	}
	for _, w := range writes {
		if err := writeJSON(filepath.Join(indexRoot, w.name), w.v); err != nil {
			return err
		}
	}

	// Nested L1 / L2 files.
	for _, tax := range occurring {
		l1 := map[string]any{}
		terms := make([]string, 0, len(g.Members[tax]))
		for term := range g.Members[tax] {
			terms = append(terms, term)
		}
		sort.Strings(terms)

		for _, term := range terms {
			if cfg, ok := g.L2Config[tax][term]; ok {
				l1[term] = cfg
			} else {
				l1[term] = map[string]any{}
			}
			// L2: member list.
			l2Dir := filepath.Join(indexRoot, tax, term)
			if err := os.MkdirAll(l2Dir, 0o755); err != nil {
				return err
			}
			if err := writeJSON(filepath.Join(l2Dir, "index.json"), g.Members[tax][term]); err != nil {
				return err
			}
		}
		if err := writeJSON(filepath.Join(indexRoot, tax, "index.json"), l1); err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) docsWithProps() map[string]map[string]any {
	out := map[string]map[string]any{}
	for name, rec := range g.Records {
		if rec.Title == "" {
			continue
		}
		out[name] = rec.Props
	}
	return out
}

func (g *Graph) docsWithTitle() map[string]string {
	out := map[string]string{}
	for name, rec := range g.Records {
		if rec.Title == "" {
			continue
		}
		out[name] = rec.Title
	}
	return out
}

func (g *Graph) docsTasksCount() map[string]TaskCount {
	out := map[string]TaskCount{}
	for name, rec := range g.Records {
		if rec.Tasks.Total > 0 {
			out[name] = rec.Tasks
		}
	}
	return out
}

func (g *Graph) indexerInfo(indexRoot string) map[string]string {
	return map[string]string{
		"product_name":    "lindexer",
		"product_version": buildinfo.Version,
		"index_dir":       indexRoot,
		"content_dir":     filepath.Join(g.Root, g.ContentDir),
		"config_dir":      filepath.Join(g.Root, g.ConfigDir),
	}
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	// Atomic-ish: write to a temp file in the same dir, then rename.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func nonNilTerms(s []StarredTerm) []StarredTerm {
	if s == nil {
		return []StarredTerm{}
	}
	return s
}
