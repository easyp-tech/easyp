package parallel

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunBoundsWorkersAndExecutesEachTaskOnce(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		limit int
	}{
		{name: "serial", limit: 1},
		{name: "two workers", limit: 2},
		{name: "more workers than tasks", limit: 20},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				const count = 6
				started := make(chan int, count)
				release := make(chan struct{})
				done := make(chan error, 1)
				var calls [count]atomic.Int64
				go func() {
					done <- Run(t.Context(), count, tt.limit, func(index int) error {
						calls[index].Add(1)
						started <- index
						<-release
						return nil
					})
				}()
				synctest.Wait()
				assert.Len(t, started, min(count, tt.limit))
				close(release)
				require.NoError(t, <-done)
				for index := range calls {
					assert.EqualValues(t, 1, calls[index].Load(), "task %d", index)
				}
			})
		})
	}
}

func TestRunReturnsErrorInTaskOrder(t *testing.T) {
	t.Parallel()
	first := errors.New("first task")
	second := errors.New("second task")
	completedSecond := make(chan struct{})
	err := Run(t.Context(), 3, 2, func(index int) error {
		switch index {
		case 0:
			<-completedSecond
			return first
		case 1:
			close(completedSecond)
			return second
		default:
			return nil
		}
	})
	require.ErrorIs(t, err, first)
}

func TestRunCancellationStopsSchedulingAndJoinsWorkers(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		release := make(chan struct{})
		done := make(chan error, 1)
		var calls atomic.Int64
		go func() {
			done <- Run(ctx, 10, 2, func(int) error {
				calls.Add(1)
				<-release
				return nil
			})
		}()
		synctest.Wait()
		require.EqualValues(t, 2, calls.Load())
		cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("Run returned while its tasks were still active")
		default:
		}
		close(release)
		require.ErrorIs(t, <-done, context.Canceled)
		assert.EqualValues(t, 2, calls.Load())
	})
}

func TestRunCanceledAndEmpty(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := Run(ctx, 3, 2, func(int) error {
		t.Error("canceled run started a task")
		return nil
	})
	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, Run(t.Context(), 0, 0, nil))
}
