package alert

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// fakeAlerter records the alerts it receives.
type fakeAlerter struct {
	mu     sync.Mutex
	titles []string
}

func (f *fakeAlerter) Alert(_ context.Context, title, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.titles = append(f.titles, title)
	return nil
}

func (f *fakeAlerter) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.titles)
}

func TestMonitorAlertsOncePerDegradation(t *testing.T) {
	degraded := false
	fa := &fakeAlerter{}
	m := NewDegradedMonitor("nb", func() (bool, string) { return degraded, "conflict" }, fa)
	ctx := context.Background()

	m.Check(ctx) // clean -> no alert
	if fa.count() != 0 {
		t.Fatalf("no alert expected while clean, got %d", fa.count())
	}

	degraded = true
	m.Check(ctx) // transition -> 1 alert
	m.Check(ctx) // still degraded -> no new alert
	m.Check(ctx)
	if fa.count() != 1 {
		t.Fatalf("expected exactly one degraded alert, got %d (%v)", fa.count(), fa.titles)
	}

	degraded = false
	m.Check(ctx) // recovery -> 1 more alert
	if fa.count() != 2 {
		t.Fatalf("expected a recovery alert, got %d (%v)", fa.count(), fa.titles)
	}
}

func TestNtfyAlerterPosts(t *testing.T) {
	var gotTitle, gotBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTitle = r.Header.Get("Title")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	a := NtfyAlerter{TopicURL: ts.URL, Client: ts.Client()}
	if err := a.Alert(context.Background(), "hello", "world"); err != nil {
		t.Fatal(err)
	}
	if gotTitle != "hello" || gotBody != "world" {
		t.Fatalf("ntfy POST mismatch: title=%q body=%q", gotTitle, gotBody)
	}
}

func TestNopAlerter(t *testing.T) {
	if err := (NopAlerter{}).Alert(context.Background(), "t", "b"); err != nil {
		t.Fatalf("nop alerter should never error: %v", err)
	}
}
