package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// makeCorpus builds a tiny corpus with content + lindenConfig + a .git dir.
func makeCorpus(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("content/note.md", "---\ntitle: Note\n---\noriginal body\n")
	mk("content/other.md", "---\ntitle: Other\n---\nkeep me\n")
	mk("lindenConfig/L1-CONF-TAX-tags.yml", "title: Tags\n")
	mk(".git/HEAD", "ref: refs/heads/main\n")      // must NOT be backed up
	mk("lindenIndex/_index_taxonomies.json", "[]") // disposable, must NOT be backed up
	return root
}

func TestBackupContentsAndExclusions(t *testing.T) {
	root := makeCorpus(t)
	var buf bytes.Buffer
	if err := Backup(root, &buf); err != nil {
		t.Fatal(err)
	}
	names := tarEntries(t, buf.Bytes())
	want := []string{"content/note.md", "content/other.md", "lindenConfig/L1-CONF-TAX-tags.yml"}
	for _, w := range want {
		if !names[w] {
			t.Errorf("backup missing %q (have %v)", w, keys(names))
		}
	}
	for _, bad := range []string{".git/HEAD", "lindenIndex/_index_taxonomies.json"} {
		if names[bad] {
			t.Errorf("backup must not include %q", bad)
		}
	}
}

func TestRestoreRecoversDeletedAndMutated(t *testing.T) {
	root := makeCorpus(t)
	var buf bytes.Buffer
	if err := Backup(root, &buf); err != nil {
		t.Fatal(err)
	}

	// Delete one record and mutate another after the backup.
	if err := os.Remove(filepath.Join(root, "content", "note.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "content", "other.md"), []byte("CORRUPTED"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Restore(bytes.NewReader(buf.Bytes()), root); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(root, "content", "note.md"))
	if err != nil {
		t.Fatalf("deleted record not recovered: %v", err)
	}
	if string(got) != "---\ntitle: Note\n---\noriginal body\n" {
		t.Fatalf("recovered content mismatch: %q", got)
	}
	got, _ = os.ReadFile(filepath.Join(root, "content", "other.md"))
	if string(got) != "---\ntitle: Other\n---\nkeep me\n" {
		t.Fatalf("mutated record not restored: %q", got)
	}
}

func TestRestoreRejectsTraversal(t *testing.T) {
	// Hand-craft a malicious archive with a ../ escape.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte("pwned")
	_ = tw.WriteHeader(&tar.Header{Name: "../escape.md", Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg})
	_, _ = tw.Write(body)
	_ = tw.Close()
	_ = gz.Close()

	root := t.TempDir()
	if err := Restore(bytes.NewReader(buf.Bytes()), filepath.Join(root, "corpus")); err == nil {
		t.Fatal("expected restore to reject a traversing entry")
	}
	if _, err := os.Stat(filepath.Join(root, "escape.md")); !os.IsNotExist(err) {
		t.Fatal("traversing entry escaped the target directory")
	}
}

func tarEntries(t *testing.T, gzData []byte) map[string]bool {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(gzData))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	out := map[string]bool{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if hdr.Typeflag == tar.TypeReg {
			out[hdr.Name] = true
		}
	}
	return out
}

func keys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
