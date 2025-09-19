package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	plugingeneratorv1 "github.com/easyp-tech/easyp-plugin-server/api/plugin-generator/v1"
	"github.com/samber/lo"
)

// RemotePluginExecutor executes plugins remotely via gRPC
type RemotePluginExecutor struct {
	logger *slog.Logger
}

// NewRemotePluginExecutor creates a new RemotePluginExecutor
func NewRemotePluginExecutor(logger *slog.Logger) *RemotePluginExecutor {
	return &RemotePluginExecutor{
		logger: logger,
	}
}

// Execute executes a remote plugin via gRPC
func (e *RemotePluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	// Парсим URL плагина для извлечения хоста и информации о плагине
	host, pluginName, version, err := e.parsePluginURL(plugin.URL)
	if err != nil {
		return nil, fmt.Errorf("parse plugin URL %s: %w", plugin.URL, err)
	}

	e.logger.DebugContext(ctx, "executing remote plugin via gRPC",
		slog.String("plugin", plugin.Name),
		slog.String("host", host),
		slog.String("plugin_name", pluginName),
		slog.String("version", version),
		slog.String("url", plugin.URL),
	)

	// Создаем контекст с таймаутом
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Устанавливаем gRPC соединение
	conn, err := grpc.NewClient(host,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server %s: %w", host, err)
	}
	defer conn.Close()

	// Создаем gRPC клиент
	client := plugingeneratorv1.NewPluginGeneratorServiceClient(conn)

	// Формируем информацию о плагине в формате "name:version"
	pluginInfo := fmt.Sprintf("%s:%s", pluginName, version)

	// Подготавливаем параметры плагина
	options := lo.MapToSlice(plugin.Options, func(k string, v string) string {
		if v == "" {
			return k
		}
		return k + "=" + v
	})

	if len(options) > 0 {
		request.Parameter = proto.String(strings.Join(options, ","))
	}

	// Создаем запрос для gRPC сервиса
	grpcRequest := &plugingeneratorv1.GenerateCodeRequest{
		CodeGeneratorRequest: request,
		PluginInfo:           pluginInfo,
	}

	// Вызываем удаленный плагин
	resp, err := client.GenerateCode(ctxWithTimeout, grpcRequest)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed for plugin %s: %w", plugin.Name, err)
	}

	// Проверяем статус ответа
	if resp.Status != "success" && resp.Status != "" {
		return nil, fmt.Errorf("remote plugin returned error status '%s': %s", resp.Status, resp.Message)
	}

	return resp.CodeGeneratorResponse, nil
}

// parsePluginURL парсит URL плагина в формате:
// - localhost:8080/python:v1.35
// - http://localhost:8080/python:v1.35
// - https://example.com/python:v1.35
func (e *RemotePluginExecutor) parsePluginURL(pluginURL string) (host, pluginName, version string, err error) {
	// Нормализуем URL, добавляем схему если её нет
	normalizedURL := pluginURL
	if !strings.HasPrefix(pluginURL, "http://") && !strings.HasPrefix(pluginURL, "https://") {
		normalizedURL = "http://" + pluginURL
	}

	// Парсим URL
	parsedURL, err := url.Parse(normalizedURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid URL: %w", err)
	}

	// Извлекаем хост (host:port)
	host = parsedURL.Host

	// Извлекаем путь и парсим plugin_name:version
	path := strings.TrimPrefix(parsedURL.Path, "/")
	if path == "" {
		return "", "", "", fmt.Errorf("plugin name not specified in URL")
	}

	// Разделяем plugin_name:version
	parts := strings.SplitN(path, ":", 2)
	pluginName = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	} else {
		version = "latest" // версия по умолчанию
	}

	if pluginName == "" {
		return "", "", "", fmt.Errorf("plugin name is empty")
	}

	return host, pluginName, version, nil
}
