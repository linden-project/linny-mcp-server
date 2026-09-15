package index

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// The scoped query variants apply authorization as a SQL predicate: the caller
// (the authz layer) supplies a `readableSubquery` — a `SELECT filename FROM docs
// …` that yields exactly the filenames the token may read — plus its bound args.
// Filtering therefore happens in SQL, never as a post-filter, and denied
// documents are indistinguishable from missing ones.

// SearchScoped runs a full-text search restricted to readable documents.
func (s *Store) SearchScoped(query string, limit int, readableSubquery string, subArgs []any) ([]SearchHit, error) {
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	args := make([]any, 0, len(subArgs)+2)
	args = append(args, query)
	args = append(args, subArgs...)
	args = append(args, limit)

	rows, err := s.db.Query(
		`SELECT filename, title, snippet(docs_fts, 2, '[', ']', '…', 8), bm25(docs_fts)
		 FROM docs_fts
		 WHERE docs_fts MATCH ? AND docs_fts.filename IN (`+readableSubquery+`)
		 ORDER BY bm25(docs_fts)
		 LIMIT ?`, args...)
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

// DocsByTermScoped returns the readable filenames tagged (taxonomy, term).
func (s *Store) DocsByTermScoped(taxonomy, term, readableSubquery string, subArgs []any) ([]string, error) {
	args := make([]any, 0, len(subArgs)+2)
	args = append(args, taxonomy, term)
	args = append(args, subArgs...)
	return s.queryStrings(
		`SELECT filename FROM membership
		 WHERE taxonomy = ? AND term = ? AND filename IN (`+readableSubquery+`)
		 ORDER BY filename`, args...)
}

// StarredDocsScoped returns the readable documents flagged `starred: true`,
// ordered by title. Scope is applied as a predicate like every other scoped
// query, so a starred document the caller may not read is indistinguishable
// from one that does not exist: no count of the excluded rows is produced, and
// producing one would need a second, deliberately unscoped query.
func (s *Store) StarredDocsScoped(readableSubquery string, subArgs []any) ([]DocRef, error) {
	rows, err := s.db.Query(
		`SELECT filename, title FROM docs d
		 WHERE d.starred = 1 AND d.filename IN (`+readableSubquery+`)
		 ORDER BY title, filename`, subArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []DocRef
	for rows.Next() {
		var d DocRef
		if err := rows.Scan(&d.Filename, &d.Title); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ListTaxonomiesScoped returns only taxonomies that have at least one readable
// document, so a fully-denied taxonomy's existence is not leaked.
func (s *Store) ListTaxonomiesScoped(readableSubquery string, subArgs []any) ([]string, error) {
	return s.queryStrings(
		`SELECT DISTINCT m.taxonomy FROM membership m
		 WHERE m.filename IN (`+readableSubquery+`)
		 ORDER BY m.taxonomy`, subArgs...)
}

// TermsForTaxonomyScoped returns only the terms of a taxonomy that have at least
// one readable document, so a term made up solely of denied documents is not
// leaked.
func (s *Store) TermsForTaxonomyScoped(taxonomy, readableSubquery string, subArgs []any) ([]string, error) {
	args := make([]any, 0, len(subArgs)+1)
	args = append(args, taxonomy)
	args = append(args, subArgs...)
	return s.queryStrings(
		`SELECT DISTINCT term FROM membership
		 WHERE taxonomy = ? AND filename IN (`+readableSubquery+`)
		 ORDER BY term`, args...)
}

// GetDocScoped returns a document only if the scope permits reading it. A denied
// document yields ok=false — identical to a document that does not exist.
func (s *Store) GetDocScoped(filename, readableSubquery string, subArgs []any) (Doc, bool, error) {
	args := make([]any, 0, len(subArgs)+1)
	args = append(args, filename)
	args = append(args, subArgs...)

	var title, propsJSON string
	err := s.db.QueryRow(
		`SELECT title, props_json FROM docs
		 WHERE filename = ? AND filename IN (`+readableSubquery+`)`, args...).
		Scan(&title, &propsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return Doc{}, false, nil // denied == missing
	}
	if err != nil {
		return Doc{}, false, err
	}

	var body string
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
