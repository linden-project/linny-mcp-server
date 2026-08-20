package gitsafe

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// TreeState is the result of inspecting a notebook's git working tree.
type TreeState struct {
	// Clean is true only when the tree is safe to write: no conflicts, no
	// in-progress operation, no unmerged entries, and not detached.
	Clean bool
	// Conflicted is true when tracked content carries committed conflict markers
	// (or, when the git binary is available, unmerged index entries exist).
	Conflicted bool
	// ConflictedPaths lists the tracked files with committed conflict markers.
	ConflictedPaths []string
	// InProgress names an in-progress git operation ("merge", "rebase",
	// "cherry-pick", "revert") or "" when none.
	InProgress string
	// Detached is true when HEAD is not a symbolic ref.
	Detached bool
	// Reason is a human-readable summary of why the tree is not clean.
	Reason string
}

// inspect examines the working tree rooted at root and returns its state.
func inspect(root string) TreeState {
	st := TreeState{}
	gitDir := resolveGitDir(root)

	if gitDir != "" {
		st.InProgress = inProgressOp(gitDir)
		st.Detached = isDetached(gitDir)
	}

	paths := scanConflictMarkers(root)
	if len(paths) > 0 {
		st.Conflicted = true
		st.ConflictedPaths = paths
	}

	// Best-effort: unmerged index entries via the git binary, if present.
	if hasUnmergedEntries(root) {
		st.Conflicted = true
	}

	var reasons []string
	if st.Conflicted {
		reasons = append(reasons, "committed conflict markers or unmerged entries present")
	}
	if st.InProgress != "" {
		reasons = append(reasons, st.InProgress+" in progress")
	}
	if st.Detached {
		reasons = append(reasons, "detached HEAD")
	}
	st.Clean = len(reasons) == 0
	st.Reason = strings.Join(reasons, "; ")
	return st
}

// resolveGitDir returns the git directory for root, handling both a `.git`
// directory and a `.git` file ("gitdir: <path>"). Returns "" if not found.
func resolveGitDir(root string) string {
	dotgit := filepath.Join(root, ".git")
	info, err := os.Stat(dotgit)
	if err != nil {
		return ""
	}
	if info.IsDir() {
		return dotgit
	}
	// `.git` is a file pointing elsewhere (worktree / submodule).
	b, err := os.ReadFile(dotgit)
	if err != nil {
		return ""
	}
	line := strings.TrimSpace(string(b))
	const p = "gitdir:"
	if !strings.HasPrefix(line, p) {
		return ""
	}
	target := strings.TrimSpace(strings.TrimPrefix(line, p))
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}
	return target
}

// inProgressOp reports an in-progress git operation by inspecting the git dir.
func inProgressOp(gitDir string) string {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(gitDir, name))
		return err == nil
	}
	switch {
	case exists("rebase-merge"), exists("rebase-apply"):
		return "rebase"
	case exists("CHERRY_PICK_HEAD"):
		return "cherry-pick"
	case exists("REVERT_HEAD"):
		return "revert"
	case exists("MERGE_HEAD"):
		return "merge"
	default:
		return ""
	}
}

// isDetached reports whether HEAD is not a symbolic ref.
func isDetached(gitDir string) bool {
	b, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return false
	}
	return !strings.HasPrefix(strings.TrimSpace(string(b)), "ref:")
}

// scanConflictMarkers walks *.md files under root (excluding the git dir) and
// returns the sorted set of files containing committed conflict markers.
func scanConflictMarkers(root string) []string {
	var hits []string
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable entries; never abort inspection
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == ".jj" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if hasConflictMarker(b) {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			hits = append(hits, rel)
		}
		return nil
	})
	sort.Strings(hits)
	return hits
}

// hasConflictMarker reports whether content has a line beginning with `<<<<<<<`
// or `>>>>>>>`. Bare `=======` is ignored to avoid matching setext headings.
func hasConflictMarker(content []byte) bool {
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "<<<<<<<") || strings.HasPrefix(line, ">>>>>>>") {
			return true
		}
	}
	return false
}

// hasUnmergedEntries uses the git binary, when present, to detect unmerged
// index entries. It is best-effort: any error (including git absent) yields
// false, leaving the marker/merge-state checks as the safety net.
func hasUnmergedEntries(root string) bool {
	git, err := exec.LookPath("git")
	if err != nil {
		return false
	}
	cmd := exec.Command(git, "-C", root, "ls-files", "-u")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}
