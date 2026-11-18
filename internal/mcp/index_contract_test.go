package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/linden-project/linden-index/corpus"
	"github.com/linden-project/linden-index/index"
)

func TestIndexContractAfterModuleSplit(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 7, Count: 10}); err != nil {
		t.Fatal(err)
	}
	g, _, err := index.Build(root)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "lindenIndex")
	if err := index.Emit(g, out); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		"_index_taxonomies.json",
		"_index_docs_starred.json",
		"_index_docs_with_props.json",
		"_index_docs_with_title.json",
		"_index_docs_tasks_count.json",
		"_index_taxonomies_starred.json",
		"_index_terms_starred.json",
		"_indexer_info.json",
	} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("%s missing: %v", name, err)
		}
	}

	var info struct {
		ProductName        string `json:"product_name"`
		IndexFormatVersion int    `json:"index_format_version"`
	}
	b, err := os.ReadFile(filepath.Join(out, "_indexer_info.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &info); err != nil {
		t.Fatal(err)
	}
	if info.ProductName != "lindexer" {
		t.Fatalf("product_name = %q", info.ProductName)
	}
	if info.IndexFormatVersion < 1 {
		t.Fatalf("index_format_version = %d", info.IndexFormatVersion)
	}
}
