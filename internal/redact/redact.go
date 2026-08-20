// Package redact applies a gitleaks-style secret-redaction filter to tool
// responses so no response can return a credential regardless of what an agent
// asks for. It is rule-based and best-effort: it reduces blast radius, it is not
// a substitute for keeping secrets out of the corpus.
package redact

import (
	"regexp"
	"strings"
)

// detector matches a credential shape. group == 0 redacts the whole match;
// group > 0 redacts only that submatch (keeping surrounding context, e.g. the
// key name in an assignment).
type detector struct {
	name  string
	re    *regexp.Regexp
	group int
}

// detectors are applied in order. More specific / structural patterns (PEM
// blocks, typed tokens) run before the broad generic-assignment rule.
var detectors = []detector{
	{
		name: "private-key",
		re:   regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`),
	},
	{
		name: "aws-access-key",
		re:   regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`),
	},
	{
		name:  "aws-secret-key",
		re:    regexp.MustCompile(`(?i)(aws_secret_access_key\s*[:=]\s*["']?)([A-Za-z0-9/+]{40})`),
		group: 2,
	},
	{
		name: "github-token",
		re:   regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{36,}\b`),
	},
	{
		name: "slack-token",
		re:   regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`),
	},
	{
		name: "jwt",
		re:   regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`),
	},
	{
		name: "iban",
		re:   regexp.MustCompile(`\b[A-Z]{2}[0-9]{2}[A-Z0-9]{11,30}\b`),
	},
	{
		// Generic assignment: redact the value, keep the key + separator (group 1).
		// The value must not start with '[' so we never re-redact a placeholder a
		// more specific detector already produced (e.g. [REDACTED:github-token]).
		name:  "generic-secret",
		re:    regexp.MustCompile(`(?i)((?:api[_-]?key|secret|token|password|passwd|access[_-]?key)\s*[:=]\s*["']?)([^\s"'\[][^\s"']{5,})`),
		group: 2,
	},
}

// Redactor scrubs credentials from text and structured values.
type Redactor struct {
	detectors []detector
}

// New returns a Redactor with the default detector set.
func New() *Redactor { return &Redactor{detectors: detectors} }

// Redact returns text with every detected credential replaced by a typed
// placeholder, plus the number of redactions performed. The removed secret never
// appears in the returned text or count.
func (r *Redactor) Redact(text string) (string, int) {
	count := 0
	for _, d := range r.detectors {
		text = d.apply(text, &count)
	}
	return text, count
}

// RedactValue deep-walks a response value, scrubbing every string it contains.
// Maps and slices are mutated in place and returned; other types pass through.
func (r *Redactor) RedactValue(v any) any {
	switch t := v.(type) {
	case string:
		red, _ := r.Redact(t)
		return red
	case map[string]any:
		for k, val := range t {
			t[k] = r.RedactValue(val)
		}
		return t
	case []any:
		for i := range t {
			t[i] = r.RedactValue(t[i])
		}
		return t
	case []string:
		for i := range t {
			t[i], _ = r.Redact(t[i])
		}
		return t
	default:
		return v
	}
}

// apply redacts all matches of one detector, incrementing *count per redaction.
func (d detector) apply(s string, count *int) string {
	locs := d.re.FindAllStringSubmatchIndex(s, -1)
	if len(locs) == 0 {
		return s
	}
	var b strings.Builder
	last := 0
	for _, m := range locs {
		start, end := m[0], m[1]
		if d.group > 0 {
			gs, ge := m[2*d.group], m[2*d.group+1]
			if gs < 0 { // group did not participate
				continue
			}
			start, end = gs, ge
		}
		b.WriteString(s[last:start])
		b.WriteString("[REDACTED:" + d.name + "]")
		last = end
		*count++
	}
	b.WriteString(s[last:])
	return b.String()
}
