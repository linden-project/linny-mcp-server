// Package authz implements deny-by-default scopes compiled to SQL. Deny rules
// are evaluated across ALL of a document's terms (intersection semantics), and
// authorization is applied as a query-time SQL predicate — never a post-filter —
// so document existence is never leaked to an unauthorized caller.
package authz

import (
	"fmt"
	"strings"
)

// Action is the operation a rule governs.
type Action int

const (
	ActionRead Action = iota
	ActionWrite
	ActionDelete
)

// selector matches documents by membership. all=true matches every document;
// otherwise it matches documents that are members of taxonomy (and term, when
// set). A named target (e.g. "inbox") is stored in taxonomy for write scopes.
type selector struct {
	all      bool
	taxonomy string
	term     string
}

// rule is one parsed scope.
type rule struct {
	action Action
	deny   bool
	sel    selector
}

// ScopeSet is the compiled set of scopes attached to a token.
type ScopeSet struct {
	rules []rule
}

// Parse turns scope strings into a ScopeSet, rejecting unknown scopes.
func Parse(scopes []string) (*ScopeSet, error) {
	ss := &ScopeSet{}
	for _, s := range scopes {
		r, err := parseScope(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		ss.rules = append(ss.rules, r)
	}
	return ss, nil
}

func parseScope(s string) (rule, error) {
	if s == "" {
		return rule{}, fmt.Errorf("authz: empty scope")
	}
	parts := strings.Split(s, ":")
	head := parts[0]

	var r rule
	switch head {
	case "read":
		r.action = ActionRead
	case "write":
		r.action = ActionWrite
	case "delete":
		r.action = ActionDelete
	case "deny":
		// deny governs visibility (read); a doc you cannot read you cannot act on.
		r.action = ActionRead
		r.deny = true
	default:
		return rule{}, fmt.Errorf("authz: unknown scope %q", s)
	}

	sel, err := parseSelector(parts[1:])
	if err != nil {
		return rule{}, fmt.Errorf("authz: scope %q: %w", s, err)
	}
	r.sel = sel
	return r, nil
}

func parseSelector(segs []string) (selector, error) {
	switch {
	case len(segs) == 1 && segs[0] == "*":
		return selector{all: true}, nil
	case len(segs) == 1 && segs[0] == "inbox":
		// write:inbox convention — a named quarantine target.
		return selector{taxonomy: "inbox"}, nil
	case len(segs) == 2 && segs[0] == "taxonomy":
		if segs[1] == "" {
			return selector{}, fmt.Errorf("empty taxonomy")
		}
		return selector{taxonomy: segs[1]}, nil
	case len(segs) == 3 && segs[0] == "taxonomy":
		if segs[1] == "" || segs[2] == "" {
			return selector{}, fmt.Errorf("empty taxonomy or term")
		}
		return selector{taxonomy: segs[1], term: segs[2]}, nil
	default:
		return selector{}, fmt.Errorf("unrecognized selector %q", strings.Join(segs, ":"))
	}
}

// CanWriteAll reports whether the scope grants unrestricted write (write:*).
func (ss *ScopeSet) CanWriteAll() bool {
	for _, r := range ss.rules {
		if r.action == ActionWrite && !r.deny && r.sel.all {
			return true
		}
	}
	return false
}

// CanWriteInbox reports whether the scope grants writing into the quarantine
// inbox (write:inbox), or has the broader write:*.
func (ss *ScopeSet) CanWriteInbox() bool {
	if ss.CanWriteAll() {
		return true
	}
	for _, r := range ss.rules {
		if r.action == ActionWrite && !r.deny && r.sel.taxonomy == "inbox" {
			return true
		}
	}
	return false
}

// ReadableFilenamesSQL compiles the read + deny rules into a subquery selecting
// the filenames the scope may read, with bound args. A document is readable iff
// some read-allow rule matches AND no deny rule matches any of its memberships.
// With no read-allow rule, the subquery selects nothing (deny by default).
func (ss *ScopeSet) ReadableFilenamesSQL() (string, []any) {
	var allow, deny []string
	var allowArgs, denyArgs []any

	membershipExists := func(sel selector) (string, []any) {
		switch {
		case sel.term != "":
			return "EXISTS (SELECT 1 FROM membership m WHERE m.filename = d.filename AND m.taxonomy = ? AND m.term = ?)",
				[]any{sel.taxonomy, sel.term}
		default:
			return "EXISTS (SELECT 1 FROM membership m WHERE m.filename = d.filename AND m.taxonomy = ?)",
				[]any{sel.taxonomy}
		}
	}

	for _, r := range ss.rules {
		if r.action != ActionRead {
			continue // write/delete rules are retained but not part of read filtering
		}
		if r.sel.all {
			if r.deny {
				deny = append(deny, "1=1")
			} else {
				allow = append(allow, "1=1")
			}
			continue
		}
		cond, condArgs := membershipExists(r.sel)
		if r.deny {
			deny = append(deny, cond)
			denyArgs = append(denyArgs, condArgs...)
		} else {
			allow = append(allow, cond)
			allowArgs = append(allowArgs, condArgs...)
		}
	}

	allowExpr := "0" // deny by default: nothing allowed
	if len(allow) > 0 {
		allowExpr = "(" + strings.Join(allow, " OR ") + ")"
	}
	denyExpr := "0" // nothing denied
	if len(deny) > 0 {
		denyExpr = "(" + strings.Join(deny, " OR ") + ")"
	}

	// Args must follow placeholder order in the rendered SQL: all allow
	// conditions first, then all deny conditions.
	args := append(allowArgs, denyArgs...)
	sql := "SELECT d.filename FROM docs d WHERE " + allowExpr + " AND NOT " + denyExpr
	return sql, args
}
