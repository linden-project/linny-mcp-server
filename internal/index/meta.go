package index

// NotebookMeta describes what a corpus declares about its taxonomy: which
// front-matter keys produce term membership, and which terms each of those keys
// declares through an L2 config file. The write path needs both — the first to
// reject values the indexer would silently ignore, the second to report a term
// that is being coined rather than reused.
type NotebookMeta struct {
	// Taxonomies holds the front-matter keys that Build treats as taxonomies.
	Taxonomies map[string]bool
	// DeclaredTerms maps such a key to its declared terms, in normalized form.
	DeclaredTerms map[string]map[string]bool
}

// LoadNotebookMeta reads a corpus's taxonomy declarations. It is best-effort in
// the same way loadNotebook is: a corpus without config files yields empty maps
// rather than an error, so a write path can always consult it.
func LoadNotebookMeta(root string) NotebookMeta {
	_, taxonomies, singular := loadNotebook(root)
	_, l2 := loadLindenConfig(root)

	m := NotebookMeta{
		Taxonomies:    make(map[string]bool, len(taxonomies)),
		DeclaredTerms: make(map[string]map[string]bool, len(taxonomies)),
	}
	for _, tax := range taxonomies {
		m.Taxonomies[tax] = true
		// Front-matter keys are the (plural) taxonomy names, while L2 config
		// files are keyed by the singular name — see the lindenConfig filename
		// convention. Map across before looking the terms up.
		terms := map[string]bool{}
		for term := range l2[singular[tax]] {
			terms[normalizeTerm(term)] = true
		}
		m.DeclaredTerms[tax] = terms
	}
	return m
}

// NormalizeTerm exposes the term normalization Build applies, so callers can
// compare a term against existing membership the same way the indexer will.
func NormalizeTerm(term string) string { return normalizeTerm(term) }

// TermValues reports the term strings a front-matter value contributes, and
// whether the value has a shape the indexer reads at all. It mirrors termsOf:
// only a string or a sequence of strings produces membership.
func TermValues(v any) (terms []string, ok bool) {
	switch t := v.(type) {
	case string:
		return []string{t}, true
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			s, isStr := e.(string)
			if !isStr {
				return nil, false
			}
			out = append(out, s)
		}
		return out, true
	default:
		return nil, false
	}
}
