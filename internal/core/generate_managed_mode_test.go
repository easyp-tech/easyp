package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestGenerateV1ManagedModeAppliesToModuleFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		module      string
		wantPackage string
	}{
		{name: "matching module", module: "example.com/contracts", wantPackage: "pinger-service/internal/grpc/gen/pinger/v1;pingerv1"},
		{name: "different module", module: "example.com/other"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeProtoWithoutGoPackage(t, root, "pinger/v1/service.proto")
			executor := &captureExecutor{}
			identities := map[string]string{"pinger/v1/service.proto": tt.module}
			app := New(Options{
				Logger:            logger.NewNop(),
				Plugins:           []Plugin{{Source: PluginSource{Name: "custom-plugin"}, Out: "."}},
				Inputs:            Inputs{InputFilesDir: []InputFilesDir{{Path: "pinger", Root: "."}}},
				FileModules:       identities,
				ManagedModeConfig: ManagedModeConfig{Enabled: true, Override: []ManagedOverrideRule{{Module: "example.com/contracts", FileOption: FileOptionGoPackagePrefix, Value: "pinger-service/internal/grpc/gen"}}},
			})
			app.localExecutor = executor
			identities["pinger/v1/service.proto"] = "changed after construction"

			err := app.Generate(t.Context(), root, "", false)

			require.NoError(t, err)
			require.Len(t, executor.requests, 1)
			request := executor.requests[0]
			assert.Equal(t, []string{"pinger/v1/service.proto"}, request.GetFileToGenerate())
			target := findFileDescriptor(t, request.GetProtoFile(), "pinger/v1/service.proto")
			assert.Equal(t, tt.wantPackage, target.GetOptions().GetGoPackage())
		})
	}
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
