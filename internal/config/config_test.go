package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"ok", Config{Notebooks: []Notebook{{Name: "a", CorpusPath: "/p"}}}, false},
		{"no notebooks", Config{}, true},
		{"empty name", Config{Notebooks: []Notebook{{Name: "", CorpusPath: "/p"}}}, true},
		{"dup name", Config{Notebooks: []Notebook{{Name: "a", CorpusPath: "/p"}, {Name: "a", CorpusPath: "/q"}}}, true},
		{"empty corpus", Config{Notebooks: []Notebook{{Name: "a", CorpusPath: ""}}}, true},
	}
	for _, c := range cases {
		if err := c.cfg.Validate(); (err != nil) != c.wantErr {
			t.Errorf("%s: Validate() err=%v wantErr=%v", c.name, err, c.wantErr)
		}
	}
}

func TestUnsetHostnameValidates(t *testing.T) {
	cfg := Config{Notebooks: []Notebook{{Name: "a", CorpusPath: "/p"}}}
	if cfg.PublicHostname != "" {
		t.Fatal("hostname should default empty")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unset hostname must validate: %v", err)
	}
}

func TestLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	in := Config{
		ListenAddress:  "127.0.0.1",
		Port:           9000,
		TokensFile:     "/run/secrets/tokens",
		PublicHostname: "secondbrain.example.com",
		Notebooks: []Notebook{
			{Name: "personal", CorpusPath: "/n/personal", StateDir: "/s/personal"},
			{Name: "business", CorpusPath: "/n/business", StateDir: "/s/business"},
		},
	}
	b, _ := json.Marshal(in)
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.PublicHostname != "secondbrain.example.com" || len(got.Notebooks) != 2 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestFromFlagsSingleNotebook(t *testing.T) {
	cfg, err := FromFlags("127.0.0.1", 8765, "/tokens", "info", false, false, "/corpus", "/state")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Notebooks) != 1 || cfg.Notebooks[0].Name != "default" {
		t.Fatalf("expected one 'default' notebook, got %+v", cfg.Notebooks)
	}
	if cfg.Notebooks[0].CorpusPath != "/corpus" {
		t.Fatalf("corpus path = %q", cfg.Notebooks[0].CorpusPath)
	}
}

func TestResolve(t *testing.T) {
	cfg := Config{Notebooks: []Notebook{
		{Name: "personal", CorpusPath: "/n/personal"},
		{Name: "business", CorpusPath: "/n/business"},
	}}
	// Default → first.
	if nb, err := cfg.Resolve(""); err != nil || nb.Name != "personal" {
		t.Fatalf("default resolve = %+v err=%v, want personal", nb, err)
	}
	// By name.
	if nb, err := cfg.Resolve("business"); err != nil || nb.CorpusPath != "/n/business" {
		t.Fatalf("named resolve = %+v err=%v", nb, err)
	}
	// Unknown.
	if _, err := cfg.Resolve("nope"); err == nil {
		t.Fatal("expected error for unknown notebook")
	}
}
