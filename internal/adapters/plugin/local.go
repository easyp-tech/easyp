package plugin

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	"github.com/easyp-tech/easyp/internal/logger"
)

// Info describes a plugin invocation and its execution directory.
type Info struct {
	Source  string
	WorkDir string
	Command []string
	Options map[string][]string
}

// LocalPluginExecutor invokes local plugin executables.
type LocalPluginExecutor struct {
	logger logger.Logger
}

// GetName identifies the local executor.
func (e *LocalPluginExecutor) GetName() string {
	return "LocalPluginExecutor from PATH"
}

// NewLocalPluginExecutor creates a new LocalPluginExecutor
func NewLocalPluginExecutor(logger logger.Logger) *LocalPluginExecutor {
	return &LocalPluginExecutor{logger: logger}
}

// isPluginInPath checks if the plugin is available in PATH
func (e *LocalPluginExecutor) isPluginInPath(source string) (string, bool) {
	command, err := exec.LookPath(source)
	return command, err == nil
}

// Execute runs a plugin in its working directory and decodes the response.
func (e *LocalPluginExecutor) Execute(ctx context.Context, plugin Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	e.logger.Debug(ctx, "executing local plugin",
		slog.String("plugin", plugin.Source),
	)

	// Prepare plugin parameters
	if parameter, ok := flattenOptions(plugin.Options); ok {
		request.Parameter = proto.String(parameter)
	}

	reqData, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("proto.Marshal request: %w", err)
	}

	workDir, err := filepath.Abs(plugin.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("Abs: %w", err)
	}
	source := plugin.Source
	if !filepath.IsAbs(source) && strings.ContainsAny(source, `/\`) {
		source = filepath.Join(workDir, source)
	}
	command, err := e.determineCommand(source)
	if err != nil {
		return nil, fmt.Errorf("determineCommand: %w", err)
	}

	// Pass the executable directly so paths are never interpreted as shell syntax.
	cmd := exec.CommandContext(ctx, command)
	cmd.Dir = workDir
	cmd.Stdin = bytes.NewReader(reqData)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		runErr := &console.RunError{Command: command, Dir: workDir, Err: err, Stderr: stderr.String()}
		return nil, fmt.Errorf("run local plugin %s: %w", plugin.Source, runErr)
	}

	// Parse response from plugin
	var resp pluginpb.CodeGeneratorResponse
	if err := proto.Unmarshal(stdout.Bytes(), &resp); err != nil {
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
