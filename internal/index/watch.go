package index

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DefaultDebounce is the quiet period after the last change before a rebuild.
const DefaultDebounce = 300 * time.Millisecond

// debounce reads from in and, after a quiet period of wait, calls fire once.
// A burst of signals within the window collapses to a single fire. It returns
// when ctx is done.
func debounce(ctx context.Context, in <-chan struct{}, wait time.Duration, fire func()) {
	var timer *time.Timer
	var timerC <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return
		case <-in:
			if timer == nil {
				timer = time.NewTimer(wait)
			} else {
				timer.Reset(wait)
			}
			timerC = timer.C
		case <-timerC:
			fire()
			timerC = nil
		}
	}
}

// Watch watches the corpus's content dir and lindenConfig for changes and calls
// onChange once per debounced burst, until ctx is cancelled. wait <= 0 uses
// DefaultDebounce.
func Watch(ctx context.Context, corpusRoot string, wait time.Duration, onChange func()) error {
	if wait <= 0 {
		wait = DefaultDebounce
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()

	contentDir, _, _ := loadNotebook(corpusRoot)
	for _, sub := range []string{contentDir, lindenConfigRel} {
		dir := filepath.Join(corpusRoot, sub)
		if _, err := os.Stat(dir); err == nil {
			if err := w.Add(dir); err != nil {
				return err
			}
		}
	}

	signals := make(chan struct{}, 1)
	go debounce(ctx, signals, wait, onChange)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-w.Events:
			if !ok {
				return nil
			}
			// Non-blocking notify; the debouncer coalesces.
			select {
			case signals <- struct{}{}:
			default:
			}
		case _, ok := <-w.Errors:
			if !ok {
				return nil
			}
		}
	}
}
