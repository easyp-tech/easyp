package plugin

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/adapters/console"
)

type Info struct {
	Source  string
	Options map[string]string
}

// LocalPluginExecutor executes plugins locally via terminal
type LocalPluginExecutor struct {
	console console.Console
	logger  *slog.Logger
}

// NewLocalPluginExecutor creates a new LocalPluginExecutor
func NewLocalPluginExecutor(console console.Console, logger *slog.Logger) *LocalPluginExecutor {
	return &LocalPluginExecutor{
		console: console,
		logger:  logger,
	}
}

// Execute executes a local plugin via terminal
func (e *LocalPluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	e.logger.DebugContext(ctx, "executing local plugin",
		slog.String("plugin", plugin.Source),
	)

	options := lo.MapToSlice(plugin.Options, func(k string, v string) string {
		if v == "" {
			return k
		}
		return k + "=" + v
	})

	if len(options) > 0 {
		request.Parameter = proto.String(strings.Join(options, ","))
	}

	reqData, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("proto.Marshal request: %w", err)
	}

	stdIn := bytes.NewReader(reqData)

	command := e.determineCommand(plugin.Source)
	stdout, err := e.console.RunCmdWithStdin(ctx, ".", stdIn, command)
	if err != nil {
		return nil, fmt.Errorf("run local plugin %s: %w", plugin.Source, err)
	}

	// Парсим ответ от плагина
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal([]byte(stdout), &resp); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal response from plugin %s: %w", plugin.Source, err)
	}

	return &resp, nil
}

func (e *LocalPluginExecutor) determineCommand(source string) string {
	// Проверяем признаки того, что это путь к файлу
	if e.looksLikePath(source) {
		return source
	}

	// Это имя плагина - добавляем префикс protoc-gen-
	command := fmt.Sprintf("protoc-gen-%s", source)

	// На Windows добавляем .exe если его нет
	if runtime.GOOS == "windows" && !strings.HasSuffix(command, ".exe") {
		command += ".exe"
	}

	return command
}

func (e *LocalPluginExecutor) looksLikePath(source string) bool {
	// 1. Absolut path.
	if filepath.IsAbs(source) {
		return true
	}

	// 2. Path.
	if strings.HasPrefix(source, ".") {
		return true
	}

	// 5. На Windows: содержит расширение исполняемого файла
	if runtime.GOOS == "windows" {
		lowered := strings.ToLower(source)
		execExtensions := []string{".exe", ".bat", ".cmd", ".ps1"}
		for _, ext := range execExtensions {
			if strings.HasSuffix(lowered, ext) {
				return true
			}
		}
	}

	if runtime.GOOS != "windows" && strings.HasSuffix(source, ".sh") {
		return true
	}

	return false
}
