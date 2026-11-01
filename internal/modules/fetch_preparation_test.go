package modules

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestResolvePreparesBoundedRevisionsAndConsumesInOrder(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		const count = 12
		started := make(chan FetchRequest, count)
		release := make(chan struct{})
		backend := &preparedTestSource{started: started, release: release}
		root := v1.Module{}
		var want []string
		for index := range count {
			name := "example.test/" + string(rune('a'+index))
			root.Requires = append(root.Requires, v1.Requirement{Module: name, Version: "v1.0.0"})
			want = append(want, name+"@v1.0.0")
		}
		root.Requires = append(root.Requires, root.Requires[0], v1.Requirement{Module: root.Requires[0].Module, Version: versionlessCommitA})
		done := make(chan error, 1)
		var lock v1.Lock
		go func() {
			var err error
			lock, err = Resolve(t.Context(), root, backend, nil)
			done <- err
		}()
		synctest.Wait()
		assert.Len(t, started, 8, "known independent revisions should fill the fetch worker pool")
		close(release)
		require.NoError(t, <-done)
		assert.EqualValues(t, count, backend.preparations.Load())
		assert.Zero(t, backend.fetches.Load(), "prepared results must be consumed without refetching")
		assert.Equal(t, want, backend.consumed)
		require.Len(t, lock.Modules, count)
		assert.Equal(t, "v1.0.0", lock.Modules[0].Version)
		assert.Equal(t, versionlessCommitA, lock.Modules[0].Commit)
	})
}

func TestResolvePreparationCancellationJoinsWorkers(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		started := make(chan FetchRequest, 12)
		release := make(chan struct{})
		backend := &preparedTestSource{started: started, release: release}
		root := v1.Module{}
		for index := range 12 {
			root.Requires = append(root.Requires, v1.Requirement{Module: "example.test/" + string(rune('a'+index)), Version: "v1.0.0"})
		}
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			_, err := Resolve(ctx, root, backend, nil)
			done <- err
		}()
		synctest.Wait()
		assert.Len(t, started, 8)
		cancel()
		synctest.Wait()
		select {
		case <-done:
			t.Error("Resolve returned while preparation workers were active")
		default:
		}
		close(release)
		require.ErrorIs(t, <-done, context.Canceled)
		assert.EqualValues(t, 8, backend.preparations.Load())
		assert.Empty(t, backend.consumed)
	})
}

func TestResolvePreparedErrorsFollowRequestOrder(t *testing.T) {
	t.Parallel()
	first := errors.New("first module unavailable")
	second := errors.New("second module unavailable")
	secondPrepared := make(chan struct{})
	backend := &preparedTestSource{prepare: func(request FetchRequest) (PreparedFetch, error) {
		if request.Source == "a" {
			<-secondPrepared
			return nil, first
		}
		close(secondPrepared)
		return nil, second
	}}
	_, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}, {Module: "b", Version: "v1.0.0"}}}, backend, nil)
	require.ErrorIs(t, err, first)
}

func TestResolveDoesNotPrepareThroughUnadvertisedSourceWrapper(t *testing.T) {
	t.Parallel()
	backend := &preparedTestSource{}
	source := &overriddenPreparedSource{preparedTestSource: backend}
	_, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}}}, source, nil)
	require.NoError(t, err)
	assert.Equal(t, 1, source.calls)
	assert.Zero(t, backend.preparations.Load(), "embedding a source must not bypass overridden Fetch methods")
}

func TestResolveUncomparablePreparationSourceRemainsSerial(t *testing.T) {
	t.Parallel()
	source := uncomparablePreparedSource{values: []string{"a"}}
	_, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}}}, source, nil)
	require.NoError(t, err)
}

type uncomparablePreparedSource struct{ values []string }

func (source uncomparablePreparedSource) PreparationSource() Source { return source }

func (uncomparablePreparedSource) PrepareFetch(context.Context, FetchRequest) (PreparedFetch, error) {
	return nil, errors.New("uncomparable source cannot advertise receiver identity")
}

func (uncomparablePreparedSource) Fetch(ctx context.Context, name, version string) (Fetched, error) {
	return (&preparedTestSource{}).Fetch(ctx, name, version)
}

type overriddenPreparedSource struct {
	*preparedTestSource
	calls int
}

func (source *overriddenPreparedSource) Fetch(ctx context.Context, name, version string) (Fetched, error) {
	source.calls++
	return source.preparedTestSource.Fetch(ctx, name, version)
}

func TestImportRootPreparationDefersRootBookkeeping(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		started := make(chan FetchRequest, 2)
		release := make(chan struct{})
		backend := &preparedRootTestSource{preparedTestSource: preparedTestSource{started: started, release: release}}
		source := newImportRootSource(backend, v1.Lock{}, map[string][]string{"a": {"api"}})
		done := make(chan error, 1)
		go func() {
			_, err := Resolve(t.Context(), v1.Module{Requires: []v1.Requirement{{Module: "a", Version: "v1.0.0"}, {Module: "b", Version: "v1.0.0"}}}, source, nil)
			done <- err
		}()
		synctest.Wait()
		assert.Len(t, started, 2)
		assert.Empty(t, source.fetched, "preparation must not mutate the root-resolution maps")
		close(release)
		require.NoError(t, <-done)
		assert.Equal(t, []string{"a@v1.0.0", "b@v1.0.0"}, backend.consumed)
		assert.Len(t, source.fetched, 2)
		if len(started) == 2 {
			requests := []FetchRequest{<-started, <-started}
			for _, request := range requests {
				assert.True(t, request.Inspect)
				if request.Source == "a" {
					assert.Equal(t, []string{"api"}, request.Roots)
				}
			}
		}
	})
}

func TestImportRootPreparationChecksTagBeforeRetainedScope(t *testing.T) {
	t.Parallel()
	old := v1.LockedModule{Source: "a", Version: "v1.0.0", Commit: strings.Repeat("b", 40), Hash: versionlessHash, Roots: []string{"api"}}
	backend := &preparedRootTestSource{}
	source := newImportRootSource(backend, v1.Lock{Modules: []v1.LockedModule{old}}, nil)
	preparer, supported := any(source).(fetchPreparer)
	require.True(t, supported, "root wrapper must advertise independent preparation")
	complete, err := preparer.PrepareFetch(t.Context(), FetchRequest{Source: "a", Version: "v1.0.0"})
	require.NoError(t, err)
	assert.Empty(t, source.fetched)
	_, err = complete(t.Context())
	require.ErrorIs(t, err, ErrLockedVersionChanged)
	assert.Zero(t, backend.selections)
	assert.Empty(t, source.fetched)
}

type preparedRootTestSource struct {
	preparedTestSource
	selections int
}

func (source *preparedRootTestSource) PreparationSource() Source { return source }

func (source *preparedRootTestSource) FetchForRootResolution(ctx context.Context, name, version string) (Fetched, error) {
	return source.complete(ctx, FetchRequest{Source: name, Version: version, Inspect: true})
}

func (source *preparedRootTestSource) FetchWithRoots(ctx context.Context, name, version string, roots []string) (Fetched, error) {
	source.selections++
	return source.complete(ctx, FetchRequest{Source: name, Version: version, Roots: roots, Inspect: true})
}

type preparedTestSource struct {
	started      chan<- FetchRequest
	release      <-chan struct{}
	prepare      func(FetchRequest) (PreparedFetch, error)
	preparations atomic.Int64
	fetches      atomic.Int64
	consumed     []string
}

func (source *preparedTestSource) PreparationSource() Source { return source }

func (source *preparedTestSource) Fetch(ctx context.Context, name, version string) (Fetched, error) {
	source.fetches.Add(1)
	return source.complete(ctx, FetchRequest{Source: name, Version: version})
}

func (source *preparedTestSource) PrepareFetch(_ context.Context, request FetchRequest) (PreparedFetch, error) {
	source.preparations.Add(1)
	if source.prepare != nil {
		return source.prepare(request)
	}
	if source.started != nil {
		source.started <- request
		<-source.release
	}
	return func(ctx context.Context) (Fetched, error) { return source.complete(ctx, request) }, nil
}

func (source *preparedTestSource) complete(_ context.Context, request FetchRequest) (Fetched, error) {
	source.consumed = append(source.consumed, request.Source+"@"+request.Version)
	fetched := Fetched{
		Module: v1.Module{Name: request.Source, Roots: []string{"."}},
		Lock:   v1.LockedModule{Source: request.Source, Version: request.Version, Commit: versionlessCommitA, Hash: "h1:" + strings.Repeat("A", 43) + "="},
	}
	if len(request.Roots) > 0 {
		fetched.Module.Roots = request.Roots
		fetched.Lock.Roots = request.Roots
	}
	if request.Inspect {
		fetched.Inspection = &RootInspection{}
	}
	return fetched, nil
}
