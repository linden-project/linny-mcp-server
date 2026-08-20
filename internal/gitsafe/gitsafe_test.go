package gitsafe

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// fakeTree builds a corpus root with a hand-crafted .git dir (a normal branch
// checkout) and a content dir. Returns the root.
func fakeTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitDir := filepath.Join(root, ".git")
	mustMkdir(t, gitDir)
	mustWrite(t, filepath.Join(gitDir, "HEAD"), "ref: refs/heads/main\n")
	mustMkdir(t, filepath.Join(root, "content"))
	mustWrite(t, filepath.Join(root, "content", "note.md"), "---\ntitle: Note\n---\nbody\n")
	return root
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, content string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInspectCleanTree(t *testing.T) {
	st := inspect(fakeTree(t))
	if !st.Clean {
		t.Fatalf("expected clean tree, got %+v", st)
	}
	if st.Conflicted || st.InProgress != "" || st.Detached {
		t.Fatalf("clean tree has flags set: %+v", st)
	}
}

func TestInspectMergeInProgress(t *testing.T) {
	root := fakeTree(t)
	mustWrite(t, filepath.Join(root, ".git", "MERGE_HEAD"), "deadbeef\n")
	st := inspect(root)
	if st.Clean || st.InProgress != "merge" {
		t.Fatalf("expected merge-in-progress, got %+v", st)
	}
}

func TestInspectRebaseInProgress(t *testing.T) {
	root := fakeTree(t)
	mustMkdir(t, filepath.Join(root, ".git", "rebase-merge"))
	st := inspect(root)
	if st.Clean || st.InProgress != "rebase" {
		t.Fatalf("expected rebase-in-progress, got %+v", st)
	}
}

func TestInspectConflictMarkers(t *testing.T) {
	root := fakeTree(t)
	mustWrite(t, filepath.Join(root, "content", "bad.md"),
		"---\ntitle: Bad\n---\n<<<<<<< HEAD\nmine\n=======\ntheirs\n>>>>>>> other\n")
	st := inspect(root)
	if !st.Conflicted {
		t.Fatalf("expected conflicted, got %+v", st)
	}
	if len(st.ConflictedPaths) != 1 || st.ConflictedPaths[0] != filepath.Join("content", "bad.md") {
		t.Fatalf("expected content/bad.md conflicted, got %v", st.ConflictedPaths)
	}
	if st.Clean {
		t.Fatal("conflicted tree must not be clean")
	}
}

func TestInspectDetached(t *testing.T) {
	root := fakeTree(t)
	mustWrite(t, filepath.Join(root, ".git", "HEAD"), "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2\n")
	st := inspect(root)
	if !st.Detached || st.Clean {
		t.Fatalf("expected detached HEAD, got %+v", st)
	}
}

func TestGuardEnsureWritableAndRecovery(t *testing.T) {
	root := fakeTree(t)
	g := NewGuard(root, false)

	if err := g.EnsureWritable(); err != nil {
		t.Fatalf("clean tree should be writable, got %v", err)
	}

	// Introduce a conflict → degraded.
	badPath := filepath.Join(root, "content", "bad.md")
	mustWrite(t, badPath, "<<<<<<< HEAD\nx\n>>>>>>> y\n")
	err := g.EnsureWritable()
	var de *DegradedError
	if !errors.As(err, &de) {
		t.Fatalf("expected DegradedError, got %v", err)
	}
	if !de.Retryable() {
		t.Fatal("degraded error must be retryable")
	}

	// Remove the conflict → automatic recovery, no manual reset.
	if err := os.Remove(badPath); err != nil {
		t.Fatal(err)
	}
	if err := g.EnsureWritable(); err != nil {
		t.Fatalf("tree recovered but still degraded: %v", err)
	}
}

func TestForcedReadOnly(t *testing.T) {
	g := NewGuard(fakeTree(t), true)
	err := g.EnsureWritable()
	var de *DegradedError
	if !errors.As(err, &de) || !de.Forced {
		t.Fatalf("expected forced DegradedError, got %v", err)
	}
}

func TestAtomicWriteNoTempLeft(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")
	if err := AtomicWrite(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil || string(b) != "hello" {
		t.Fatalf("content = %q err=%v", b, err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected only the target file, found %d entries (temp left behind?)", len(entries))
	}
}

func TestWriteIfUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "note.md")

	// Create: expected hash "" (must not yet exist).
	if err := WriteIfUnchanged(path, []byte("v1"), "", 0o644); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Stale: wrong expected hash → retryable StaleWriteError, file unchanged.
	err := WriteIfUnchanged(path, []byte("v2"), "deadbeef", 0o644)
	var se *StaleWriteError
	if !errors.As(err, &se) || !se.Retryable() {
		t.Fatalf("expected retryable StaleWriteError, got %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "v1" {
		t.Fatalf("file must be unchanged after stale write, got %q", b)
	}

	// Fresh: supply the current hash → success.
	cur, err := HashFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteIfUnchanged(path, []byte("v2"), cur, 0o644); err != nil {
		t.Fatalf("fresh write: %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "v2" {
		t.Fatalf("expected v2, got %q", b)
	}
}

func TestHashFileMissing(t *testing.T) {
	h, err := HashFile(filepath.Join(t.TempDir(), "nope.md"))
	if err != nil || h != "" {
		t.Fatalf("missing file should hash to empty string, got %q err=%v", h, err)
	}
}
