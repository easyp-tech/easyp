package plugin

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
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

	// Prepare plugin parameters
	options := lo.MapToSlice(plugin.Options, func(k string, v string) string {
		if v == "" {
			return k
		}
		return k + "=" + v
	})

	// Update parameter in request
	if len(options) > 0 {
		request.Parameter = proto.String(strings.Join(options, ","))
	}

	// Marshal request to protobuf
	reqData, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("proto.Marshal request: %w", err)
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
func (e *BuiltinPluginExecutor) runWasmPlugin(ctx context.Context, pluginName string, stdin io.Reader) ([]byte, error) {
	// Get WASM module and arguments for the plugin
	wasmBin, args, err := getWasmModule(pluginName)
	if err != nil {
		return nil, fmt.Errorf("get wasm module for plugin %s: %w", pluginName, err)
	}

	// Create context with allocator
	ctx = experimental.WithMemoryAllocator(ctx, allocator.NewNonMoving())

	// Create wazero runtime
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCoreFeatures(api.CoreFeaturesV2|experimental.CoreFeaturesThreads))

	// Close runtime at the end
	defer rt.Close(ctx)

	// Instantiate WASI
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	// Instantiate memory module
	if len(wasmMemory) > 0 {
		if _, err = rt.InstantiateWithConfig(ctx, wasmMemory, wazero.NewModuleConfig().WithName("env")); err != nil {
			return nil, fmt.Errorf("failed to instantiate memory module: %w", err)
		}
	}

	// Read stdin into buffer
	var stdinBuf bytes.Buffer
	if _, err = io.Copy(&stdinBuf, stdin); err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
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
	_, err = rt.InstantiateWithConfig(ctx, wasmBin, cfg)
	if err != nil {
		if sErr, ok := err.(*sys.ExitError); ok { //nolint:errorlint
			return nil, fmt.Errorf("wasm plugin exited with code %d", sErr.ExitCode())
		}
		return nil, fmt.Errorf("failed to instantiate wasm module: %w", err)
	}

	return stdoutBuf.Bytes(), nil
}
