package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "audit.log")
	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := log.Append(Entry{Identity: "a", Tool: "create_doc", Slug: "x.md", Outcome: "ok"}); err != nil {
		t.Fatal(err)
	}
	if err := log.Append(Entry{Identity: "a", Tool: "archive", Slug: "x.md", Outcome: "ok"}); err != nil {
		t.Fatal(err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}

	entries := readEntries(t, path)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Tool != "create_doc" || entries[1].Tool != "archive" {
		t.Fatalf("order not preserved: %+v", entries)
	}
	if entries[0].Time == "" {
		t.Fatal("time should be auto-filled")
	}

	// Re-opening appends; the first entry is unchanged.
	log2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := log2.Append(Entry{Identity: "a", Tool: "set_front_matter", Outcome: "ok"}); err != nil {
		t.Fatal(err)
	}
	_ = log2.Close()

	entries = readEntries(t, path)
	if len(entries) != 3 || entries[0].Tool != "create_doc" {
		t.Fatalf("append-only violated after reopen: %+v", entries)
	}
}

func TestLogPathOutsideCorpus(t *testing.T) {
	// The caller places the log under stateDir; assert Open honors an arbitrary
	// path outside any corpus dir and creates parents.
	dir := t.TempDir()
	corpus := filepath.Join(dir, "corpus")
	state := filepath.Join(dir, "state")
	if err := os.MkdirAll(corpus, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(state, "audit.log")
	log, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = log.Close() }()
	if strings.HasPrefix(path, corpus+string(os.PathSeparator)) {
		t.Fatal("audit log must not be inside the corpus")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}

func readEntries(t *testing.T, path string) []Entry {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	var out []Entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("bad audit line %q: %v", line, err)
		}
		out = append(out, e)
	}
	return out
}
