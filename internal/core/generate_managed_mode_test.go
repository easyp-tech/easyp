package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/easyp-tech/easyp/internal/core/models"
	"github.com/easyp-tech/easyp/internal/logger"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestGenerateManagedModeAppliesToGitRepoInputs(t *testing.T) {
	t.Parallel()

	workspaceRoot := t.TempDir()
	gitRepoRoot := t.TempDir()
	writeProtoWithoutGoPackage(t, gitRepoRoot, "pinger/v1/service.proto")

	executor := &captureExecutor{}

	moduleName := "http://gitlab.com/foo/contracts"
	moduleVersion := "v0.0.1"
	lockInfo := models.LockFileInfo{
		Name:    moduleName,
		Version: moduleVersion,
	}

	storage := &StorageMock{}
	lockFile := &LockFileMock{}

	lockFile.EXPECT().IsEmpty().Return(false).Once()
	lockFile.EXPECT().DepsIter().Return(func(yield func(models.LockFileInfo) bool) {}).Twice()
	lockFile.EXPECT().Read(moduleName).Return(lockInfo, nil).Twice()
	storage.EXPECT().GetInstallDir(moduleName, moduleVersion).Return(gitRepoRoot).Twice()

	app := &Core{
		logger: logger.NewNop(),
		plugins: []Plugin{
			{
				Source: PluginSource{Name: "custom-plugin"},
				Out:    ".",
			},
		},
		inputs: Inputs{
			InputGitRepos: []InputGitRepo{
				{
					URL:          moduleName + "@" + moduleVersion,
					SubDirectory: "pinger/v1",
				},
			},
		},
		storage:         storage,
		lockFile:        lockFile,
		managedMode:     ManagedModeConfig{Enabled: true, Override: []ManagedOverrideRule{{FileOption: FileOptionGoPackagePrefix, Value: "pinger-service/internal/grpc/gen"}}},
		localExecutor:   executor,
		remoteExecutor:  executor,
		builtinExecutor: executor,
		commandExecutor: executor,
	}

	err := app.Generate(context.Background(), workspaceRoot, ".", "", false)
	require.NoError(t, err)
	require.Len(t, executor.requests, 1)

	req := executor.requests[0]
	require.Equal(t, []string{"pinger/v1/service.proto"}, req.GetFileToGenerate())

	target := findFileDescriptor(t, req.GetProtoFile(), "pinger/v1/service.proto")
	require.Equal(t, "pinger-service/internal/grpc/gen/pinger/v1;pingerv1", target.GetOptions().GetGoPackage())
}

func writeProtoWithoutGoPackage(t *testing.T, root string, relPath string) {
	t.Helper()

	fullPath := filepath.Join(root, relPath)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	require.NoError(t, err)

	protoContent := `syntax = "proto3";
package pinger.v1;

message PingRequest {}
`

	err = os.WriteFile(fullPath, []byte(protoContent), 0644)
	require.NoError(t, err)
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
