package gitsafe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveGitDirFilePointer(t *testing.T) {
	root := t.TempDir()
	realGit := filepath.Join(t.TempDir(), "actual.git")
	if err := os.MkdirAll(realGit, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realGit, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// A `.git` FILE pointing elsewhere (worktree/submodule form).
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: "+realGit+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolveGitDir(root); got != realGit {
		t.Fatalf("resolveGitDir = %q, want %q", got, realGit)
	}
	// A merge marker in the pointed-to git dir must be detected via inspect.
	if err := os.WriteFile(filepath.Join(realGit, "MERGE_HEAD"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if st := inspect(root); st.InProgress != "merge" || st.Clean {
		t.Fatalf("expected merge in progress via gitdir file, got %+v", st)
	}
}

func TestResolveGitDirMissing(t *testing.T) {
	if got := resolveGitDir(t.TempDir()); got != "" {
		t.Fatalf("no .git should resolve to empty, got %q", got)
	}
}

func TestInProgressOpVariants(t *testing.T) {
	cases := map[string]string{
		"CHERRY_PICK_HEAD": "cherry-pick",
		"REVERT_HEAD":      "revert",
	}
	for marker, want := range cases {
		gitDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(gitDir, marker), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := inProgressOp(gitDir); got != want {
			t.Fatalf("%s -> %q, want %q", marker, got, want)
		}
	}
	// rebase-apply directory -> "rebase"
	gitDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(gitDir, "rebase-apply"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := inProgressOp(gitDir); got != "rebase" {
		t.Fatalf("rebase-apply -> %q, want rebase", got)
	}
	// none
	if got := inProgressOp(t.TempDir()); got != "" {
		t.Fatalf("no marker -> %q, want empty", got)
	}
}

func TestIsDetachedMissingHEAD(t *testing.T) {
	if isDetached(t.TempDir()) {
		t.Fatal("missing HEAD should not report detached")
	}
}

func TestCheckRefArgEmpty(t *testing.T) {
	if err := checkRefArg(""); err == nil {
		t.Fatal("empty ref/date should error")
	}
}

func TestHistoryNoCommitsForPath(t *testing.T) {
	root := gitRepo(t) // from history_test.go; skips if git absent
	commits, err := History(root, "content/does_not_exist.md", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 0 {
		t.Fatalf("expected no history for a nonexistent path, got %d", len(commits))
	}
}
