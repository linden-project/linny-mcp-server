// Package hugoref builds the Hugo reference index for a corpus, so
// `lindexer verify --hugo` can diff our index against Hugo's actual output. The
// reference Hugo site (layouts + config) is vendored from linny-notebook-template
// and embedded; the corpus's content and lindenConfig are combined with it in a
// temp working directory (the real corpus is never mutated).
package hugoref

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed all:site
var siteFS embed.FS

// BuildReference assembles a Hugo site from the embedded reference layouts/config
// plus the corpus's content and lindenConfig, runs `hugo`, and returns the
// produced index directory and a cleanup func. It requires the `hugo` binary.
func BuildReference(corpusRoot string) (indexDir string, cleanup func(), err error) {
	hugo, err := exec.LookPath("hugo")
	if err != nil {
		return "", nil, fmt.Errorf("verify --hugo: hugo binary not found on PATH: %w", err)
	}

	work, err := os.MkdirTemp("", "linny-hugoref-*")
	if err != nil {
		return "", nil, err
	}
	out, err := os.MkdirTemp("", "linny-hugoout-*")
	if err != nil {
		_ = os.RemoveAll(work)
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(work); _ = os.RemoveAll(out) }

	if err := writeEmbedded(work); err != nil {
		cleanup()
		return "", nil, err
	}
	for _, sub := range []string{"content", "lindenConfig"} {
		if err := copyTree(filepath.Join(corpusRoot, sub), filepath.Join(work, sub)); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("verify --hugo: copying %s: %w", sub, err)
		}
	}

	cmd := exec.Command(hugo, "--source", work, "--destination", out, "--logLevel", "error")
	if b, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("verify --hugo: hugo failed: %w\n%s", err, b)
	}
	return out, cleanup, nil
}

// writeEmbedded materializes the embedded site/ tree into dst (stripping "site/").
func writeEmbedded(dst string) error {
	return fs.WalkDir(siteFS, "site", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("site", p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := siteFS.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

// copyTree recursively copies src into dst (files only; regular files).
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		in, err := os.Open(p)
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		f, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, in); err != nil {
			_ = f.Close()
			return err
		}
		return f.Close()
	})
}
