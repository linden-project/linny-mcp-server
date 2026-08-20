// Package authz implements deny-by-default scopes with SQL-level filtering.
// Deny rules are evaluated across ALL of a document's terms (intersection
// semantics), and existence is never leaked to the caller.
package authz
