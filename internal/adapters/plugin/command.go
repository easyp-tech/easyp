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

// CommandPluginExecutor executes plugins via custom command
type CommandPluginExecutor struct {
	console console.Console
	logger  *slog.Logger
}

// NewCommandPluginExecutor creates a new CommandPluginExecutor
func NewCommandPluginExecutor(console console.Console, logger *slog.Logger) *CommandPluginExecutor {
	return &CommandPluginExecutor{
		console: console,
		logger:  logger,
	}
}

// GetName returns the name of the executor
func (e *CommandPluginExecutor) GetName() string {
	return "CommandPluginExecutor"
}

// Execute executes a plugin via custom command
func (e *CommandPluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	// Use Command if available, otherwise fall back to parsing Source
	commandParts := plugin.Command
	if len(commandParts) == 0 {
		// Fallback: parse command from source (for backward compatibility)
		commandParts = strings.Fields(plugin.Source)
	}

	if len(commandParts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	e.logger.DebugContext(ctx, "executing plugin via command",
		slog.String("command", strings.Join(commandParts, " ")),
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

	// Execute command using console
	command := commandParts[0]
	var commandParams []string
	if len(commandParts) > 1 {
		commandParams = commandParts[1:]
	}

	stdout, err := e.console.RunCmdWithStdin(ctx, ".", stdIn, command, commandParams...)
	if err != nil {
		return nil, fmt.Errorf("run command %s: %w", strings.Join(commandParts, " "), err)
	}

	// Parse response from plugin
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal([]byte(stdout), &resp); err != nil {
		return nil, fmt.Errorf("proto.Unmarshal response from command %s: %w", strings.Join(commandParts, " "), err)
	}

	return &resp, nil
}
