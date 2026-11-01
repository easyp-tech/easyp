// Package parallel runs bounded independent tasks with deterministic error order.
package parallel

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
)

// Run runs count tasks with at most limit workers and waits for all of them.
// Each task owns its index, so callers can write independent indexed results
// without locking. Errors are returned in index order, not completion order.
// Cancellation stops new tasks and takes precedence after active tasks finish.
// A non-positive limit uses the number of available CPUs.
func Run(ctx context.Context, count, limit int, task func(index int) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if count <= 0 {
		return nil
	}
	if limit <= 0 {
		limit = runtime.NumCPU()
	}
	workers := min(count, limit)
	errors := make([]error, count)
	var next atomic.Int64
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			for {
				if ctx.Err() != nil {
					return
				}
				index := int(next.Add(1) - 1)
				if index >= count {
					return
				}
				if ctx.Err() != nil {
					return
				}
				errors[index] = task(index)
			}
		}()
	}
	group.Wait()
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
