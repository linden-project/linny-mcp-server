// Package audit writes an append-only audit log, kept OUTSIDE the corpus (in
// stateDir). Every write operation is recorded with a diff so there is a durable,
// tamper-evident record of what an agent changed. The log is operator-facing and
// is intentionally NOT redacted — egress to the agent is redacted elsewhere.
package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry is one audit record. It is serialized as a single JSON line.
type Entry struct {
	Time     string `json:"time"`           // RFC3339; filled by Append if empty
	Identity string `json:"identity"`       // token name
	Tool     string `json:"tool"`           // e.g. create_doc, archive
	Slug     string `json:"slug,omitempty"` // target document
	Diff     string `json:"diff,omitempty"` // what changed
	Outcome  string `json:"outcome"`        // ok | denied | error
}

// Log is an append-only audit log.
type Log struct {
	mu  sync.Mutex
	f   *os.File
	now func() time.Time
}

// Open opens (creating if needed) an append-only audit log at path. The parent
// directory is created. The file is only ever appended to.
func Open(path string) (*Log, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &Log{f: f, now: time.Now}, nil
}

// Append writes one entry as a JSON line. If Time is empty it is set to now.
func (l *Log) Append(e Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.Time == "" {
		e.Time = l.now().UTC().Format(time.RFC3339Nano)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if _, err := l.f.Write(append(b, '\n')); err != nil {
		return fmt.Errorf("audit: append: %w", err)
	}
	return nil
}

// Close closes the underlying file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Close()
}
