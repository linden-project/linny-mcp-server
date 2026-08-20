package index

import "testing"

func TestSplitFrontMatter(t *testing.T) {
	fm, body, err := splitFrontMatter("---\ntitle: Hi\n---\nBody line\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm != "title: Hi\n" {
		t.Fatalf("front matter = %q", fm)
	}
	if body != "Body line\n" {
		t.Fatalf("body = %q", body)
	}
}

func TestSplitFrontMatterNoFence(t *testing.T) {
	if _, _, err := splitFrontMatter("no front matter here\n"); err == nil {
		t.Fatal("expected error for missing front matter")
	}
}

func TestParseRecordLowercasesKeys(t *testing.T) {
	rec, err := parseRecord("x.md", "---\nTitle: Cap\nCustomer: Eric\n---\nbody\n")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := rec.Props["title"]; !ok {
		t.Fatalf("expected lowercased key 'title', got %v", rec.Props)
	}
	if rec.Title != "Cap" {
		t.Fatalf("title = %q", rec.Title)
	}
}

func TestCountTasks(t *testing.T) {
	body := "intro\n- [ ] a\n- [x] b\n  - [ ] nested open\n- [X] not-closed-uppercase\n"
	tc := countTasks(body)
	// Two open (a, nested open), one closed (b). Uppercase [X] is not counted.
	if tc.Open != 2 || tc.Closed != 1 || tc.Total != 3 {
		t.Fatalf("counts = %+v, want open2 closed1 total3", tc)
	}
}

func TestConflictLines(t *testing.T) {
	c := "ok\n<<<<<<< HEAD\nmine\n=======\ntheirs\n>>>>>>> branch\n"
	hits := conflictLines(c)
	if len(hits) != 2 {
		t.Fatalf("expected 2 unambiguous markers, got %d: %v", len(hits), hits)
	}
	// A setext heading underline must not be treated as a conflict.
	if len(conflictLines("Title\n=======\ntext\n")) != 0 {
		t.Fatal("setext underline should not count as a conflict")
	}
}
