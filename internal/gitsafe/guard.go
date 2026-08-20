package gitsafe

import "fmt"

// Guard gates writes to a single notebook corpus based on the live git
// working-tree state. It holds no sticky state: every call re-inspects, so the
// server enters and leaves degraded mode automatically.
type Guard struct {
	root          string
	forceReadOnly bool
}

// NewGuard returns a Guard for the corpus at root. If forceReadOnly is true,
// writes are always refused regardless of tree state.
func NewGuard(root string, forceReadOnly bool) *Guard {
	return &Guard{root: root, forceReadOnly: forceReadOnly}
}

// State inspects the working tree now and returns its state.
func (g *Guard) State() TreeState { return inspect(g.root) }

// ForcedReadOnly reports whether the guard was started in forced read-only mode.
func (g *Guard) ForcedReadOnly() bool { return g.forceReadOnly }

// EnsureWritable returns nil when a write may proceed, or a retryable
// DegradedError otherwise.
func (g *Guard) EnsureWritable() error {
	if g.forceReadOnly {
		return &DegradedError{Forced: true, Reason: "server is in forced read-only mode"}
	}
	st := g.State()
	if !st.Clean {
		return &DegradedError{State: st, Reason: st.Reason}
	}
	return nil
}

// DegradedError is returned when a write is refused because the tree is not
// clean-and-merged (or the server is forced read-only). It is retryable: once
// the tree recovers, the same write will succeed.
type DegradedError struct {
	State  TreeState
	Forced bool
	Reason string
}

func (e *DegradedError) Error() string {
	return fmt.Sprintf("refused: notebook is in degraded read-only mode (%s); retry once the working tree is clean", e.Reason)
}

// Retryable reports that the caller may retry later.
func (e *DegradedError) Retryable() bool { return true }

// StaleWriteError is returned when an optimistic write finds the file changed
// underneath the caller's read. It is retryable after a fresh read.
type StaleWriteError struct {
	Path string
}

func (e *StaleWriteError) Error() string {
	return fmt.Sprintf("stale write: %s changed since it was read; re-read and retry", e.Path)
}

// Retryable reports that the caller may retry after re-reading.
func (e *StaleWriteError) Retryable() bool { return true }
