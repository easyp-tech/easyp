package plugin

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

func TestBuiltinRunCompilesOnceAndIsolatesInvocations(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		const count = 4
		var compilations atomic.Int64
		started := make(chan api.Memory, count)
		release := make(chan struct{})
		releaseAll := sync.OnceFunc(func() { close(release) })
		ctx := experimental.WithFunctionListenerFactory(t.Context(), experimental.FunctionListenerFactoryFunc(func(def api.FunctionDefinition) experimental.FunctionListener {
			if !slices.Contains(def.ExportNames(), "_start") {
				return nil
			}
			compilations.Add(1)
			return experimental.FunctionListenerFunc(func(_ context.Context, mod api.Module, _ api.FunctionDefinition, _ []uint64, _ experimental.StackIterator) {
				started <- mod.Memory()
				<-release
			})
		}))
		ctx, run := EnsureBuiltinRun(ctx)
		require.NotNil(t, run)
		defer func() {
			releaseAll()
			assert.NoError(t, run.Close(context.Background()))
		}()

		module := builtinEchoModule(t)
		type result struct {
			input  string
			output []byte
			err    error
		}
		results := make(chan result, count)
		for i := range count {
			input := strings.Repeat(string(rune('a'+i)), i+1)
			go func() {
				output, err := runBuiltinModule(ctx, module, []string{"echo"}, strings.NewReader(input))
				results <- result{input: input, output: output, err: err}
			}()
		}

		synctest.Wait()
		require.Len(t, started, count, "all invocations must overlap")
		assert.EqualValues(t, 1, compilations.Load())
		memories := make([]api.Memory, 0, count)
		for range count {
			memory := <-started
			require.NotNil(t, memory)
			for _, previous := range memories {
				assert.NotSame(t, previous, memory)
			}
			memories = append(memories, memory)
		}

		releaseAll()
		synctest.Wait()
		require.Len(t, results, count)
		for range count {
			result := <-results
			require.NoError(t, result.err)
			assert.Equal(t, result.input, string(result.output))
		}

		// Reuse after the previous runtimes closed must keep the compiled code alive.
		output, err := runBuiltinModule(ctx, module, []string{"echo"}, strings.NewReader("later"))
		require.NoError(t, err)
		assert.Equal(t, "later", string(output))
		assert.EqualValues(t, 1, compilations.Load())

		first, err := run.compiledModule(ctx, wasmMemory)
		require.NoError(t, err)
		second, err := run.compiledModule(ctx, bytes.Clone(wasmMemory))
		require.NoError(t, err)
		assert.Same(t, first, second, "equal binary contents share one compiled module")
	})
}

func TestEnsureBuiltinRunReusesScope(t *testing.T) {
	t.Parallel()

	ctx, run := EnsureBuiltinRun(t.Context())
	require.NotNil(t, run)
	nested, owner := EnsureBuiltinRun(ctx)
	assert.Same(t, ctx, nested)
	assert.Nil(t, owner)
	require.NoError(t, run.Close(context.Background()))
	require.NoError(t, run.Close(context.Background()))

	output, err := runBuiltinModule(ctx, builtinEchoModule(t), []string{"echo"}, strings.NewReader(""))
	require.ErrorContains(t, err, "closed")
	assert.Nil(t, output)
}

func TestBuiltinRunCloseJoinsInvocations(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		started := make(chan struct{}, 1)
		release := make(chan struct{})
		releaseAll := sync.OnceFunc(func() { close(release) })
		ctx := experimental.WithFunctionListenerFactory(t.Context(), experimental.FunctionListenerFactoryFunc(func(def api.FunctionDefinition) experimental.FunctionListener {
			if !slices.Contains(def.ExportNames(), "_start") {
				return nil
			}
			return experimental.FunctionListenerFunc(func(context.Context, api.Module, api.FunctionDefinition, []uint64, experimental.StackIterator) {
				started <- struct{}{}
				<-release
			})
		}))
		ctx, run := EnsureBuiltinRun(ctx)
		defer func() {
			releaseAll()
			assert.NoError(t, run.Close(context.Background()))
		}()
		results := make(chan error, 1)
		module := builtinEchoModule(t)
		go func() {
			_, err := runBuiltinModule(ctx, module, []string{"echo"}, strings.NewReader("first"))
			results <- err
		}()
		synctest.Wait()
		require.Len(t, started, 1)

		closed := make(chan error, 1)
		go func() { closed <- run.Close(context.Background()) }()
		synctest.Wait()
		assert.Empty(t, closed, "Close must join the invocation")
		_, err := runBuiltinModule(ctx, module, []string{"echo"}, strings.NewReader("late"))
		require.ErrorContains(t, err, "closed")

		releaseAll()
		synctest.Wait()
		require.Len(t, closed, 1)
		require.NoError(t, <-closed)
		require.Len(t, results, 1)
		require.ErrorIs(t, <-results, context.Canceled)
	})
}

func TestBuiltinRunCanceledWaiterDoesNotCancelCompilation(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		var compilations atomic.Int64
		compiling := make(chan struct{}, 1)
		release := make(chan struct{})
		releaseAll := sync.OnceFunc(func() { close(release) })
		ctx := experimental.WithFunctionListenerFactory(t.Context(), experimental.FunctionListenerFactoryFunc(func(def api.FunctionDefinition) experimental.FunctionListener {
			if slices.Contains(def.ExportNames(), "_start") {
				compilations.Add(1)
				compiling <- struct{}{}
				<-release
			}
			return nil
		}))
		ctx, run := EnsureBuiltinRun(ctx)
		defer func() {
			releaseAll()
			assert.NoError(t, run.Close(context.Background()))
		}()
		module := builtinEchoModule(t)
		caller, cancel := context.WithCancel(ctx)
		defer cancel()
		first := make(chan error, 1)
		go func() {
			_, err := runBuiltinModule(caller, module, []string{"echo"}, strings.NewReader("canceled"))
			first <- err
		}()
		synctest.Wait()
		require.Len(t, compiling, 1)
		cancel()
		synctest.Wait()
		require.Len(t, first, 1, "the initiating caller must stop waiting promptly")
		require.ErrorIs(t, <-first, context.Canceled)

		second := make(chan error, 1)
		go func() {
			output, err := runBuiltinModule(ctx, module, []string{"echo"}, strings.NewReader("survived"))
			if err == nil && string(output) != "survived" {
				err = errors.New("unexpected output: " + string(output))
			}
			second <- err
		}()
		synctest.Wait()
		assert.Empty(t, second)
		releaseAll()
		synctest.Wait()
		require.Len(t, second, 1)
		require.NoError(t, <-second)
		assert.EqualValues(t, 1, compilations.Load())
	})
}

func TestBuiltinRunCloseJoinsCompilation(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		compiling := make(chan struct{}, 1)
		release := make(chan struct{})
		releaseAll := sync.OnceFunc(func() { close(release) })
		ctx := experimental.WithFunctionListenerFactory(t.Context(), experimental.FunctionListenerFactoryFunc(func(api.FunctionDefinition) experimental.FunctionListener {
			compiling <- struct{}{}
			<-release
			return nil
		}))
		ctx, run := EnsureBuiltinRun(ctx)
		defer func() {
			releaseAll()
			assert.NoError(t, run.Close(context.Background()))
		}()
		module := builtinEchoModule(t)
		result := make(chan error, 1)
		go func() {
			_, err := run.compiledModule(ctx, module)
			result <- err
		}()
		synctest.Wait()
		require.Len(t, compiling, 1)

		closed := make(chan error, 1)
		go func() { closed <- run.Close(context.Background()) }()
		synctest.Wait()
		assert.Empty(t, closed, "Close must wait for compilation to finish")
		require.Len(t, result, 1)
		require.ErrorIs(t, <-result, context.Canceled)

		releaseAll()
		synctest.Wait()
		require.Len(t, closed, 1)
		require.NoError(t, <-closed)
	})
}

func TestBuiltinRunCompilesEachDistinctBinary(t *testing.T) {
	t.Parallel()

	var compilations atomic.Int64
	ctx := experimental.WithFunctionListenerFactory(t.Context(), experimental.FunctionListenerFactoryFunc(func(api.FunctionDefinition) experimental.FunctionListener {
		compilations.Add(1)
		return nil
	}))
	ctx, run := EnsureBuiltinRun(ctx)
	defer func() { assert.NoError(t, run.Close(context.Background())) }()
	module := builtinEchoModule(t)
	first, err := run.compiledModule(ctx, module)
	require.NoError(t, err)
	copy, err := run.compiledModule(ctx, bytes.Clone(module))
	require.NoError(t, err)
	assert.Same(t, first, copy)

	// Append a custom section to create different valid binary contents.
	different := append(bytes.Clone(module), 0, 3, 1, 'x', 0)
	second, err := run.compiledModule(ctx, different)
	require.NoError(t, err)
	assert.NotSame(t, first, second)
	assert.EqualValues(t, 2, compilations.Load())
}

func TestBuiltinRunFailedCompilationIsMemoizedAndCleanupErrorsSurvive(t *testing.T) {
	t.Parallel()

	ctx, run := EnsureBuiltinRun(t.Context())
	runtimeCloseErr := errors.New("runtime close failed")
	cacheCloseErr := errors.New("cache close failed")
	var compilations, runtimeCloses, cacheCloses atomic.Int64
	closeContexts := make(chan error, 2)
	run.runtimeFactory = func(ctx context.Context, config wazero.RuntimeConfig) wazero.Runtime {
		return &builtinRuntimeProbe{
			Runtime: wazero.NewRuntimeWithConfig(ctx, config), compilations: &compilations,
			closes: &runtimeCloses, closeContexts: closeContexts, closeErr: runtimeCloseErr,
		}
	}
	run.cache = &builtinCacheProbe{CompilationCache: run.cache, closes: &cacheCloses, closeErr: cacheCloseErr}

	_, firstErr := run.compiledModule(ctx, []byte("invalid WASM"))
	require.Error(t, firstErr)
	_, secondErr := run.compiledModule(ctx, []byte("invalid WASM"))
	require.EqualError(t, secondErr, firstErr.Error())
	assert.EqualValues(t, 1, compilations.Load())

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	err := run.Close(canceled)
	require.ErrorIs(t, err, runtimeCloseErr)
	require.ErrorIs(t, err, cacheCloseErr)
	assert.NoError(t, <-closeContexts, "cleanup must receive an uncanceled context")
	assert.EqualValues(t, 1, runtimeCloses.Load())
	assert.EqualValues(t, 1, cacheCloses.Load())
	assert.Equal(t, err, run.Close(canceled))
	assert.EqualValues(t, 1, runtimeCloses.Load())
	assert.EqualValues(t, 1, cacheCloses.Load())
}

func TestBuiltinRunInvocationRetainsCleanupErrors(t *testing.T) {
	t.Parallel()

	readErr := errors.New("stdin failed")
	tests := []struct {
		name   string
		reader func() io.Reader
		want   error
	}{
		{name: "successful_execution", reader: func() io.Reader { return strings.NewReader("hello") }},
		{name: "failed_execution", reader: func() io.Reader { return builtinErrorReader{err: readErr} }, want: readErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, run := EnsureBuiltinRun(t.Context())
			closeErr := errors.New("runtime close failed")
			var compilations, closes atomic.Int64
			closeContexts := make(chan error, 2)
			run.runtimeFactory = func(ctx context.Context, config wazero.RuntimeConfig) wazero.Runtime {
				return &builtinRuntimeProbe{
					Runtime: wazero.NewRuntimeWithConfig(ctx, config), compilations: &compilations,
					closes: &closes, closeContexts: closeContexts, closeErr: closeErr,
				}
			}

			output, err := runBuiltinModule(ctx, builtinEchoModule(t), []string{"echo"}, tt.reader())

			require.ErrorIs(t, err, closeErr)
			assert.Nil(t, output)
			if tt.want != nil {
				assert.ErrorIs(t, err, tt.want)
			}
			assert.NoError(t, <-closeContexts)
			require.ErrorIs(t, run.Close(context.Background()), closeErr)
			assert.NoError(t, <-closeContexts)
			assert.EqualValues(t, 2, closes.Load())
		})
	}
}

func TestBuiltinRunCanceledCommandDoesNotCompile(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	ctx, run := EnsureBuiltinRun(ctx)
	cancel()
	var runtimes atomic.Int64
	run.runtimeFactory = func(ctx context.Context, config wazero.RuntimeConfig) wazero.Runtime {
		runtimes.Add(1)
		return wazero.NewRuntimeWithConfig(ctx, config)
	}

	output, err := runBuiltinModule(ctx, builtinEchoModule(t), []string{"echo"}, strings.NewReader(""))

	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, output)
	assert.Zero(t, runtimes.Load())
	require.NoError(t, run.Close(ctx))
}

func TestBuiltinRunCancellationTerminatesWasm(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		started := make(chan struct{}, 1)
		ctx := experimental.WithFunctionListenerFactory(t.Context(), experimental.FunctionListenerFactoryFunc(func(def api.FunctionDefinition) experimental.FunctionListener {
			if !slices.Contains(def.ExportNames(), "_start") {
				return nil
			}
			return experimental.FunctionListenerFunc(func(context.Context, api.Module, api.FunctionDefinition, []uint64, experimental.StackIterator) {
				started <- struct{}{}
			})
		}))
		ctx, run := EnsureBuiltinRun(ctx)
		defer func() { assert.NoError(t, run.Close(context.Background())) }()
		caller, cancel := context.WithCancel(ctx)
		defer cancel()
		module := builtinWasmHex(t, "0061736d0100000001040160000003020100070a01065f737461727400000a0901070003400c000b0b")
		results := make(chan error, 1)
		go func() {
			_, err := runBuiltinModule(caller, module, []string{"loop"}, strings.NewReader(""))
			results <- err
		}()
		<-started
		cancel()
		require.ErrorIs(t, <-results, context.Canceled)
	})
}

func builtinEchoModule(t *testing.T) []byte {
	t.Helper()
	// This WASI fixture imports env.memory and echoes stdin. Both a mutable
	// global and memory byte 100 trap when reused by a later invocation.
	return builtinWasmHex(t, "0061736d01000000010c0260047f7f7f7f017f60000002550316776173695f736e617073686f745f70726576696577310766645f72656164000016776173695f736e617073686f745f70726576696577310866645f7772697465000003656e76066d656d6f727902010d808004030201010606017f0141000b071302066d656d6f72790200065f737461727400020a50014e0023000440000b4101240041e4002d00000440000b41e40041013a000041004180083602004104418002360200410041004101410810001a41044108280200360200410141004101410c10011a0b")
}

func builtinWasmHex(t *testing.T, encoded string) []byte {
	t.Helper()
	binary, err := hex.DecodeString(encoded)
	require.NoError(t, err)
	return binary
}

type builtinRuntimeProbe struct {
	wazero.Runtime
	compilations  *atomic.Int64
	closes        *atomic.Int64
	closeContexts chan<- error
	closeErr      error
}

func (r *builtinRuntimeProbe) CompileModule(ctx context.Context, binary []byte) (wazero.CompiledModule, error) {
	r.compilations.Add(1)
	return r.Runtime.CompileModule(ctx, binary)
}

func (r *builtinRuntimeProbe) Close(ctx context.Context) error {
	r.closes.Add(1)
	r.closeContexts <- ctx.Err()
	return errors.Join(r.Runtime.Close(ctx), r.closeErr)
}

type builtinCacheProbe struct {
	wazero.CompilationCache
	closes   *atomic.Int64
	closeErr error
}

func (c *builtinCacheProbe) Close(ctx context.Context) error {
	c.closes.Add(1)
	return errors.Join(c.CompilationCache.Close(ctx), c.closeErr)
}

type builtinErrorReader struct{ err error }

func (r builtinErrorReader) Read([]byte) (int, error) { return 0, r.err }
