// Package defense holds hostile-corpus guardrails: the quarantine policy for
// agent writes and the data-delimiter wrapping for returned note bodies. The
// corpus is untrusted input (prompt injection); these reduce blast radius.
package defense

// Policy configures the hostile-corpus guardrails.
type Policy struct {
	// QuarantineTaxonomy/Term is where agent-created documents land by default.
	QuarantineTaxonomy string
	QuarantineTerm     string
	// Disabled turns off quarantine-on-create (agent writes go straight in). This
	// removes a hostile-corpus defense; use only when you trust the client.
	Disabled bool
	// confirmTools are operations that require out-of-band confirmation and must
	// not be performed by an in-band tool call alone.
	confirmTools map[string]bool
}

// DefaultPolicy returns the standard guardrail policy.
func DefaultPolicy() Policy {
	return Policy{
		QuarantineTaxonomy: "status",
		QuarantineTerm:     "agent-draft",
		confirmTools: map[string]bool{
			"delete":     true,
			"bulk_retag": true,
		},
	}
}

// ApplyQuarantine ensures front matter places the document in the quarantine
// term. It merges into any existing value for the quarantine taxonomy: a missing
// key becomes a one-element list; a scalar becomes a two-element list; a list
// gains the term if absent. Front matter is mutated in place.
func (p Policy) ApplyQuarantine(front map[string]any) {
	if front == nil || p.Disabled {
		return
	}
	existing, ok := front[p.QuarantineTaxonomy]
	if !ok {
		front[p.QuarantineTaxonomy] = []any{p.QuarantineTerm}
		return
	}
	switch v := existing.(type) {
	case string:
		if v == p.QuarantineTerm {
			return
		}
		front[p.QuarantineTaxonomy] = []any{v, p.QuarantineTerm}
	case []any:
		for _, e := range v {
			if s, ok := e.(string); ok && s == p.QuarantineTerm {
				return // already quarantined
			}
		}
		front[p.QuarantineTaxonomy] = append(v, p.QuarantineTerm)
	default:
		front[p.QuarantineTaxonomy] = []any{p.QuarantineTerm}
	}
}

// IsQuarantined reports whether front matter carries the quarantine term.
func (p Policy) IsQuarantined(front map[string]any) bool {
	v, ok := front[p.QuarantineTaxonomy]
	if !ok {
		return false
	}
	switch t := v.(type) {
	case string:
		return t == p.QuarantineTerm
	case []any:
		for _, e := range t {
			if s, ok := e.(string); ok && s == p.QuarantineTerm {
				return true
			}
		}
	}
	return false
}

// RequiresConfirmation reports whether a tool needs out-of-band confirmation.
func (p Policy) RequiresConfirmation(tool string) bool {
	return p.confirmTools[tool]
}
