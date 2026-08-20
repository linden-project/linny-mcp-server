package gitsafe

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitRepo initializes a temp git repo with two commits touching content/foo.md
// and returns its root. It skips the test if git is unavailable.
func gitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	root := t.TempDir()
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=t@t")
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "content"), 0o755); err != nil {
		t.Fatal(err)
	}
	foo := filepath.Join(root, "content", "foo.md")
	if err := os.WriteFile(foo, []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("init", "-q")
	run("add", "-A")
	run("commit", "-qm", "add foo")
	if err := os.WriteFile(foo, []byte("v1\nv2 added line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-qm", "update foo")
	return root
}

func TestHistory(t *testing.T) {
	root := gitRepo(t)
	commits, err := History(root, "content/foo.md", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d: %+v", len(commits), commits)
	}
	// Newest first.
	if commits[0].Subject != "update foo" || commits[1].Subject != "add foo" {
		t.Fatalf("unexpected order/subjects: %+v", commits)
	}
	if commits[0].Hash == "" || commits[0].Author != "Test" || commits[0].Date == "" {
		t.Fatalf("commit fields incomplete: %+v", commits[0])
	}
}

func TestDiff(t *testing.T) {
	root := gitRepo(t)
	out, err := Diff(root, "content/foo.md", "HEAD~1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "v2 added line") {
		t.Fatalf("diff should show the added line, got:\n%s", out)
	}
}

func TestChangedSince(t *testing.T) {
	root := gitRepo(t)
	paths, err := ChangedSince(root, "1970-01-01")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range paths {
		if p == "content/foo.md" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected content/foo.md in changed paths, got %v", paths)
	}
}

func TestRefArgRejectsFlags(t *testing.T) {
	root := gitRepo(t)
	if _, err := Diff(root, "content/foo.md", "--output=/tmp/x"); err == nil {
		t.Fatal("expected flag-like ref to be rejected")
	}
	if _, err := ChangedSince(root, "-1"); err == nil {
		t.Fatal("expected flag-like since to be rejected")
	}
}
