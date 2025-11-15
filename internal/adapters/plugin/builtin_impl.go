//go:build builtin_plugins
// +build builtin_plugins

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
)

func builtinPluginExecutorEnabled() bool {
	return true
}

// Маппинг имен плагинов на WASM-модули (через go:linkname)
var builtinPluginMap = map[builtinPlugin][]byte{
	builtinPluginCpp:          protocGenCpp,
	builtinPluginCsharp:       protocGenCsharp,
	builtinPluginJava:         protocGenJava,
	builtinPluginKotlin:       protocGenKotlin,
	builtinPluginObjc:         protocGenObjc,
	builtinPluginPhp:          protocGenPhp,
	builtinPluginPyi:          protocGenPyi,
	builtinPluginPython:       protocGenPython,
	builtinPluginRuby:         protocGenRuby,
	builtinPluginRust:         protocGenRust,
	builtinPluginUpb:          protocGenUPB,
	builtinPluginUpbMinitable: protocGenUPBMinitable,
	builtinPluginUpbDefs:      protocGenUPBDefs,
}

// getWasmModule получает WASM-модуль по имени плагина
func getWasmModule(pluginName string) ([]byte, error) {
	if pluginName == "memory" {
		return wasmMemory, nil
	}

	wasmData, ok := builtinPluginMap[builtinPlugin(pluginName)]
	if !ok {
		return nil, fmt.Errorf("plugin %s is not supported", pluginName)
	}

	return wasmData, nil
}

// BuiltinPluginExecutor executes builtin plugins via WASM
type BuiltinPluginExecutor struct {
	logger *slog.Logger
}

// NewBuiltinPluginExecutor creates a new BuiltinPluginExecutor
func NewBuiltinPluginExecutor(logger *slog.Logger) *BuiltinPluginExecutor {
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
	e.logger.DebugContext(ctx, "executing builtin plugin",
		slog.String("plugin", plugin.Name),
	)

	// Получаем WASM-модуль для плагина
	wasmBin, err := getWasmModule(plugin.Name)
	if err != nil {
		return nil, fmt.Errorf("get wasm module for plugin %s: %w", plugin.Name, err)
	}

	// Подготавливаем параметры плагина
	options := lo.MapToSlice(plugin.Options, func(k string, v string) string {
		if v == "" {
			return k
		}
		return k + "=" + v
	})

	// Обновляем параметр в запросе
	if len(options) > 0 {
		request.Parameter = proto.String(strings.Join(options, ","))
	}

	// Маршалим запрос в protobuf
	reqData, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("proto.Marshal request: %w", err)
	}

	// Создаем буфер для stdin
	stdIn := bytes.NewReader(reqData)

	// Запускаем WASM-плагин
	stdout, err := e.runWasmPlugin(ctx, plugin.Name, wasmBin, stdIn)
	if err != nil {
		return nil, fmt.Errorf("run wasm plugin %s: %w", plugin.Name, err)
	}

	// Парсим ответ от плагина
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal response from plugin %s: %w", plugin.Name, err)
	}

	return &resp, nil
}

// runWasmPlugin запускает WASM-модуль с кастомными stdin/stdout
func (e *BuiltinPluginExecutor) runWasmPlugin(ctx context.Context, pluginName string, wasmBin []byte, stdin io.Reader) ([]byte, error) {
	// Создаем контекст с allocator
	ctx = experimental.WithMemoryAllocator(ctx, allocator.NewNonMoving())

	// Создаем wazero runtime
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCoreFeatures(api.CoreFeaturesV2|experimental.CoreFeaturesThreads))

	// Закрываем runtime в конце
	defer rt.Close(ctx)

	// Инстанцируем WASI
	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	// Инстанцируем memory модуль
	if len(wasmMemory) > 0 {
		if _, err := rt.InstantiateWithConfig(ctx, wasmMemory, wazero.NewModuleConfig().WithName("env")); err != nil {
			return nil, fmt.Errorf("failed to instantiate memory module: %w", err)
		}
	}

	// Читаем stdin в буфер
	var stdinBuf bytes.Buffer
	if _, err := io.Copy(&stdinBuf, stdin); err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	// Создаем буфер для stdout
	var stdoutBuf bytes.Buffer

	// Настраиваем конфигурацию модуля
	args := []string{fmt.Sprintf("protoc-gen-%s", pluginName)}
	cfg := wazero.NewModuleConfig().
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithStderr(&bytes.Buffer{}). // stderr в отдельный буфер
		WithStdout(&stdoutBuf).
		WithStdin(&stdinBuf).
		WithRandSource(rand.Reader).
		WithArgs(args...)

	// Инстанцируем WASM-модуль
	_, err := rt.InstantiateWithConfig(ctx, wasmBin, cfg)
	if err != nil {
		if sErr, ok := err.(*sys.ExitError); ok { //nolint:errorlint
			return nil, fmt.Errorf("wasm plugin exited with code %d", sErr.ExitCode())
		}
		return nil, fmt.Errorf("failed to instantiate wasm module: %w", err)
	}

	return stdoutBuf.Bytes(), nil
}
