package index

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
)

func TestDebounceCoalescesBurst(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	in := make(chan struct{}, 8)

	var mu sync.Mutex
	fires := 0
	go debounce(ctx, in, 40*time.Millisecond, func() {
		mu.Lock()
		fires++
		mu.Unlock()
	})

	// Rapid burst within the window.
	for i := 0; i < 5; i++ {
		in <- struct{}{}
		time.Sleep(5 * time.Millisecond)
	}
	// Wait comfortably past the debounce window.
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	got := fires
	mu.Unlock()
	if got != 1 {
		t.Fatalf("expected exactly one fire for a burst, got %d", got)
	}
}

func TestWatchTriggersRebuild(t *testing.T) {
	root := t.TempDir()
	if err := corpus.Generate(root, corpus.Options{Seed: 3, Count: 5, EnableEdgeCases: false}); err != nil {
		t.Fatal(err)
	}

	var rebuilds atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- Watch(ctx, root, 50*time.Millisecond, func() { rebuilds.Add(1) })
	}()

	// Give the watcher a moment to register, then modify a record.
	time.Sleep(150 * time.Millisecond)
	if err := os.WriteFile(filepath.Join(root, "content", "trigger.md"),
		[]byte("---\ntitle: Trigger\n---\nhi\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Poll for the rebuild callback (generous timeout for CI).
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if rebuilds.Load() >= 1 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if rebuilds.Load() < 1 {
		t.Fatal("expected a rebuild after writing a record")
	}
	cancel()
	<-done
}
