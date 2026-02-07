package plugin

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/logger"
)

type Info struct {
	Source  string
	Command []string
	Options map[string]string
}

// LocalPluginExecutor executes plugins locally via terminal
type LocalPluginExecutor struct {
	console console.Console
	logger  logger.Logger
}

func (e *LocalPluginExecutor) GetName() string {
	return "LocalPluginExecutor from PATH"
}

// NewLocalPluginExecutor creates a new LocalPluginExecutor
func NewLocalPluginExecutor(console console.Console, logger logger.Logger) *LocalPluginExecutor {
	return &LocalPluginExecutor{
		console: console,
		logger:  logger,
	}
}

// isPluginInPath checks if the plugin is available in PATH
func (e *LocalPluginExecutor) isPluginInPath(source string) (string, bool) {
	command, err := exec.LookPath(source)
	return command, err == nil
}

// Execute executes a local plugin via terminal
func (e *LocalPluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	e.logger.Debug(ctx, "executing local plugin",
		slog.String("plugin", plugin.Source),
	)

	// Prepare plugin parameters
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

	command, err := e.determineCommand(plugin.Source)
	if err != nil {
		return nil, fmt.Errorf("determineCommand: %w", err)
	}

	stdout, err := e.console.RunCmdWithStdin(ctx, ".", stdIn, command)
	if err != nil {
		return nil, fmt.Errorf("run local plugin %s: %w", plugin.Source, err)
	}

	// Parse response from plugin
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal([]byte(stdout), &resp); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal response from plugin %s: %w", plugin.Source, err)
	}

	return &resp, nil
}

func (e *LocalPluginExecutor) determineCommand(source string) (string, error) {
	// This is a plugin name - add protoc-gen- prefix
	command := fmt.Sprintf("protoc-gen-%s", source)

	if command, ok := e.isPluginInPath(command); ok {
		return command, nil
	}

	// Check if this looks like a file path
	if command, ok := e.isPluginInPath(source); ok {
		return command, nil
	}

	return "", fmt.Errorf("can't determine command from source: %s", source)
}
