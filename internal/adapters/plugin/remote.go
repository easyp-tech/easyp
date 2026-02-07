package plugin

import (
	"context"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	plugingeneratorv1 "github.com/easyp-tech/service/api/generator/v1"
	"github.com/samber/lo"

	"github.com/easyp-tech/easyp/internal/logger"
)

// RemotePluginExecutor executes plugins remotely via gRPC
type RemotePluginExecutor struct {
	logger logger.Logger
}

func (e *RemotePluginExecutor) GetName() string {
	return "RemotePluginExecutor from URL"
}

// NewRemotePluginExecutor creates a new RemotePluginExecutor
func NewRemotePluginExecutor(logger logger.Logger) *RemotePluginExecutor {
	return &RemotePluginExecutor{
		logger: logger,
	}
}

// Execute executes a remote plugin via gRPC
func (e *RemotePluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	// Парсим URL плагина для извлечения хоста и информации о плагине
	host, pluginName, version, err := e.parsePluginURL(plugin.Source)
	if err != nil {
		return nil, fmt.Errorf("parse plugin URL %s: %w", plugin.Source, err)
	}

	e.logger.Debug(ctx, "executing remote plugin via gRPC",
		slog.String("plugin", plugin.Source),
		slog.String("host", host),
		slog.String("plugin_name", pluginName),
		slog.String("version", version),
	)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var transportCreds credentials.TransportCredentials
	if strings.HasPrefix(plugin.Source, "https://") || (!strings.HasPrefix(plugin.Source, "http://") && !strings.Contains(host, "localhost") && !strings.Contains(host, "127.0.0.1")) {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("failed to get system cert pool: %w", err)
		}

		transportCreds = credentials.NewClientTLSFromCert(pool, "")
	} else {
		transportCreds = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(host, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server %s: %w", host, err)
	}
	defer conn.Close()

	// Создаем gRPC клиент
	client := plugingeneratorv1.NewServiceAPIClient(conn)

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

	grpcRequest := &plugingeneratorv1.GenerateCodeRequest{
		CodeGeneratorRequest: request,
		PluginName:           pluginInfo,
	}

	resp, err := client.GenerateCode(ctxWithTimeout, grpcRequest)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed for plugin %s: %w", plugin.Source, err)
	}

	err = conn.Close()
	if err != nil {
		e.logger.Warn(ctx, "failed to close gRPC connection",
			slog.String("plugin", plugin.Source),
			slog.String("error", err.Error()),
		)
	}

	return resp.CodeGeneratorResponse, nil
}

// - localhost:8080/python:v1.35
// - http://localhost:8080/python:v1.35
// - https://example.com/python:v1.35
func (e *RemotePluginExecutor) parsePluginURL(pluginURL string) (host, pluginName, version string, err error) {
	normalizedURL := pluginURL
	if !strings.HasPrefix(pluginURL, "http://") && !strings.HasPrefix(pluginURL, "https://") {
		normalizedURL = "http://" + pluginURL
	}

	parsedURL, err := url.Parse(normalizedURL)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid URL: %w", err)
	}

	host = parsedURL.Host

	path := strings.TrimPrefix(parsedURL.Path, "/")
	if path == "" {
		return "", "", "", fmt.Errorf("plugin name not specified in URL")
	}

	parts := strings.SplitN(path, ":", 2)
	pluginName = parts[0]
	if len(parts) > 1 {
		version = parts[1]
	} else {
		version = "latest"
	}

	if pluginName == "" {
		return "", "", "", fmt.Errorf("plugin name is empty")
	}

	return host, pluginName, version, nil
}
