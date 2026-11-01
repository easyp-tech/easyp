package plugin

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/wasilibs/wazero-helpers/allocator"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/logger"
)

// getWasmModule gets the WASM module and arguments by plugin name
// All plugins (protobuf and gRPC) use protoc-gen-universal.wasm
// Base protobuf plugins use plugin name as argument (e.g., "cpp")
// gRPC plugins use plugin name with grpc_ prefix as argument (e.g., "grpc_cpp")
func getWasmModule(pluginName string) ([]byte, []string, error) {
	if pluginName == "memory" {
		return wasmMemory, nil, nil
	}

	// Check if plugin is supported
	if !IsBuiltinPlugin(pluginName) {
		return nil, nil, fmt.Errorf("plugin %s is not supported", pluginName)
	}

	// All builtin plugins use protoc-gen-universal.wasm with plugin name as argument
	args := []string{"protoc-gen-universal", pluginName}
	return protocGenUniversal, args, nil
}

// BuiltinPluginExecutor executes builtin plugins via WASM
type BuiltinPluginExecutor struct {
	logger logger.Logger
}

// NewBuiltinPluginExecutor creates a new BuiltinPluginExecutor
func NewBuiltinPluginExecutor(logger logger.Logger) *BuiltinPluginExecutor {
	return &BuiltinPluginExecutor{
		logger: logger,
	}
}

// GetName returns the name of the executor
func (e *BuiltinPluginExecutor) GetName() string {
	return "BuiltinPluginExecutor (WASM)"
}

// Execute executes a builtin plugin via WASM
func (e *BuiltinPluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	e.logger.Debug(ctx, "executing builtin plugin",
		slog.String("plugin", plugin.Source),
	)

	// Update parameter in request
	if parameter, ok := flattenOptions(plugin.Options); ok {
		request.Parameter = proto.String(parameter)
	}

	// Marshal request to protobuf
	reqData, err := marshalRequest(request)
	if err != nil {
		return nil, fmt.Errorf("marshalRequest: %w", err)
	}

	// Create buffer for stdin
	stdIn := bytes.NewReader(reqData)

	// Run WASM plugin
	stdout, err := e.runWasmPlugin(ctx, plugin.Source, stdIn)
	if err != nil {
		return nil, fmt.Errorf("run wasm plugin %s: %w", plugin.Source, err)
	}

	// Parse response from plugin
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal response from plugin %s: %w", plugin.Source, err)
	}

	return &resp, nil
}

// runWasmPlugin runs WASM module with custom stdin/stdout
func (e *BuiltinPluginExecutor) runWasmPlugin(ctx context.Context, pluginName string, stdin io.Reader) (output []byte, runErr error) {
	wasmBin, args, err := getWasmModule(pluginName)
	if err != nil {
		return nil, fmt.Errorf("getWasmModule: %w", err)
	}
	return runBuiltinModule(ctx, wasmBin, args, stdin)
}

func runBuiltinModule(ctx context.Context, binary []byte, args []string, stdin io.Reader) (output []byte, runErr error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ctx, owner := EnsureBuiltinRun(ctx)
	if owner != nil {
		defer func() {
			err := owner.Close(context.WithoutCancel(ctx))
			if err != nil {
				output = nil
				runErr = errors.Join(runErr, fmt.Errorf("Close: %w", err))
			}
		}()
	}
	run := ctx.Value(builtinRunContextKey{}).(*BuiltinRun)
	err := run.acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire: %w", err)
	}
	defer run.work.Done()

	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(run.ctx, cancel)
	defer func() {
		stop()
		cancel()
	}()

	var memory wazero.CompiledModule
	if len(wasmMemory) > 0 {
		memory, err = run.compiledModule(ctx, wasmMemory)
		if err != nil {
			return nil, fmt.Errorf("compiledModule: %w", err)
		}
	}
	module, err := run.compiledModule(ctx, binary)
	if err != nil {
		return nil, fmt.Errorf("compiledModule: %w", err)
	}

	ctx = experimental.WithMemoryAllocator(ctx, allocator.NewNonMoving())
	rt := run.runtimeFactory(ctx, run.config)

	// A failed runtime close must not turn into successful generated output.
	defer func() {
		err := rt.Close(context.WithoutCancel(ctx))
		if err != nil {
			output = nil
			runErr = errors.Join(runErr, fmt.Errorf("Close: %w", err))
		}
	}()

	_, err = wasi_snapshot_preview1.Instantiate(ctx, rt)
	if err != nil {
		return nil, fmt.Errorf("Instantiate: %w", err)
	}

	if memory != nil {
		_, err = rt.InstantiateModule(ctx, memory, wazero.NewModuleConfig().WithName("env"))
		if err != nil {
			return nil, fmt.Errorf("InstantiateModule: %w", err)
		}
	}

	// Read stdin into buffer
	var stdinBuf bytes.Buffer
	_, err = io.Copy(&stdinBuf, stdin)
	if err != nil {
		return nil, fmt.Errorf("Copy: %w", err)
	}

	// Create buffer for stdout
	var stdoutBuf bytes.Buffer

	// Configure module
	cfg := wazero.NewModuleConfig().
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStderr(&bytes.Buffer{}). // stderr to separate buffer
		WithStdout(&stdoutBuf).
		WithStdin(&stdinBuf).
		WithRandSource(rand.Reader).
		WithArgs(args...)

	// Instantiate WASM module
	_, err = rt.InstantiateModule(ctx, module, cfg)
	if err != nil {
		if ctx.Err() != nil {
			err = errors.Join(err, ctx.Err())
		}
		return nil, fmt.Errorf("InstantiateModule: %w", err)
	}

	return stdoutBuf.Bytes(), nil
}
