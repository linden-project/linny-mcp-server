package gitsafe

import (
	"fmt"
	"os/exec"
	"strings"
)

// Commit is one entry in a document's git history.
type Commit struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Date    string `json:"date"` // ISO-8601 (author date)
	Subject string `json:"subject"`
}

// historyFormat separates fields with US (\x1f) so subjects may contain anything.
const historyFormat = "%H%x1f%an%x1f%aI%x1f%s"

// History returns up to limit commits touching relpath, newest first.
func History(root, relpath string, limit int) ([]Commit, error) {
	if limit <= 0 {
		limit = 50
	}
	out, err := runGit(root, "log", fmt.Sprintf("-n%d", limit), "--format="+historyFormat, "--", relpath)
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		f := strings.Split(line, "\x1f")
		if len(f) != 4 {
			continue
		}
		commits = append(commits, Commit{Hash: f[0], Author: f[1], Date: f[2], Subject: f[3]})
	}
	return commits, nil
}

// Diff returns the textual diff of relpath between ref and the working tree.
func Diff(root, relpath, ref string) (string, error) {
	if err := checkRefArg(ref); err != nil {
		return "", err
	}
	return runGit(root, "diff", ref, "--", relpath)
}

// ChangedSince returns the distinct paths changed since the given date/revision.
func ChangedSince(root, since string) ([]string, error) {
	if err := checkRefArg(since); err != nil {
		return nil, err
	}
	out, err := runGit(root, "log", "--since="+since, "--name-only", "--format=")
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		p := strings.TrimSpace(line)
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		paths = append(paths, p)
	}
	return paths, nil
}

// checkRefArg rejects values that would be interpreted as git flags. Arguments
// are passed without a shell, so this is the only injection surface.
func checkRefArg(v string) error {
	if v == "" {
		return fmt.Errorf("gitsafe: empty ref/date")
	}
	if strings.HasPrefix(v, "-") {
		return fmt.Errorf("gitsafe: ref/date must not begin with '-': %q", v)
	}
	return nil
}

// runGit runs a git subcommand in root and returns stdout.
func runGit(root string, args ...string) (string, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		return "", fmt.Errorf("gitsafe: git binary not found: %w", err)
	}
	full := append([]string{"-C", root}, args...)
	out, err := exec.Command(git, full...).Output()
	if err != nil {
		return "", fmt.Errorf("gitsafe: git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}
