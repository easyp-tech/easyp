package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGenerateV1ManagedModeAppliesToModuleFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeProtoWithoutGoPackage(t, root, "pinger/v1/service.proto")
	executor := &captureExecutor{}
	app := testCoreWithPlugins([]Plugin{{Source: PluginSource{Name: "custom-plugin"}, Out: "."}}, executor)
	app.inputs.InputFilesDir = []InputFilesDir{{Path: "pinger", Root: "."}}
	app.managedMode = ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{{
			Module:     "example.com/contracts",
			FileOption: FileOptionGoPackagePrefix,
			Value:      "pinger-service/internal/grpc/gen",
		}},
	}
	app.SetFileModules(map[string]string{"pinger/v1/service.proto": "example.com/contracts"})

	require.NoError(t, app.Generate(t.Context(), root, "", false))
	require.Len(t, executor.requests, 1)
	req := executor.requests[0]
	require.Equal(t, []string{"pinger/v1/service.proto"}, req.GetFileToGenerate())
	target := findFileDescriptor(t, req.GetProtoFile(), "pinger/v1/service.proto")
	require.Equal(t, "pinger-service/internal/grpc/gen/pinger/v1;pingerv1", target.GetOptions().GetGoPackage())
}

func writeProtoWithoutGoPackage(t *testing.T, root, relPath string) {
	t.Helper()
	fullPath := filepath.Join(root, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0o755))
	protoContent := `syntax = "proto3";
package pinger.v1;

message PingRequest {}
`
	require.NoError(t, os.WriteFile(fullPath, []byte(protoContent), 0o644))
}

func findFileDescriptor(t *testing.T, files []*descriptorpb.FileDescriptorProto, name string) *descriptorpb.FileDescriptorProto {
	t.Helper()
	for _, file := range files {
		if file.GetName() == name {
			return file
		}
	}
	t.Fatalf("file descriptor %q not found", name)
	return nil
}
