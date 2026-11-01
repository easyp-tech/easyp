package plugin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

type builtinRunContextKey struct{}

// BuiltinRun owns compiled builtin modules for one command. Each invocation
// uses its own runtime, module instances, memory, and I/O.
type BuiltinRun struct {
	ctx            context.Context
	cancel         context.CancelFunc
	cache          wazero.CompilationCache
	config         wazero.RuntimeConfig
	runtimeFactory func(context.Context, wazero.RuntimeConfig) wazero.Runtime

	mu       sync.Mutex
	closed   bool
	compiler wazero.Runtime
	modules  map[[sha256.Size]byte]*builtinCompilation
	work     sync.WaitGroup
	close    sync.Once
	closeErr error
}

type builtinCompilation struct {
	ready  chan struct{}
	module wazero.CompiledModule
	err    error
}

// EnsureBuiltinRun reuses the command's builtin compilation scope, or creates
// one when absent. A non-nil owner must be closed after all callers finish.
// Nested callers receive a nil owner and must leave the existing run open.
func EnsureBuiltinRun(ctx context.Context) (context.Context, *BuiltinRun) {
	if _, ok := ctx.Value(builtinRunContextKey{}).(*BuiltinRun); ok {
		return ctx, nil
	}

	commandCtx, cancel := context.WithCancel(ctx)
	cache := wazero.NewCompilationCache()
	run := &BuiltinRun{
		ctx: commandCtx, cancel: cancel, cache: cache,
		config: wazero.NewRuntimeConfig().
			WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesThreads).
			WithCloseOnContextDone(true).
			WithCompilationCache(cache),
		runtimeFactory: wazero.NewRuntimeWithConfig,
		modules:        make(map[[sha256.Size]byte]*builtinCompilation),
	}
	return context.WithValue(ctx, builtinRunContextKey{}, run), run
}

// Close stops new invocations, cancels active work, joins preparation and
// invocations, then releases compiled modules and their shared engine.
// Cleanup continues even when ctx is canceled. Repeated calls return the same
// cleanup error. Callers should normally join their invocations before Close.
func (r *BuiltinRun) Close(ctx context.Context) error {
	r.close.Do(func() {
		r.mu.Lock()
		r.closed = true
		r.cancel()
		r.mu.Unlock()
		r.work.Wait()

		cleanupCtx := context.WithoutCancel(ctx)
		keys := make([][sha256.Size]byte, 0, len(r.modules))
		for key := range r.modules {
			keys = append(keys, key)
		}
		slices.SortFunc(keys, func(left, right [sha256.Size]byte) int {
			return bytes.Compare(left[:], right[:])
		})
		for _, key := range keys {
			module := r.modules[key].module
			if module == nil {
				continue
			}
			err := module.Close(cleanupCtx)
			if err != nil {
				r.closeErr = errors.Join(r.closeErr, fmt.Errorf("Close: %w", err))
			}
		}
		if r.compiler != nil {
			err := r.compiler.Close(cleanupCtx)
			if err != nil {
				r.closeErr = errors.Join(r.closeErr, fmt.Errorf("Close: %w", err))
			}
		}
		err := r.cache.Close(cleanupCtx)
		if err != nil {
			r.closeErr = errors.Join(r.closeErr, fmt.Errorf("Close: %w", err))
		}
	})
	return r.closeErr
}

func (r *BuiltinRun) acquire(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return fmt.Errorf("builtin run is closed")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := r.ctx.Err(); err != nil {
		return err
	}
	r.work.Add(1)
	return nil
}

func (r *BuiltinRun) compiledModule(ctx context.Context, binary []byte) (wazero.CompiledModule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key := sha256.Sum256(binary)
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil, fmt.Errorf("builtin run is closed")
	}
	if err := r.ctx.Err(); err != nil {
		r.mu.Unlock()
		return nil, err
	}
	compilation, ok := r.modules[key]
	if !ok {
		compilation = &builtinCompilation{ready: make(chan struct{})}
		r.modules[key] = compilation
		if r.compiler == nil {
			r.compiler = r.runtimeFactory(r.ctx, r.config)
		}
		r.work.Add(1)
		go r.compile(binary, compilation)
	}
	r.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.ctx.Done():
		return nil, r.ctx.Err()
	case <-compilation.ready:
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := r.ctx.Err(); err != nil {
			return nil, err
		}
		return compilation.module, compilation.err
	}
}

func (r *BuiltinRun) compile(binary []byte, compilation *builtinCompilation) {
	defer r.work.Done()
	defer close(compilation.ready)

	// Preparation belongs to the command. Canceling the first waiting target
	// must not poison the compiled module for the other targets.
	if err := r.ctx.Err(); err != nil {
		compilation.err = err
		return
	}
	module, err := r.compiler.CompileModule(r.ctx, binary)
	compilation.module = module
	if err != nil {
		compilation.err = fmt.Errorf("CompileModule: %w", err)
	}
}
