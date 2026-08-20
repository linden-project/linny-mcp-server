package gitsafe

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// HashFile returns the hex SHA-256 of the file at path, or "" (with a nil error)
// when the file does not exist. The empty hash therefore denotes "no such file",
// which callers use to express create-if-absent semantics.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// AtomicWrite writes data to path atomically: a temp file in the same directory
// is written, fsync'd, and renamed into place, then the directory is fsync'd so
// the rename is durable. Consumers never observe a partial file.
func AtomicWrite(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Clean up the temp file on any failure.
	defer func() {
		if err != nil {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmpName, perm); err != nil {
		return err
	}
	if err = os.Rename(tmpName, path); err != nil {
		return err
	}
	return fsyncDir(dir)
}

// fsyncDir flushes a directory entry so a rename survives a crash.
func fsyncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = d.Close() }()
	// Directory fsync is not supported on every platform/filesystem; ignore
	// EINVAL-style failures rather than failing an otherwise-good write.
	if err := d.Sync(); err != nil {
		return nil //nolint:nilerr // best-effort durability; the rename already happened
	}
	return nil
}

// WriteIfUnchanged writes data to path atomically only if the file's current
// content hash equals expectedHash (use "" for "must not yet exist"). Otherwise
// it returns a retryable *StaleWriteError and leaves the file untouched.
func WriteIfUnchanged(path string, data []byte, expectedHash string, perm os.FileMode) error {
	current, err := HashFile(path)
	if err != nil {
		return err
	}
	if current != expectedHash {
		return &StaleWriteError{Path: path}
	}
	return AtomicWrite(path, data, perm)
}
