package index

// Record is one parsed markdown record from the corpus.
type Record struct {
	// Filename is the basename including ".md" — the identity key used across
	// all index files.
	Filename string
	// Props is the full front matter with keys lowercased. Nil if the record had
	// no (or unparseable) front matter.
	Props map[string]any
	// Body is the markdown after the front matter.
	Body string
	// Tasks holds the Markdown task-list counts.
	Tasks TaskCount
	// Title is the record's title (from Props["title"]), empty if absent.
	Title string
}

// TaskCount holds Markdown task-list counts for a record.
type TaskCount struct {
	Open   int `json:"open"`
	Closed int `json:"closed"`
	Total  int `json:"total"`
}

// TermConfig is a term's L2 configuration object, embedded verbatim into the L1
// index. It is an untyped map because clients read arbitrary keys (title,
// infotext, starred, views, ...).
type TermConfig map[string]any

// Graph is the built taxonomy graph plus everything needed to emit the index.
type Graph struct {
	// Root is the corpus root; ContentDir/ConfigDir are the resolved dirs,
	// recorded so _indexer_info.json can report real paths.
	Root       string
	ContentDir string
	ConfigDir  string
	// Taxonomies is the ordered set of taxonomy (plural) names.
	Taxonomies []string
	// Singular maps each taxonomy's plural name to its singular. The L1
	// term-config index is looked up by the singular name, matching Hugo's
	// `.Data.Singular`-keyed L2-CONF lookup.
	Singular map[string]string
	// Members maps taxonomy -> term -> set of member filenames.
	Members map[string]map[string][]string
	// Records is every successfully parsed record, keyed by filename.
	Records map[string]*Record
	// StarredDocs is the filenames of records with starred:true.
	StarredDocs []string
	// L1Config maps taxonomy -> its L1 config (title/infotext/starred/...).
	L1Config map[string]map[string]any
	// L2Config maps taxonomy -> term -> its L2 config object.
	L2Config map[string]map[string]TermConfig
	// StarredTaxonomies and StarredTerms are derived from lindenConfig.
	StarredTaxonomies []string
	StarredTerms      []StarredTerm
}

// StarredTerm is an entry of _index_terms_starred.json.
type StarredTerm struct {
	Taxonomy string `json:"taxonomy"`
	Term     string `json:"term"`
}

// BuildReport records non-fatal issues surfaced during a build.
type BuildReport struct {
	// Malformed lists records whose front matter could not be parsed.
	Malformed []Issue
	// Conflicted lists records containing committed git conflict markers.
	Conflicted []Issue
	// RecordCount is the number of successfully indexed records.
	RecordCount int
}

// Issue is a per-file build note.
type Issue struct {
	File   string
	Detail string
}

// HasProblems reports whether the build surfaced conflict markers (the
// load-bearing signal for git-safety degraded mode).
func (r *BuildReport) HasProblems() bool {
	return len(r.Conflicted) > 0
}
