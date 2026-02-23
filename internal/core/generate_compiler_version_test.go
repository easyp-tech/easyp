package core

import (
	"context"
	"errors"
	"iter"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/easyp-tech/easyp/internal/adapters/console"
	pluginexecutor "github.com/easyp-tech/easyp/internal/adapters/plugin"
	"github.com/easyp-tech/easyp/internal/core/models"
	"github.com/easyp-tech/easyp/internal/logger"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type emptyLockFile struct{}

func (emptyLockFile) Read(moduleName string) (models.LockFileInfo, error) {
	return models.LockFileInfo{}, errors.New("lock file info not found")
}

func (emptyLockFile) Write(moduleName string, revisionVersion string, installedPackageHash models.ModuleHash) error {
	return nil
}

func (emptyLockFile) IsEmpty() bool {
	return true
}

func (emptyLockFile) DepsIter() iter.Seq[models.LockFileInfo] {
	return func(yield func(models.LockFileInfo) bool) {}
}

type captureExecutor struct {
	requests []*pluginpb.CodeGeneratorRequest
}

func (c *captureExecutor) Execute(_ context.Context, _ pluginexecutor.Info, request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	c.requests = append(c.requests, proto.Clone(request).(*pluginpb.CodeGeneratorRequest))
	return &pluginpb.CodeGeneratorResponse{}, nil
}

func (c *captureExecutor) GetName() string {
	return "captureExecutor"
}

func TestGenerateSetsCompilerVersionInRequest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeTestProto(t, root, "api/payment/v2/payment.proto")

	executor := &captureExecutor{}
	app := testCoreWithPlugins(
		[]Plugin{
			{
				Source: PluginSource{Name: "custom-plugin"},
				Out:    ".",
			},
		},
		executor,
	)

	if err := app.Generate(context.Background(), root, ".", "", false); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(executor.requests) != 1 {
		t.Fatalf("expected one plugin request, got %d", len(executor.requests))
	}

	compilerVersion := executor.requests[0].GetCompilerVersion()
	if compilerVersion == nil {
		t.Fatalf("compiler_version is nil")
	}

	suffix := compilerVersion.GetSuffix()
	if !strings.Contains(suffix, "bufbuild-protocompile-") {
		t.Fatalf("compiler_version suffix = %q, expected bufbuild-protocompile marker", suffix)
	}
	if !strings.Contains(suffix, "-easyp") {
		t.Fatalf("compiler_version suffix = %q, expected easyp marker", suffix)
	}
}

func TestGenerateGoHeaderUsesCompilerVersion(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		t.Skipf("protoc-gen-go not found in PATH: %v", err)
	}

	root := t.TempDir()
	writeTestProto(t, root, "api/payment/v2/payment.proto")

	plugins := []Plugin{
		{
			Source: PluginSource{Name: "go"},
			Out:    ".",
			Options: map[string][]string{
				"paths": {"source_relative"},
			},
		},
	}

	checkGRPCFile := false
	if _, err := exec.LookPath("protoc-gen-go-grpc"); err == nil {
		checkGRPCFile = true
		plugins = append(plugins, Plugin{
			Source: PluginSource{Name: "go-grpc"},
			Out:    ".",
			Options: map[string][]string{
				"paths":                         {"source_relative"},
				"require_unimplemented_servers": {"false"},
			},
		})
	}

	localExecutor := pluginexecutor.NewLocalPluginExecutor(console.New(), logger.NewNop())
	app := testCoreWithPlugins(plugins, localExecutor)

	if err := app.Generate(context.Background(), root, ".", "", false); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	goFilePath := filepath.Join(root, "api/payment/v2/payment.pb.go")
	goFileContent, err := os.ReadFile(goFilePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%s): %v", goFilePath, err)
	}

	assertNoUnknownCompilerVersion(t, string(goFileContent))
	assertCompilerVersionMarkerPresent(t, string(goFileContent))

	if !checkGRPCFile {
		t.Log("protoc-gen-go-grpc not found; gRPC file check skipped")
		return
	}

	grpcFilePath := filepath.Join(root, "api/payment/v2/payment_grpc.pb.go")
	grpcFileContent, err := os.ReadFile(grpcFilePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%s): %v", grpcFilePath, err)
	}

	assertNoUnknownCompilerVersion(t, string(grpcFileContent))
	assertCompilerVersionMarkerPresent(t, string(grpcFileContent))
}

func testCoreWithPlugins(plugins []Plugin, localExecutor pluginexecutor.Executor) *Core {
	return &Core{
		logger:  logger.NewNop(),
		plugins: plugins,
		inputs: Inputs{
			InputFilesDir: []InputFilesDir{
				{
					Path: "api",
					Root: ".",
				},
			},
		},
		lockFile:        emptyLockFile{},
		localExecutor:   localExecutor,
		remoteExecutor:  localExecutor,
		builtinExecutor: localExecutor,
		commandExecutor: localExecutor,
	}
}

func writeTestProto(t *testing.T, root string, relPath string) {
	t.Helper()

	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("os.MkdirAll(%s): %v", filepath.Dir(fullPath), err)
	}

	protoContent := `syntax = "proto3";
package api.payment.v2;

option go_package = "example.com/test/api/payment/v2;paymentv2";

message PingRequest {}
message PingResponse {}

service PaymentService {
  rpc Ping(PingRequest) returns (PingResponse);
}
`

	if err := os.WriteFile(fullPath, []byte(protoContent), 0644); err != nil {
		t.Fatalf("os.WriteFile(%s): %v", fullPath, err)
	}
}

func assertNoUnknownCompilerVersion(t *testing.T, content string) {
	t.Helper()

	if strings.Contains(content, "protoc        (unknown)") {
		t.Fatalf("generated file still contains protoc (unknown)")
	}
	if strings.Contains(content, "protoc             (unknown)") {
		t.Fatalf("generated file still contains protoc (unknown)")
	}
}

func assertCompilerVersionMarkerPresent(t *testing.T, content string) {
	t.Helper()

	if !strings.Contains(content, "bufbuild-protocompile-") {
		t.Fatalf("generated file does not contain bufbuild-protocompile marker")
	}
	if !strings.Contains(content, "-easyp") {
		t.Fatalf("generated file does not contain easyp marker")
	}
}
