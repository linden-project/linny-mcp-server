package mcp

import (
	"fmt"
	"strings"
)

// diffCellLimit caps the LCS table. A note far past this is not worth a
// character-accurate audit entry, and the fallback says so rather than
// pretending nothing changed.
const diffCellLimit = 1_000_000

// diffContext is the number of unchanged lines kept around each hunk.
const diffContext = 3

// unifiedDiff renders a line-based unified diff of old against new. The audit
// log records what a write changed; storing the full new document there would
// keep a second copy of every note that is ever edited.
func unifiedDiff(oldText, newText string) string {
	if oldText == newText {
		return ""
	}
	a, b := splitLines(oldText), splitLines(newText)
	if len(a)*len(b) > diffCellLimit {
		return fmt.Sprintf("@@ diff omitted: %d -> %d lines @@\n", len(a), len(b))
	}

	var out strings.Builder
	for _, h := range hunks(editScript(a, b)) {
		fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n", h.oldStart, h.oldLines, h.newStart, h.newLines)
		for _, e := range h.edits {
			out.WriteString(string(e.op))
			out.WriteString(e.line)
			out.WriteString("\n")
		}
	}
	return out.String()
}

// splitLines splits on newlines, dropping the empty element a trailing newline
// produces so a final "\n" is not reported as a change.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}

type edit struct {
	op   byte // ' ' keep, '-' remove, '+' add
	line string
}

// editScript walks an LCS table back into a line-by-line edit script.
func editScript(a, b []string) []edit {
	// lcs[i][j] = length of the longest common subsequence of a[i:] and b[j:].
	lcs := make([][]int32, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int32, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
				continue
			}
			lcs[i][j] = max32(lcs[i+1][j], lcs[i][j+1])
		}
	}

	var out []edit
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			out = append(out, edit{' ', a[i]})
			i, j = i+1, j+1
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, edit{'-', a[i]})
			i++
		default:
			out = append(out, edit{'+', b[j]})
			j++
		}
	}
	for ; i < len(a); i++ {
		out = append(out, edit{'-', a[i]})
	}
	for ; j < len(b); j++ {
		out = append(out, edit{'+', b[j]})
	}
	return out
}

func max32(x, y int32) int32 {
	if x > y {
		return x
	}
	return y
}

type hunk struct {
	oldStart, oldLines int
	newStart, newLines int
	edits              []edit
}

// hunks groups an edit script into unified-diff hunks, keeping diffContext
// unchanged lines on each side of a run of changes.
func hunks(edits []edit) []hunk {
	// Mark which entries must appear: every change, plus its context.
	keep := make([]bool, len(edits))
	for i, e := range edits {
		if e.op == ' ' {
			continue
		}
		lo, hi := i-diffContext, i+diffContext
		if lo < 0 {
			lo = 0
		}
		if hi >= len(edits) {
			hi = len(edits) - 1
		}
		for k := lo; k <= hi; k++ {
			keep[k] = true
		}
	}

	var out []hunk
	oldLine, newLine := 1, 1
	i := 0
	for i < len(edits) {
		if !keep[i] {
			if edits[i].op != '+' {
				oldLine++
			}
			if edits[i].op != '-' {
				newLine++
			}
			i++
			continue
		}
		h := hunk{oldStart: oldLine, newStart: newLine}
		for ; i < len(edits) && keep[i]; i++ {
			h.edits = append(h.edits, edits[i])
			if edits[i].op != '+' {
				oldLine++
				h.oldLines++
			}
			if edits[i].op != '-' {
				newLine++
				h.newLines++
			}
		}
		out = append(out, h)
	}
	return out
}
