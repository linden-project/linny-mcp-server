// Package backup provides a verified snapshot/restore of a notebook's
// source-of-truth data (the content dir and lindenConfig). The disposable
// index/state and VCS directories are excluded — they are rebuildable. Restore
// sanitizes paths so a malicious archive cannot escape the target directory.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// backedUpDirs are the source-of-truth subdirectories included in a snapshot.
var backedUpDirs = []string{"content", "lindenConfig"}

// Backup writes a tar.gz snapshot of the corpus's source-of-truth data to w.
func Backup(corpusRoot string, w io.Writer) error {
	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)

	for _, sub := range backedUpDirs {
		base := filepath.Join(corpusRoot, sub)
		if _, err := os.Stat(base); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(corpusRoot, p)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = rel
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			_, err = io.Copy(tw, f)
			return err
		})
		if err != nil {
			_ = tw.Close()
			_ = gz.Close()
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// Restore extracts a tar.gz snapshot into corpusRoot. Entries whose paths escape
// corpusRoot are rejected.
func Restore(r io.Reader, corpusRoot string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)

	cleanRoot := filepath.Clean(corpusRoot)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := safeJoin(cleanRoot, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := writeFile(tr, target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		default:
			// Skip symlinks and other special types — the corpus is plain files.
		}
	}
}

// safeJoin joins name onto root, rejecting paths that escape root.
func safeJoin(root, name string) (string, error) {
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("backup: refusing absolute path in archive: %q", name)
	}
	target := filepath.Join(root, filepath.FromSlash(name))
	cleaned := filepath.Clean(target)
	if cleaned != root && !strings.HasPrefix(cleaned, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("backup: archive entry escapes target: %q", name)
	}
	return cleaned, nil
}

func writeFile(r io.Reader, path string, mode os.FileMode) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
