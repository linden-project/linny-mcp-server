package corpus

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// writeLindenConfig emits L1 (per-taxonomy) and L2 (per-term) config files for
// exactly the taxonomies/terms that occur in the generated content. Some are
// marked starred deterministically so the starred indexes have data.
func writeLindenConfig(dir string, used map[string]map[string]bool) error {
	cfgPath := filepath.Join(dir, ConfigDir)
	if err := os.MkdirAll(cfgPath, 0o755); err != nil {
		return err
	}

	taxNames := make([]string, 0, len(used))
	for t := range used {
		taxNames = append(taxNames, t)
	}
	sort.Strings(taxNames)

	for _, tax := range taxNames {
		// Config files are named by the SINGULAR taxonomy name — the key the Hugo
		// reference (`.Data.Singular`) uses to resolve L2-CONF, so the L1 config
		// index and the starred indexes resolve consistently.
		sing := singularOf(tax)
		// Deterministic "starred" taxonomy: the first alphabetically.
		l1Starred := tax == taxNames[0]
		l1 := fmt.Sprintf("title: %s\ninfotext: About %s\nstarred: %t\n",
			strings.Title(tax), tax, l1Starred) //nolint:staticcheck // ascii
		if err := os.WriteFile(filepath.Join(cfgPath, "L1-CONF-TAX-"+sing+".yml"), []byte(l1), 0o644); err != nil {
			return err
		}

		terms := make([]string, 0, len(used[tax]))
		for t := range used[tax] {
			terms = append(terms, t)
		}
		sort.Strings(terms)
		for i, term := range terms {
			// Deterministic "starred" term: every third term.
			l2Starred := i%3 == 0
			l2 := fmt.Sprintf("title: %s\ninfotext: About %s in %s\nstarred: %t\narchive: false\n",
				strings.Title(term), term, tax, l2Starred) //nolint:staticcheck // ascii
			fname := "L2-CONF-TAX-" + sing + "-TRM-" + normalizeTerm(term) + ".yml"
			if err := os.WriteFile(filepath.Join(cfgPath, fname), []byte(l2), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// normalizeTerm lowercases and replaces spaces with dashes, matching the
// linny.vim term normalization used for on-disk names.
func normalizeTerm(term string) string {
	return strings.ReplaceAll(strings.ToLower(term), " ", "-")
}

// writeHugoConfig emits a Hugo config declaring the same taxonomies, so the
// reference indexer can build this corpus for the `verify` path. The
// singular->plural map uses the plural for both when no natural singular
// exists (matching linny-notebook-template).
func writeHugoConfig(dir string) error {
	p := filepath.Join(dir, HugoConfig)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("title: \"Synthetic Linny Notebook\"\n")
	b.WriteString("disableKinds: [\"RSS\", \"section\", \"sitemap\", \"robotsTXT\", \"404\"]\n\n")
	b.WriteString("contentDir: content\n")
	b.WriteString("dataDir: lindenConfig\n")
	b.WriteString("publishDir: lindenIndex\n\n")
	b.WriteString("taxonomies:\n")
	// singular: plural, derived from the taxonomy table so the Hugo config, the
	// config filenames, and the indexer all agree on the singular↔plural mapping.
	for _, tx := range taxonomies {
		fmt.Fprintf(&b, "  %s: %q\n", singularOf(tx.name), tx.name)
	}
	return os.WriteFile(p, []byte(b.String()), 0o644)
}
