// Package alert delivers out-of-band notifications (via self-hosted ntfy) and a
// monitor that fires on transitions into git-safety degraded mode. Alerts are
// NEVER written into the corpus — that is the wrong channel when the corpus is
// what's broken.
package alert

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Alerter delivers a single notification.
type Alerter interface {
	Alert(ctx context.Context, title, body string) error
}

// NopAlerter drops alerts. Used when no ntfy topic is configured.
type NopAlerter struct{}

// Alert does nothing.
func (NopAlerter) Alert(context.Context, string, string) error { return nil }

// NtfyAlerter POSTs a notification to a self-hosted ntfy topic URL.
type NtfyAlerter struct {
	TopicURL string       // e.g. https://ntfy.example.com/linny-mcp
	Client   *http.Client // defaults to http.DefaultClient
}

// Alert issues an HTTP POST with the body and a Title header (ntfy convention).
func (n NtfyAlerter) Alert(ctx context.Context, title, body string) error {
	client := n.Client
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.TopicURL, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Title", title)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// StatusFunc reports whether the tree is currently degraded, with a reason.
type StatusFunc func() (degraded bool, reason string)

// DegradedMonitor fires an alert on the transition into degraded mode (once),
// and a recovery notice when the tree becomes clean again. It is safe for
// concurrent Check calls.
type DegradedMonitor struct {
	status  StatusFunc
	alerter Alerter
	name    string // included in the alert title (e.g. the notebook name)

	mu          sync.Mutex
	wasDegraded bool
}

// NewDegradedMonitor builds a monitor.
func NewDegradedMonitor(name string, status StatusFunc, alerter Alerter) *DegradedMonitor {
	if alerter == nil {
		alerter = NopAlerter{}
	}
	return &DegradedMonitor{status: status, alerter: alerter, name: name}
}

// Check samples the status and alerts on a transition. It returns the current
// degraded state.
func (m *DegradedMonitor) Check(ctx context.Context) bool {
	degraded, reason := m.status()

	m.mu.Lock()
	transitionedToDegraded := degraded && !m.wasDegraded
	transitionedToClean := !degraded && m.wasDegraded
	m.wasDegraded = degraded
	m.mu.Unlock()

	switch {
	case transitionedToDegraded:
		_ = m.alerter.Alert(ctx, "linny-mcp degraded: "+m.name,
			"The notebook working tree is degraded (read-only): "+reason)
	case transitionedToClean:
		_ = m.alerter.Alert(ctx, "linny-mcp recovered: "+m.name,
			"The notebook working tree is clean again; writes are enabled.")
	}
	return degraded
}

// Run polls Check every interval until ctx is cancelled. Intended to run in a
// background goroutine.
func (m *DegradedMonitor) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	m.Check(ctx) // sample immediately
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.Check(ctx)
		}
	}
}
