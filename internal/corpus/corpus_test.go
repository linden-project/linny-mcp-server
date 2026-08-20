package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func genInto(t *testing.T, opts Options) string {
	t.Helper()
	dir := t.TempDir()
	if err := Generate(dir, opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return dir
}

func TestDeterministic(t *testing.T) {
	opts := Options{Seed: 42, Count: 50, EnableEdgeCases: true}
	a := genInto(t, opts)
	b := genInto(t, opts)

	fa := readAll(t, a)
	fb := readAll(t, b)

	if len(fa) != len(fb) {
		t.Fatalf("file count differs: %d vs %d", len(fa), len(fb))
	}
	for name, ca := range fa {
		cb, ok := fb[name]
		if !ok {
			t.Fatalf("file %s missing from second run", name)
		}
		if ca != cb {
			t.Fatalf("file %s differs between identical-seed runs", name)
		}
	}
}

func TestFlatAndParseable(t *testing.T) {
	dir := genInto(t, Options{Seed: 7, Count: 30, EnableEdgeCases: false})
	content := filepath.Join(dir, ContentDir)
	entries, err := os.ReadDir(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 30 {
		t.Fatalf("expected 30 records, got %d", len(entries))
	}
	for _, e := range entries {
		if e.IsDir() {
			t.Fatalf("content dir must be flat, found subdir %s", e.Name())
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			t.Fatalf("unexpected file %s", e.Name())
		}
		b, _ := os.ReadFile(filepath.Join(content, e.Name()))
		if !strings.HasPrefix(string(b), "---\n") {
			t.Fatalf("record %s missing front matter", e.Name())
		}
	}
}

func TestConfigConsistency(t *testing.T) {
	dir := genInto(t, Options{Seed: 3, Count: 40, EnableEdgeCases: true})
	// Hugo config declares the taxonomies.
	hc, err := os.ReadFile(filepath.Join(dir, HugoConfig))
	if err != nil {
		t.Fatal(err)
	}
	for _, tax := range []string{"tags", "projects", "customer", "type", "subject"} {
		if !strings.Contains(string(hc), tax) {
			t.Fatalf("hugo config missing taxonomy %q", tax)
		}
	}
	// Every taxonomy used has an L1 config file.
	cfg := filepath.Join(dir, ConfigDir)
	for _, tax := range []string{"tags", "subject"} {
		if _, err := os.Stat(filepath.Join(cfg, "L1-CONF-TAX-"+tax+".yml")); err != nil {
			t.Fatalf("missing L1 config for %q: %v", tax, err)
		}
	}
}

func TestEdgeCasesPresent(t *testing.T) {
	dir := genInto(t, Options{Seed: 5, Count: 10, EnableEdgeCases: true})
	all := readAll(t, dir)

	joinAll := strings.Builder{}
	for _, c := range all {
		joinAll.WriteString(c)
		joinAll.WriteString("\n")
	}
	corpus := joinAll.String()

	if !strings.Contains(corpus, "<<<<<<<") {
		t.Error("expected a committed conflict marker in the corpus")
	}
	if !strings.Contains(corpus, "AKIA") || !strings.Contains(corpus, "BEGIN RSA PRIVATE KEY") {
		t.Error("expected fake-secret fixtures in the corpus")
	}
	if _, ok := all[filepath.Join(ContentDir, "work_and_health.md")]; !ok {
		t.Error("expected the work_and_health scope-intersection fixture")
	}
	if _, ok := all[filepath.Join(ContentDir, "malformed_yaml.md")]; !ok {
		t.Error("expected a malformed_yaml record")
	}
}

// readAll returns a map of relative path -> content for all files under dir.
func readAll(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		out[rel] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
