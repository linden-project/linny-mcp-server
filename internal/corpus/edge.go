package corpus

import (
	"fmt"
	"sort"
	"strings"
)

// edgeRecord is a hand-crafted record exercising a deliberate edge case. Its
// content is emitted verbatim (renderRaw) so malformed YAML etc. is preserved.
type edgeRecord struct {
	slug string
	taxa map[string][]string
	raw  string
}

func (e edgeRecord) renderRaw() string { return e.raw }

// edgeRecords returns the fixed set of edge-case records. Deterministic: no RNG.
func edgeRecords() []edgeRecord {
	return []edgeRecord{
		{
			slug: "unicode_notes",
			taxa: map[string][]string{"tags": {"note"}, "subject": {"linny"}},
			raw: "---\n" +
				"title: Ünìcode — café, naïve, 日本語, emoji 🌱\n" +
				"crdate: \"2024-02-29\"\n" +
				"tags: [note, reference]\n" +
				"subject: linny\n" +
				"---\n" +
				"Grüße from the second brain. 日本語のテキスト and some math ∑∫√.\n",
		},
		{
			slug: "long_front_matter",
			taxa: map[string][]string{"tags": {"reference"}, "customer": {"eric"}},
			raw:  longFrontMatter(),
		},
		{
			slug: "empty_body",
			taxa: map[string][]string{"tags": {"note"}},
			raw: "---\n" +
				"title: Empty Body\n" +
				"crdate: \"2024-01-01\"\n" +
				"tags: note\n" +
				"---\n",
		},
		{
			// Malformed YAML: unterminated flow sequence. A robust indexer must
			// not crash; it reports/skips per the spec.
			slug: "malformed_yaml",
			taxa: nil,
			raw: "---\n" +
				"title: Malformed [unterminated\n" +
				"crdate: \"not a date\n" +
				"tags: {broken: \n" +
				"---\n" +
				"Body after malformed front matter.\n",
		},
		{
			// Committed git conflict markers inside tracked content.
			slug: "conflict_markers",
			taxa: map[string][]string{"tags": {"todo"}},
			raw: "---\n" +
				"title: Conflicted Note\n" +
				"crdate: \"2024-03-03\"\n" +
				"tags: todo\n" +
				"---\n" +
				"Some intro text.\n\n" +
				"<<<<<<< HEAD\n" +
				"our side of the merge\n" +
				"=======\n" +
				"their side of the merge\n" +
				">>>>>>> feature-branch\n\n" +
				"Trailing text.\n",
		},
		{
			// FAKE secrets (not real) to exercise the egress redaction filter.
			slug: "fake_secrets",
			taxa: map[string][]string{"tags": {"finance"}, "customer": {"globex"}},
			raw: "---\n" +
				"title: Account Setup (fake secrets)\n" +
				"crdate: \"2024-04-04\"\n" +
				"tags: [finance, reference]\n" +
				"customer: globex\n" +
				"---\n" +
				"These are NOT real credentials — synthetic fixtures for redaction tests.\n\n" +
				"AWS key: AKIAIOSFODNN7EXAMPLE\n" +
				"aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY\n" +
				"token: ghp_0123456789abcdefghijklmnopqrstuvwxyzAB\n" +
				"IBAN: NL91ABNA0417164300\n\n" +
				"-----BEGIN RSA PRIVATE KEY-----\n" +
				"MIIBOgIBAAJBAKj34GkxFhD90vcNLYLInFEX4MPLE0nlyForTestingNotRealAtAll==\n" +
				"-----END RSA PRIVATE KEY-----\n",
		},
		{
			// The scope-intersection fixture: a doc tagged BOTH work and health.
			slug: "work_and_health",
			taxa: map[string][]string{"tags": {"work", "health"}, "subject": {"linny"}},
			raw: "---\n" +
				"title: Work And Health\n" +
				"crdate: \"2024-05-05\"\n" +
				"tags: [health, work]\n" +
				"subject: linny\n" +
				"---\n" +
				"A note that belongs to both work and health; used to test that a deny\n" +
				"on health excludes it even though work is allowed.\n",
		},
	}
}

func longFrontMatter() string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("title: Long Front Matter\n")
	b.WriteString("crdate: \"2024-06-06\"\n")
	b.WriteString("tags: reference\n")
	b.WriteString("customer: eric\n")
	// A large but valid set of extra scalar keys.
	keys := make([]string, 0, 60)
	for i := 0; i < 60; i++ {
		keys = append(keys, fmt.Sprintf("extra_key_%02d", i))
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: value for %s with some padding text to make it long\n", k, k)
	}
	b.WriteString("---\n")
	b.WriteString("Body after a very long front matter block.\n")
	return b.String()
}
