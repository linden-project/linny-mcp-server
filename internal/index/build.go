package index

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Build performs a full rebuild of the taxonomy graph from a corpus root
// (a directory containing the content dir and lindenConfig). It never aborts on
// a single bad record: malformed front matter and conflict markers are recorded
// in the returned BuildReport.
func Build(root string) (*Graph, *BuildReport, error) {
	contentDir, taxonomies, singular := loadNotebook(root)
	l1, l2 := loadLindenConfig(root)

	g := &Graph{
		Root:       root,
		ContentDir: contentDir,
		ConfigDir:  lindenConfigRel,
		Taxonomies: taxonomies,
		Singular:   singular,
		Members:    map[string]map[string][]string{},
		Records:    map[string]*Record{},
		L1Config:   l1,
		L2Config:   l2,
	}
	report := &BuildReport{}

	taxSet := map[string]bool{}
	for _, t := range taxonomies {
		taxSet[t] = true
	}

	contentPath := filepath.Join(root, contentDir)
	entries, err := os.ReadDir(contentPath)
	if err != nil {
		return nil, nil, fmt.Errorf("index: reading content dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(contentPath, e.Name()))
		if err != nil {
			return nil, nil, err
		}
		content := string(raw)

		if hits := conflictLines(content); len(hits) > 0 {
			report.Conflicted = append(report.Conflicted, Issue{
				File:   e.Name(),
				Detail: fmt.Sprintf("%d conflict marker line(s), first: %q", len(hits), hits[0]),
			})
		}

		rec, err := parseRecord(e.Name(), content)
		if err != nil {
			report.Malformed = append(report.Malformed, Issue{File: e.Name(), Detail: err.Error()})
			continue
		}
		g.Records[e.Name()] = rec
		report.RecordCount++

		if b, ok := rec.Props["starred"].(bool); ok && b {
			g.StarredDocs = append(g.StarredDocs, e.Name())
		}

		for key, val := range rec.Props {
			if !taxSet[key] {
				continue
			}
			for _, term := range termsOf(val) {
				addMember(g.Members, key, term, e.Name())
			}
		}
	}

	finalize(g)
	return g, report, nil
}

// termsOf extracts normalized term strings from a front-matter value.
func termsOf(v any) []string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{normalizeTerm(t)}
	case []any:
		var out []string
		for _, e := range t {
			if s, ok := e.(string); ok && s != "" {
				out = append(out, normalizeTerm(s))
			}
		}
		return out
	default:
		return nil
	}
}

func addMember(members map[string]map[string][]string, tax, term, file string) {
	if members[tax] == nil {
		members[tax] = map[string][]string{}
	}
	members[tax][term] = append(members[tax][term], file)
}

// finalize sorts/dedups membership and derives the starred index sets, so all
// output is deterministic.
func finalize(g *Graph) {
	for _, terms := range g.Members {
		for term, files := range terms {
			terms[term] = sortedUnique(files)
		}
	}
	g.StarredDocs = sortedUnique(g.StarredDocs)

	// Starred taxonomies: every L1-CONF with starred:true. Derived from the config
	// filename identifiers (the singular name), matching Hugo's `.Site.Data` scan —
	// unified with the L1 term-config lookup convention (§9.1).
	for tax, cfg := range g.L1Config {
		if b, ok := cfg["starred"].(bool); ok && b {
			g.StarredTaxonomies = append(g.StarredTaxonomies, tax)
		}
	}
	sort.Strings(g.StarredTaxonomies)

	// Starred terms: every L2-CONF with starred:true (no occurrence filter — Hugo
	// has none), keyed by the config-filename (singular) taxonomy name.
	for tax, terms := range g.L2Config {
		for term, cfg := range terms {
			if b, ok := cfg["starred"].(bool); ok && b {
				g.StarredTerms = append(g.StarredTerms, StarredTerm{Taxonomy: tax, Term: term})
			}
		}
	}
	sort.Slice(g.StarredTerms, func(i, j int) bool {
		if g.StarredTerms[i].Taxonomy != g.StarredTerms[j].Taxonomy {
			return g.StarredTerms[i].Taxonomy < g.StarredTerms[j].Taxonomy
		}
		return g.StarredTerms[i].Term < g.StarredTerms[j].Term
	})
}

func sortedUnique(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
