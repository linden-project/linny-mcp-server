package index

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// SearchHit is one ranked full-text search result.
type SearchHit struct {
	Filename string  `json:"filename"`
	Title    string  `json:"title"`
	Snippet  string  `json:"snippet"`
	Score    float64 `json:"score"` // FTS5 bm25; lower is more relevant
}

// Doc is a single document as returned by GetDoc.
type Doc struct {
	Filename string         `json:"filename"`
	Title    string         `json:"title"`
	Props    map[string]any `json:"props"`
	Body     string         `json:"body"`
}

const defaultSearchLimit = 20

// Search runs a full-text query over titles and bodies, returning results ranked
// by FTS5 bm25 (best first) with a snippet from the body. A query that matches
// nothing returns an empty slice and no error.
func (s *Store) Search(query string, limit int) ([]SearchHit, error) {
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	rows, err := s.db.Query(
		`SELECT filename, title, snippet(docs_fts, 2, '[', ']', '…', 8), bm25(docs_fts)
		 FROM docs_fts
		 WHERE docs_fts MATCH ?
		 ORDER BY bm25(docs_fts)
		 LIMIT ?`, query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var hits []SearchHit
	for rows.Next() {
		var h SearchHit
		if err := rows.Scan(&h.Filename, &h.Title, &h.Snippet, &h.Score); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

// ListTaxonomies returns the taxonomy names that occur in the index, sorted.
func (s *Store) ListTaxonomies() ([]string, error) {
	return s.queryStrings(`SELECT name FROM taxonomies ORDER BY name`)
}

// TermsForTaxonomy returns the terms of a taxonomy, sorted.
func (s *Store) TermsForTaxonomy(taxonomy string) ([]string, error) {
	return s.queryStrings(`SELECT term FROM terms WHERE taxonomy = ? ORDER BY term`, taxonomy)
}

// DocsByTerm returns the filenames tagged with (taxonomy, term), sorted.
func (s *Store) DocsByTerm(taxonomy, term string) ([]string, error) {
	return s.queryStrings(
		`SELECT filename FROM membership WHERE taxonomy = ? AND term = ? ORDER BY filename`,
		taxonomy, term)
}

// GetDoc returns a document by filename. ok is false when it is not indexed.
func (s *Store) GetDoc(filename string) (doc Doc, ok bool, err error) {
	var title, propsJSON string
	err = s.db.QueryRow(`SELECT title, props_json FROM docs WHERE filename = ?`, filename).
		Scan(&title, &propsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Doc{}, false, nil
	}
	if err != nil {
		return Doc{}, false, err
	}

	var body string
	// Body lives in the FTS table; ignore a missing row defensively.
	if err := s.db.QueryRow(`SELECT body FROM docs_fts WHERE filename = ?`, filename).Scan(&body); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Doc{}, false, err
	}

	props := map[string]any{}
	if propsJSON != "" {
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return Doc{}, false, err
		}
	}
	return Doc{Filename: filename, Title: title, Props: props, Body: body}, true, nil
}

func (s *Store) queryStrings(query string, args ...any) ([]string, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []string{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
