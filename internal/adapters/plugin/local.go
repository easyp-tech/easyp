package plugin

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/adapters/console"
)

type Info struct {
	URL     string
	Name    string
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
		slog.String("plugin", plugin.Name),
	)

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

	// Вызываем плагин через терминал
	stdout, err := e.console.RunCmdWithStdin(ctx, ".", stdIn, fmt.Sprintf("protoc-gen-%s", plugin.Name))
	if err != nil {
		return nil, fmt.Errorf("run local plugin %s: %w", plugin.Name, err)
	}

	// Парсим ответ от плагина
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal([]byte(stdout), &resp); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal response from plugin %s: %w", plugin.Name, err)
	}

	return &resp, nil
}
