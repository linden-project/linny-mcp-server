package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Discrepancy is one difference found between two index trees.
type Discrepancy struct {
	File   string `json:"file"`
	Detail string `json:"detail"`
}

// VerifyDirs diffs two index directories and returns the discrepancies. Arrays
// are compared as sets (the spec declares index arrays unordered); objects are
// compared key-by-key; `_indexer_info.json` ignores identity/placeholder fields.
func VerifyDirs(ours, reference string) ([]Discrepancy, error) {
	ourFiles, err := indexJSONFiles(ours)
	if err != nil {
		return nil, err
	}
	refFiles, err := indexJSONFiles(reference)
	if err != nil {
		return nil, err
	}

	var d []Discrepancy
	for rel := range ourFiles {
		if _, ok := refFiles[rel]; !ok {
			d = append(d, Discrepancy{File: rel, Detail: "present in ours, absent in reference"})
		}
	}
	for rel := range refFiles {
		if _, ok := ourFiles[rel]; !ok {
			d = append(d, Discrepancy{File: rel, Detail: "present in reference, absent in ours"})
		}
	}
	for rel, ov := range ourFiles {
		rv, ok := refFiles[rel]
		if !ok {
			continue
		}
		base := filepath.Base(rel)
		if !valueEqual(normalizeForFile(base, ov), normalizeForFile(base, rv)) {
			d = append(d, Discrepancy{File: rel, Detail: "content differs"})
		}
	}
	sort.Slice(d, func(i, j int) bool {
		if d[i].File != d[j].File {
			return d[i].File < d[j].File
		}
		return d[i].Detail < d[j].Detail
	})
	return d, nil
}

// indexJSONFiles reads every *.json under root into a map of relpath -> value.
func indexJSONFiles(root string) (map[string]any, error) {
	out := map[string]any{}
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		var v any
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = v
		return nil
	})
	return out, err
}

// normalizeForFile drops fields that are expected to differ between a conforming
// indexer and the Hugo reference for _indexer_info.json.
func normalizeForFile(base string, v any) any {
	if base != fileIndexerInfo {
		return v
	}
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	out := map[string]any{}
	for k, val := range m {
		switch k {
		// Identity + environment-specific fields: Hugo emits the paths as the
		// literal "TODO" while a real indexer populates them, and product/engine
		// identity differs by construction. None are consumed, so ignore them.
		case "product_name", "product_version", "hugo_version",
			"index_dir", "content_dir", "config_dir":
			continue
		}
		if s, ok := val.(string); ok && s == "TODO" {
			continue
		}
		out[k] = val
	}
	return out
}

// valueEqual compares two decoded JSON values; top-level and nested arrays are
// compared as sets.
func valueEqual(a, b any) bool {
	switch av := a.(type) {
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		return sameSetCanonical(av, bv)
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, va := range av {
			vb, ok := bv[k]
			if !ok || !valueEqual(va, vb) {
				return false
			}
		}
		return true
	default:
		return canonical(a) == canonical(b)
	}
}

// sameSetCanonical compares two slices as multisets of canonical JSON strings.
func sameSetCanonical(a, b []any) bool {
	ca := make([]string, len(a))
	cb := make([]string, len(b))
	for i := range a {
		ca[i] = canonical(a[i])
	}
	for i := range b {
		cb[i] = canonical(b[i])
	}
	sort.Strings(ca)
	sort.Strings(cb)
	for i := range ca {
		if ca[i] != cb[i] {
			return false
		}
	}
	return true
}

// canonical returns a stable JSON encoding (Go sorts object keys).
func canonical(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
