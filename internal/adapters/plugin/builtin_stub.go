//go:build !builtin_plugins
// +build !builtin_plugins

package plugin

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/protobuf/types/pluginpb"
)

// BuiltinPluginExecutor is a stub when builtin_plugins tag is not set
type BuiltinPluginExecutor struct {
	logger *slog.Logger
}

// NewBuiltinPluginExecutor creates a stub executor when builtin plugins are disabled
func NewBuiltinPluginExecutor(logger *slog.Logger) *BuiltinPluginExecutor {
	return &BuiltinPluginExecutor{
		logger: logger,
	}
}

// GetName returns the name of the executor
func (e *BuiltinPluginExecutor) GetName() string {
	return "BuiltinPluginExecutor (disabled)"
}

// Execute returns an error indicating that builtin plugins are disabled
func (e *BuiltinPluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	return nil, fmt.Errorf("builtin plugins are disabled (build without builtin_plugins tag)")
}

func builtinPluginExecutorEnabled() bool {
	return false
}
