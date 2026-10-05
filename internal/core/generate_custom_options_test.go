package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/pluginpb"

	pluginexecutor "github.com/easyp-tech/easyp/internal/adapters/plugin"
	"github.com/easyp-tech/easyp/internal/logger"
)

func TestGenerateCommandPreservesCustomMethodOptions(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		managed bool
	}{
		{name: "unmanaged"},
		{name: "managed Go package", managed: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeGenerationSources(t, root, map[string]string{
				"mcp/options.proto": `syntax = "proto3";
package mcp;
import "google/protobuf/descriptor.proto";
message Method { string name = 1; string title = 2; bool hidden = 4; }
message Service { string namespace = 1; }
extend google.protobuf.MethodOptions { Method method = 91002; }
extend google.protobuf.ServiceOptions { Service service = 91001; }
`,
				"api/service.proto": `syntax = "proto3";
package api.v1;
import "mcp/options.proto";
option go_package = "example.test/api/v1;apiv1";
message Request {}
message Response {}
service GeneratorAPI {
  option (mcp.service) = {namespace: "catalog"};
  rpc GenerateCode(Request) returns (Response) { option (mcp.method) = {hidden: true}; }
  rpc Plugins(Request) returns (Response) { option (mcp.method) = {name: "plugins_list", title: "List plugins"}; }
}
`,
			})
			console := &customOptionsConsole{}
			executor := pluginexecutor.NewCommandPluginExecutor(console, logger.NewNop())
			app := testCoreWithPlugins([]Plugin{{
				Source:  PluginSource{Command: []string{"custom-plugin", "argument with spaces"}},
				Out:     "gen",
				Options: map[string][]string{"paths": {"source_relative"}},
			}}, executor)
			app.pluginWorkDir = root
			app.managedMode = ManagedModeConfig{Enabled: tt.managed, Override: []ManagedOverrideRule{{
				FileOption: FileOptionGoPackagePrefix, Value: "example.test/generated",
			}}}

			plan, err := app.PrepareGeneration(t.Context(), root)
			require.NoError(t, err)
			require.NoError(t, plan.Execute(t.Context()))

			require.NotNil(t, console.request)
			assert.Equal(t, []string{"api/service.proto"}, console.request.GetFileToGenerate())
			assert.Equal(t, "paths=source_relative", console.request.GetParameter())
			assert.Equal(t, root, console.dir)
			assert.Equal(t, []string{"custom-plugin", "argument with spaces"}, console.argv)
			assert.Equal(t, []string{"google/protobuf/descriptor.proto", "mcp/options.proto", "api/service.proto"}, descriptorNames(console.request.GetProtoFile()))
			generated, err := os.ReadFile(filepath.Join(root, "gen/tools.txt"))
			require.NoError(t, err)
			assert.Equal(t, "catalog.plugins_list: List plugins\n", string(generated))
		})
	}
}

// This plugin reads the serialized request with its own extension resolver, as
// an external command does; inspecting the compiler's in-memory options alone
// would miss losses when requests are cloned, managed, or marshaled.
type customOptionsConsole struct {
	request *pluginpb.CodeGeneratorRequest
	dir     string
	argv    []string
}

func (*customOptionsConsole) RunCmd(context.Context, string, string, ...string) (string, error) {
	return "", fmt.Errorf("unexpected command without stdin")
}

func (c *customOptionsConsole) RunCmdWithStdin(_ context.Context, dir string, stdin io.Reader, command string, args ...string) (string, error) {
	raw, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("ReadAll: %w", err)
	}
	var request pluginpb.CodeGeneratorRequest
	if err := proto.Unmarshal(raw, &request); err != nil {
		return "", fmt.Errorf("Unmarshal: %w", err)
	}
	files, err := protodesc.NewFiles(&descriptorpb.FileDescriptorSet{File: request.GetProtoFile()})
	if err != nil {
		return "", fmt.Errorf("NewFiles: %w", err)
	}
	types := dynamicpb.NewTypes(files)
	if err := (proto.UnmarshalOptions{Resolver: types}).Unmarshal(raw, &request); err != nil {
		return "", fmt.Errorf("Unmarshal: %w", err)
	}
	c.request, c.dir, c.argv = &request, dir, append([]string{command}, args...)
	methodOption, err := types.FindExtensionByName("mcp.method")
	if err != nil {
		return "", fmt.Errorf("FindExtensionByName: %w", err)
	}
	serviceOption, err := types.FindExtensionByName("mcp.service")
	if err != nil {
		return "", fmt.Errorf("FindExtensionByName: %w", err)
	}
	var content string
	for _, file := range request.GetProtoFile() {
		if file.GetName() != request.GetFileToGenerate()[0] {
			continue
		}
		for _, service := range file.GetService() {
			metadata := proto.GetExtension(service.GetOptions(), serviceOption).(proto.Message).ProtoReflect()
			namespace := metadata.Get(metadata.Descriptor().Fields().ByName("namespace")).String()
			for _, method := range service.GetMethod() {
				metadata := proto.GetExtension(method.GetOptions(), methodOption).(proto.Message).ProtoReflect()
				fields := metadata.Descriptor().Fields()
				if metadata.Get(fields.ByName("hidden")).Bool() {
					continue
				}
				name := metadata.Get(fields.ByName("name")).String()
				title := metadata.Get(fields.ByName("title")).String()
				content += namespace + "." + name + ": " + title + "\n"
			}
		}
	}
	response, err := proto.Marshal(&pluginpb.CodeGeneratorResponse{File: []*pluginpb.CodeGeneratorResponse_File{{
		Name: proto.String("tools.txt"), Content: proto.String(content),
	}}})
	if err != nil {
		return "", fmt.Errorf("Marshal: %w", err)
	}
	return string(response), nil
}
