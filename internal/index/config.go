package index

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// notebook config file locations, relative to the corpus root.
const (
	hugoConfigRel     = "config/_default/config.yaml"
	lindenConfigRel   = "lindenConfig"
	defaultContentDir = "content"
)

// hugoConfig is the subset of the Hugo config we read.
type hugoConfig struct {
	ContentDir string            `yaml:"contentDir"`
	Taxonomies map[string]string `yaml:"taxonomies"` // singular -> plural
}

// loadNotebook resolves the content directory and the taxonomy set for a corpus
// root. The taxonomy set is the union of the Hugo config's plural taxonomy names
// and the taxonomies declared by L1-CONF-TAX-*.yml files.
func loadNotebook(root string) (contentDir string, taxonomies []string, singular map[string]string) {
	contentDir = defaultContentDir
	taxSet := map[string]bool{}
	singular = map[string]string{} // plural -> singular

	if b, err := os.ReadFile(filepath.Join(root, hugoConfigRel)); err == nil {
		var hc hugoConfig
		if yaml.Unmarshal(b, &hc) == nil {
			if hc.ContentDir != "" {
				contentDir = hc.ContentDir
			}
			// Hugo's taxonomies map is singular -> plural.
			for sing, plural := range hc.Taxonomies {
				taxSet[plural] = true
				singular[plural] = sing
			}
		}
	}

	// Merge taxonomies declared by L1 config files.
	entries, _ := os.ReadDir(filepath.Join(root, lindenConfigRel))
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "L1-CONF-TAX-") && strings.HasSuffix(name, ".yml") {
			tax := strings.TrimSuffix(strings.TrimPrefix(name, "L1-CONF-TAX-"), ".yml")
			taxSet[tax] = true
		}
	}

	for t := range taxSet {
		taxonomies = append(taxonomies, t)
		if _, ok := singular[t]; !ok {
			singular[t] = t // no singular declared: identity
		}
	}
	sort.Strings(taxonomies)
	return contentDir, taxonomies, singular
}

// loadLindenConfig reads the L1 and L2 config objects for the given taxonomies.
// loadLindenConfig scans the lindenConfig directory and keys L1/L2 config by the
// taxonomy name as it appears in the config FILENAME — the same identifier the
// Hugo reference derives from `.Site.Data`. Callers key the L1 term-config lookup
// and the starred indexes off these (singular) names.
func loadLindenConfig(root string) (l1 map[string]map[string]any, l2 map[string]map[string]TermConfig) {
	l1 = map[string]map[string]any{}
	l2 = map[string]map[string]TermConfig{}

	dir := filepath.Join(root, lindenConfigRel)
	entries, _ := os.ReadDir(dir)

	// L1 files: L1-CONF-TAX-<tax>.yml
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "L1-CONF-TAX-") || !strings.HasSuffix(name, ".yml") {
			continue
		}
		tax := strings.TrimSuffix(strings.TrimPrefix(name, "L1-CONF-TAX-"), ".yml")
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var m map[string]any
		if yaml.Unmarshal(b, &m) == nil {
			l1[tax] = m
		}
	}

	// L2 files: L2-CONF-TAX-<tax>-TRM-<term>.yml
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "L2-CONF-TAX-") || !strings.HasSuffix(name, ".yml") {
			continue
		}
		stem := strings.TrimSuffix(strings.TrimPrefix(name, "L2-CONF-TAX-"), ".yml")
		i := strings.Index(stem, "-TRM-")
		if i < 0 {
			continue
		}
		tax := stem[:i]
		term := stem[i+len("-TRM-"):]
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var m map[string]any
		if yaml.Unmarshal(b, &m) != nil {
			continue
		}
		if l2[tax] == nil {
			l2[tax] = map[string]TermConfig{}
		}
		l2[tax][term] = TermConfig(m)
	}
	return l1, l2
}

// normalizeTerm lowercases and replaces spaces with dashes, matching the
// linny.vim on-disk term naming.
func normalizeTerm(term string) string {
	return strings.ReplaceAll(strings.ToLower(term), " ", "-")
}
