package index

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go SQLite driver (FTS5 compiled in)
)

// Store is the SQLite+FTS5 persistence of a built Graph. It lives in stateDir
// and is a disposable cache: deleting the file and rebuilding from the corpus is
// always a valid recovery step.
type Store struct {
	db *sql.DB
}

// schemaDDL creates every table if absent. docs_fts is an FTS5 virtual table.
const schemaDDL = `
CREATE TABLE IF NOT EXISTS docs (
	filename     TEXT PRIMARY KEY,
	title        TEXT    NOT NULL DEFAULT '',
	props_json   TEXT    NOT NULL DEFAULT '{}',
	tasks_open   INTEGER NOT NULL DEFAULT 0,
	tasks_closed INTEGER NOT NULL DEFAULT 0,
	tasks_total  INTEGER NOT NULL DEFAULT 0,
	starred      INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS taxonomies (name TEXT PRIMARY KEY);
CREATE TABLE IF NOT EXISTS terms (
	taxonomy    TEXT    NOT NULL,
	term        TEXT    NOT NULL,
	config_json TEXT    NOT NULL DEFAULT '{}',
	starred     INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (taxonomy, term)
);
CREATE TABLE IF NOT EXISTS membership (
	taxonomy TEXT NOT NULL,
	term     TEXT NOT NULL,
	filename TEXT NOT NULL,
	PRIMARY KEY (taxonomy, term, filename)
);
CREATE INDEX IF NOT EXISTS idx_membership_file ON membership(filename);
CREATE INDEX IF NOT EXISTS idx_membership_term ON membership(taxonomy, term);
CREATE VIRTUAL TABLE IF NOT EXISTS docs_fts USING fts5(filename UNINDEXED, title, body);
`

// dropDDL removes every table so Populate can rebuild cleanly.
const dropDDL = `
DROP TABLE IF EXISTS docs;
DROP TABLE IF EXISTS taxonomies;
DROP TABLE IF EXISTS terms;
DROP TABLE IF EXISTS membership;
DROP TABLE IF EXISTS docs_fts;
`

// OpenStore opens (creating if needed) the SQLite store at path and ensures the
// schema exists. The connection is tuned for a single-writer disposable cache.
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// One connection: the store is written by a single indexer process, which
	// keeps FTS writes simple and avoids SQLITE_BUSY.
	db.SetMaxOpenConns(1)

	for _, pragma := range []string{
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=OFF", // disposable cache: durability not needed
		"PRAGMA journal_mode=MEMORY",
	} {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("index store: %s: %w", pragma, err)
		}
	}
	if _, err := db.Exec(schemaDDL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("index store: schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// Populate rebuilds the store from a graph in a single transaction. It drops and
// recreates the schema first, so running it again is idempotent and tolerant of
// schema drift across versions.
func (s *Store) Populate(g *Graph) (err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(dropDDL); err != nil {
		return fmt.Errorf("index store: drop: %w", err)
	}
	if _, err = tx.Exec(schemaDDL); err != nil {
		return fmt.Errorf("index store: recreate: %w", err)
	}

	if err = insertDocs(tx, g); err != nil {
		return err
	}
	if err = insertTaxonomyGraph(tx, g); err != nil {
		return err
	}
	return tx.Commit()
}

func insertDocs(tx *sql.Tx, g *Graph) error {
	docStmt, err := tx.Prepare(`INSERT INTO docs
		(filename, title, props_json, tasks_open, tasks_closed, tasks_total, starred)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = docStmt.Close() }()

	ftsStmt, err := tx.Prepare(`INSERT INTO docs_fts (filename, title, body) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = ftsStmt.Close() }()

	starred := map[string]bool{}
	for _, f := range g.StarredDocs {
		starred[f] = true
	}

	for filename, rec := range g.Records {
		props, err := json.Marshal(rec.Props)
		if err != nil {
			return err
		}
		if _, err := docStmt.Exec(filename, rec.Title, string(props),
			rec.Tasks.Open, rec.Tasks.Closed, rec.Tasks.Total, boolToInt(starred[filename])); err != nil {
			return fmt.Errorf("index store: insert doc %s: %w", filename, err)
		}
		if _, err := ftsStmt.Exec(filename, rec.Title, rec.Body); err != nil {
			return fmt.Errorf("index store: insert fts %s: %w", filename, err)
		}
	}
	return nil
}

func insertTaxonomyGraph(tx *sql.Tx, g *Graph) error {
	taxStmt, err := tx.Prepare(`INSERT OR IGNORE INTO taxonomies (name) VALUES (?)`)
	if err != nil {
		return err
	}
	defer func() { _ = taxStmt.Close() }()

	termStmt, err := tx.Prepare(`INSERT OR IGNORE INTO terms (taxonomy, term, config_json, starred) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = termStmt.Close() }()

	memStmt, err := tx.Prepare(`INSERT OR IGNORE INTO membership (taxonomy, term, filename) VALUES (?, ?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = memStmt.Close() }()

	starredTerm := map[string]bool{}
	for _, st := range g.StarredTerms {
		starredTerm[st.Taxonomy+"\x00"+st.Term] = true
	}

	for _, tax := range g.Taxonomies {
		terms := g.Members[tax]
		if len(terms) == 0 {
			continue // only occurring taxonomies are stored
		}
		if _, err := taxStmt.Exec(tax); err != nil {
			return err
		}
		for term, files := range terms {
			cfg := map[string]any{}
			if c, ok := g.L2Config[tax][term]; ok {
				cfg = c
			}
			cfgJSON, err := json.Marshal(cfg)
			if err != nil {
				return err
			}
			if _, err := termStmt.Exec(tax, term, string(cfgJSON), boolToInt(starredTerm[tax+"\x00"+term])); err != nil {
				return err
			}
			for _, f := range files {
				if _, err := memStmt.Exec(tax, term, f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
