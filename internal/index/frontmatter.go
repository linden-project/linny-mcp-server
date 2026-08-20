package index

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// errNoFrontMatter indicates the record did not open with a `---` fence.
var errNoFrontMatter = errors.New("no front matter")

// parseRecord parses a raw markdown record into a Record. It lowercases
// front-matter keys (matching Hugo's `.Params`) and counts task items. A nil
// error with a nil Props is never returned; malformed front matter yields a
// non-nil error and the caller decides how to report it.
func parseRecord(filename, content string) (*Record, error) {
	fm, body, err := splitFrontMatter(content)
	if err != nil {
		return nil, err
	}

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(fm), &raw); err != nil {
		return nil, err
	}
	props := lowerKeys(raw)

	title, _ := props["title"].(string)

	rec := &Record{
		Filename: filename,
		Props:    props,
		Body:     body,
		Title:    title,
		Tasks:    countTasks(body),
	}
	return rec, nil
}

// splitFrontMatter separates a leading `---`-fenced YAML block from the body.
func splitFrontMatter(content string) (fm, body string, err error) {
	// Tolerate a leading BOM / whitespace-free "---\n".
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return "", content, errNoFrontMatter
	}
	rest := content[strings.IndexByte(content, '\n')+1:]
	// Find the closing fence at the start of a line.
	end := findClosingFence(rest)
	if end < 0 {
		return "", content, errors.New("unterminated front matter")
	}
	fm = rest[:end]
	// Body starts after the closing fence line.
	afterFence := rest[end:]
	if nl := strings.IndexByte(afterFence, '\n'); nl >= 0 {
		body = afterFence[nl+1:]
	}
	return fm, body, nil
}

// findClosingFence returns the byte offset of a line that is exactly `---`.
func findClosingFence(s string) int {
	offset := 0
	for _, line := range splitKeepIndex(s) {
		trimmed := strings.TrimRight(s[line.start:line.end], "\r")
		if trimmed == "---" {
			return line.start
		}
		offset = line.end
	}
	_ = offset
	return -1
}

type lineSpan struct{ start, end int }

// splitKeepIndex yields line spans (start inclusive, end exclusive of newline).
func splitKeepIndex(s string) []lineSpan {
	var spans []lineSpan
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			spans = append(spans, lineSpan{start, i})
			start = i + 1
		}
	}
	if start <= len(s) {
		spans = append(spans, lineSpan{start, len(s)})
	}
	return spans
}

// lowerKeys returns a copy of m with top-level keys lowercased.
func lowerKeys(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[strings.ToLower(k)] = v
	}
	return out
}

// countTasks counts Markdown task-list items. A task marker must be the first
// non-space content of the line: optional indentation, then `- [ ]` or `- [x]`
// (lowercase x for closed, per the Hugo reference).
func countTasks(body string) TaskCount {
	var tc TaskCount
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimLeft(raw, " \t")
		switch {
		case strings.HasPrefix(line, "- [ ] "):
			tc.Open++
		case strings.HasPrefix(line, "- [x] "):
			tc.Closed++
		}
	}
	tc.Total = tc.Open + tc.Closed
	return tc
}

// conflictLines returns the conflict-marker lines found in content. A file is
// conflicted if it contains a line beginning with `<<<<<<<` or `>>>>>>>`
// (unambiguous git markers); `=======` alone is ignored to avoid matching
// Markdown setext headings.
func conflictLines(content string) []string {
	var hits []string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "<<<<<<<") || strings.HasPrefix(line, ">>>>>>>") {
			hits = append(hits, line)
		}
	}
	return hits
}
