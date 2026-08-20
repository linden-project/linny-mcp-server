// Package gitsafe inspects the git working tree before every write, enters
// degraded read-only mode on a conflicted or in-progress tree, performs atomic
// writes (temp+fsync+rename), and enforces optimistic concurrency via
// read-time content hashes. It never takes ownership of the external git-sync.
package gitsafe
