package gitmodules

import (
	"context"
	"fmt"
	"os"
	"sync"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

// A cache handle belongs to one command. Only its successfully installed,
// immutable snapshots share verification and file inventories within that run.
type cacheRunState struct {
	mu        sync.Mutex
	snapshots map[string]*cachedSnapshot
}

type cachedSnapshot struct {
	gate     chan struct{}
	verified bool
	files    []string
	identity os.FileInfo
}

func (snapshot *cachedSnapshot) lock(ctx context.Context) error {
	select {
	case snapshot.gate <- struct{}{}:
		if err := ctx.Err(); err != nil {
			snapshot.unlock()
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (snapshot *cachedSnapshot) unlock() { <-snapshot.gate }

// accept records a snapshot hashed before its atomic publication. The caller
// holds the per-snapshot gate; no second file-content pass is needed.
func (snapshot *cachedSnapshot) accept(installed string, files []string) error {
	info, err := os.Lstat(installed)
	if err != nil {
		return fmt.Errorf("Lstat: %w", err)
	}
	// Persistent verification stamps require matching fingerprints before and
	// after the content hash. Publication has only the already computed hash;
	// a later command verifies the installed tree before writing a stamp.
	snapshot.files, snapshot.identity, snapshot.verified = files, info, true
	return nil
}

func (c *Cache) runState() *cacheRunState {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	if c.state == nil {
		c.state = &cacheRunState{snapshots: make(map[string]*cachedSnapshot)}
	}
	return c.state
}

func (c *Cache) scoped() *Cache {
	return &Cache{root: c.root, bsrResolver: c.bsrResolver, logger: c.logger, proxy: c.proxy, state: c.runState()}
}

func (c *Cache) snapshotState(installed string) *cachedSnapshot {
	state := c.runState()
	state.mu.Lock()
	defer state.mu.Unlock()
	if snapshot := state.snapshots[installed]; snapshot != nil {
		return snapshot
	}
	snapshot := &cachedSnapshot{gate: make(chan struct{}, 1)}
	state.snapshots[installed] = snapshot
	return snapshot
}

// SourceFiles returns the sorted logical file inventory of a snapshot verified
// by this cache handle. Paths are relative to its installation directory and
// include metadata so consumers can preserve nested module ownership.
// The returned slice is independent of the command's immutable cached inventory.
// A nil slice means this handle has no verified inventory and callers must use
// their ordinary source traversal. A verified empty snapshot returns an empty,
// non-nil slice. This method never verifies or trusts a directory's existence.
func (c *Cache) SourceFiles(entry v1.LockedModule) ([]string, error) {
	snapshot := c.snapshotState(v1ModuleCachePath(c.root, entry))
	if err := snapshot.lock(context.Background()); err != nil {
		return nil, fmt.Errorf("lock: %w", err)
	}
	defer snapshot.unlock()
	if !snapshot.verified {
		return nil, nil
	}
	files := make([]string, len(snapshot.files))
	copy(files, snapshot.files)
	return files, nil
}
