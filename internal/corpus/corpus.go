// Package corpus deterministically generates a synthetic Linny notebook for
// tests. It never reads or references the real private corpus: every byte is
// synthesized from fixed word lists and a caller-provided seed, so the same
// Options always yields a byte-identical tree.
//
// Layout written under the target directory:
//
//	content/                      flat markdown records (the corpus)
//	lindenConfig/                 L1-CONF-TAX-*.yml, L2-CONF-TAX-*-TRM-*.yml
//	config/_default/config.yaml   Hugo config declaring the same taxonomies
//
// This matches linny-notebook-template so the reference (Hugo) indexer can
// build the same corpus for the indexer `verify` path.
package corpus

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Layout constants (relative to the target directory).
const (
	ContentDir = "content"
	ConfigDir  = "lindenConfig"
	HugoConfig = "config/_default/config.yaml"
)

// Options controls generation. The zero value is not valid; use DefaultOptions.
type Options struct {
	Seed            int64 // PRNG seed; identical seeds produce identical corpora
	Count           int   // number of normal records
	EnableEdgeCases bool  // append the deliberate edge-case records
}

// DefaultOptions returns a reasonable default (deterministic, edge cases on).
func DefaultOptions() Options {
	return Options{Seed: 1, Count: 200, EnableEdgeCases: true}
}

// taxonomy is a plural front-matter key with its candidate terms. list=true
// means the field may carry multiple terms (a list value).
type taxonomy struct {
	name  string
	terms []string
	list  bool
}

// taxonomies is fixed and ordered (never range a map when emitting — ordering
// must be deterministic).
var taxonomies = []taxonomy{
	{name: "tags", terms: []string{"note", "idea", "todo", "work", "health", "finance", "reference"}, list: true},
	{name: "projects", terms: []string{"acme", "apollo", "borealis"}},
	{name: "customer", terms: []string{"eric", "globex", "initech"}},
	{name: "type", terms: []string{"intro", "howto", "log", "spec"}},
	{name: "subject", terms: []string{"linny", "nix", "golang"}},
}

var titleWords = []string{
	"morning", "review", "notes", "meeting", "plan", "design", "backup", "server",
	"budget", "recipe", "travel", "reading", "sprint", "retro", "idea", "draft",
	"config", "network", "garden", "invoice", "contact", "roadmap", "spec", "log",
}

var bodyWords = []string{
	"The", "quick", "brown", "fox", "notes", "about", "the", "project", "and",
	"its", "taxonomy", "were", "written", "down", "carefully", "for", "later",
	"reference", "in", "the", "second", "brain", "system", "today", "again",
}

// Generate writes a synthetic corpus into dir. It creates dir if needed.
func Generate(dir string, opts Options) error {
	if opts.Count < 0 {
		return fmt.Errorf("corpus: negative count %d", opts.Count)
	}
	r := rand.New(rand.NewSource(opts.Seed))

	contentPath := filepath.Join(dir, ContentDir)
	if err := os.MkdirAll(contentPath, 0o755); err != nil {
		return err
	}

	// Track which (taxonomy, term) pairs actually occur so lindenConfig matches.
	used := map[string]map[string]bool{}
	markUsed := func(tax, term string) {
		if used[tax] == nil {
			used[tax] = map[string]bool{}
		}
		used[tax][term] = true
	}

	seenSlug := map[string]bool{}
	for i := 0; i < opts.Count; i++ {
		rec := genRecord(r, i)
		for tax, terms := range rec.taxa {
			for _, t := range terms {
				markUsed(tax, t)
			}
		}
		// Guarantee unique filenames deterministically.
		slug := rec.slug
		for seenSlug[slug] {
			slug = fmt.Sprintf("%s_%d", rec.slug, i)
		}
		seenSlug[slug] = true
		if err := os.WriteFile(filepath.Join(contentPath, slug+".md"), []byte(rec.render()), 0o644); err != nil {
			return err
		}
	}

	if opts.EnableEdgeCases {
		for _, e := range edgeRecords() {
			for tax, terms := range e.taxa {
				for _, t := range terms {
					markUsed(tax, t)
				}
			}
			name := e.slug + ".md"
			if err := os.WriteFile(filepath.Join(contentPath, name), []byte(e.renderRaw()), 0o644); err != nil {
				return err
			}
		}
		// Ensure the scope-intersection fixture exists: a doc tagged work AND health.
		markUsed("tags", "work")
		markUsed("tags", "health")
	}

	if err := writeLindenConfig(dir, used); err != nil {
		return err
	}
	return writeHugoConfig(dir)
}

// record is a normal, well-formed synthetic record.
type record struct {
	slug    string
	title   string
	crdate  string
	starred bool
	taxa    map[string][]string
	body    string
}

func genRecord(r *rand.Rand, i int) record {
	// Title: 2-3 words + index for uniqueness pressure.
	n := 2 + r.Intn(2)
	parts := make([]string, n)
	for j := range parts {
		parts[j] = titleWords[r.Intn(len(titleWords))]
	}
	title := strings.Title(strings.Join(parts, " ")) //nolint:staticcheck // deterministic ascii titles
	slug := slugify(strings.Join(parts, " "))

	// crdate: deterministic date derived from index (no wall clock).
	day := 1 + (i % 27)
	month := 1 + (i % 12)
	year := 2023 + (i % 3)
	crdate := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

	taxa := map[string][]string{}
	for _, tx := range taxonomies {
		// ~70% chance a record participates in a given taxonomy.
		if r.Intn(10) < 7 {
			if tx.list {
				k := 1 + r.Intn(2)
				seen := map[string]bool{}
				var vals []string
				for len(vals) < k {
					t := tx.terms[r.Intn(len(tx.terms))]
					if !seen[t] {
						seen[t] = true
						vals = append(vals, t)
					}
				}
				sort.Strings(vals)
				taxa[tx.name] = vals
			} else {
				taxa[tx.name] = []string{tx.terms[r.Intn(len(tx.terms))]}
			}
		}
	}
	// Every record belongs to at least one taxonomy.
	if len(taxa) == 0 {
		taxa["subject"] = []string{"linny"}
	}

	body := genBody(r)

	return record{
		slug:    slug,
		title:   title,
		crdate:  crdate,
		starred: r.Intn(10) == 0, // ~10%
		taxa:    taxa,
		body:    body,
	}
}

func genBody(r *rand.Rand) string {
	var b strings.Builder
	sentences := 2 + r.Intn(3)
	for s := 0; s < sentences; s++ {
		words := 6 + r.Intn(6)
		var line []string
		for w := 0; w < words; w++ {
			line = append(line, bodyWords[r.Intn(len(bodyWords))])
		}
		b.WriteString(strings.Join(line, " "))
		b.WriteString(".\n\n")
	}
	// ~30% of records carry a task list, so docs_tasks_count has data.
	if r.Intn(10) < 3 {
		open := r.Intn(3)
		closed := r.Intn(3)
		for k := 0; k < open; k++ {
			b.WriteString("- [ ] open task\n")
		}
		for k := 0; k < closed; k++ {
			b.WriteString("- [x] done task\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (rec record) render() string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", rec.title)
	fmt.Fprintf(&b, "crdate: %q\n", rec.crdate)
	if rec.starred {
		b.WriteString("starred: true\n")
	}
	// Emit taxonomy keys in fixed order for determinism.
	for _, tx := range taxonomies {
		vals, ok := rec.taxa[tx.name]
		if !ok {
			continue
		}
		if tx.list {
			fmt.Fprintf(&b, "%s: [%s]\n", tx.name, strings.Join(vals, ", "))
		} else {
			fmt.Fprintf(&b, "%s: %s\n", tx.name, vals[0])
		}
	}
	b.WriteString("---\n")
	b.WriteString(rec.body)
	return b.String()
}

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = strings.NewReplacer(" ", "_", "/", "_", ":", "_").Replace(s)
	return s
}
