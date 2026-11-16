package protoc_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/easyp-tech/easyp/internal/easypc/protoc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// REQ-8.1 & REQ-8.8: In-memory WireBytes() is bit-for-bit identical to CLI output.
func TestCLIAndLibraryEquivalence(t *testing.T) {
	tmpDir := t.TempDir()
	protoFile := filepath.Join(tmpDir, "test.proto")
	if err := os.WriteFile(protoFile, []byte(`syntax = "proto3";
package sample;

message Item {
  string id = 1;
  int32 count = 2;
}
`), 0644); err != nil {
		t.Fatal(err)
	}
	cliOut := filepath.Join(tmpDir, "cli.fds")

	// Run CLI
	binPath := filepath.Join(tmpDir, "easyp-protoc")
	if out, err := exec.Command("go", "build", "-o", binPath, "../cmd/easyp-protoc").CombinedOutput(); err != nil {
		t.Fatalf("build CLI failed: %v\nOutput: %s", err, string(out))
	}
	cmd := exec.Command(binPath, "-I", tmpDir, "-o", cliOut, "test.proto")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("CLI failed: %v\nOutput: %s", err, string(out))
	}
	cliBytes, err := os.ReadFile(cliOut)
	if err != nil {
		t.Fatal(err)
	}

	// Call Go library
	res, err := protoc.Compile(
		context.Background(),
		[]string{"test.proto"},
		protoc.WithImportPaths(tmpDir),
	)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if !bytes.Equal(res.WireBytes(), cliBytes) {
		t.Fatalf("WireBytes() mismatch with CLI: library len %d != CLI len %d", len(res.WireBytes()), len(cliBytes))
	}
}

// REQ-8.2: MemResolver compiles proto definitions from memory.
func TestMemResolver(t *testing.T) {
	mem := protoc.MemResolver{
		"my_test.proto": []byte(`syntax = "proto3";
package test.pkg;

message SampleMessage {
  string name = 1;
  int32 id = 2;
}
`),
	}

	res, err := protoc.Compile(
		context.Background(),
		[]string{"my_test.proto"},
		protoc.WithResolver(mem),
	)
	if err != nil {
		t.Fatalf("Compile with MemResolver failed: %v", err)
	}

	ds := res.DescriptorSet()
	if ds == nil || len(ds.GetFile()) != 1 {
		t.Fatalf("Expected 1 file descriptor, got %v", ds)
	}
	fd := ds.GetFile()[0]
	if fd.GetPackage() != "test.pkg" {
		t.Errorf("Expected package test.pkg, got %s", fd.GetPackage())
	}
	if len(fd.GetMessageType()) != 1 || fd.GetMessageType()[0].GetName() != "SampleMessage" {
		t.Errorf("Expected SampleMessage in message types")
	}
}

// REQ-8.3: MultiResolver checks resolvers in order, taking the first match.
func TestMultiResolverPrecedence(t *testing.T) {
	r1 := protoc.MemResolver{
		"shared.proto": []byte(`syntax = "proto3";
package first;
message FirstMessage {}
`),
	}
	r2 := protoc.MemResolver{
		"shared.proto": []byte(`syntax = "proto3";
package second;
message SecondMessage {}
`),
	}

	res, err := protoc.Compile(
		context.Background(),
		[]string{"shared.proto"},
		protoc.WithResolver(r1),
		protoc.WithResolver(r2),
	)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	ds := res.DescriptorSet()
	if len(ds.GetFile()) != 1 || ds.GetFile()[0].GetPackage() != "first" {
		t.Fatalf("Expected first resolver to take precedence, got: %v", ds.GetFile())
	}
}

// REQ-8.4: Standard Well-Known Types resolve automatically with WithStandardImports(true).
func TestStandardImports(t *testing.T) {
	mem := protoc.MemResolver{
		"event.proto": []byte(`syntax = "proto3";
package test;

import "google/protobuf/timestamp.proto";

message Event {
  string id = 1;
  google.protobuf.Timestamp timestamp = 2;
}
`),
	}

	// Without standard imports, should fail
	_, err := protoc.Compile(
		context.Background(),
		[]string{"event.proto"},
		protoc.WithResolver(mem),
		protoc.WithStandardImports(false),
	)
	if err == nil {
		t.Fatal("Expected compilation to fail when standard imports are disabled and WKT not on disk")
	}

	// With standard imports, should succeed
	res, err := protoc.Compile(
		context.Background(),
		[]string{"event.proto"},
		protoc.WithResolver(mem),
		protoc.WithStandardImports(true),
		protoc.WithIncludeImports(true),
	)
	if err != nil {
		t.Fatalf("Expected compilation to succeed with standard imports: %v", err)
	}

	ds := res.DescriptorSet()
	if len(ds.GetFile()) < 2 {
		t.Fatalf("Expected at least 2 files in DescriptorSet with imports, got %d", len(ds.GetFile()))
	}
}

// REQ-8.5: WithErrorHandler receives diagnostics with path, line, col, and message.
func TestWithErrorHandler(t *testing.T) {
	mem := protoc.MemResolver{
		"bad.proto": []byte(`syntax = "proto4";
message Foo {}
`),
	}

	var diags []protoc.Diagnostic
	_, err := protoc.Compile(
		context.Background(),
		[]string{"bad.proto"},
		protoc.WithResolver(mem),
		protoc.WithErrorHandler(func(d protoc.Diagnostic) {
			diags = append(diags, d)
		}),
	)
	if err == nil {
		t.Fatal("Expected error compilation")
	}
	if !errors.Is(err, protoc.ErrCompilationFailed) {
		t.Errorf("Expected ErrCompilationFailed, got %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("Expected at least 1 diagnostic recorded")
	}
	if diags[0].Line != 1 {
		t.Errorf("Expected line 1, got %d", diags[0].Line)
	}
	if diags[0].Path != "bad.proto" {
		t.Errorf("Expected path bad.proto, got %s", diags[0].Path)
	}
}

// REQ-8.6: WithIncludeSourceInfo and WithConcurrency options are respected.
func TestSourceInfoAndConcurrency(t *testing.T) {
	mem := protoc.MemResolver{
		"doc.proto": []byte(`syntax = "proto3";
// Leading comment for message
message Document {
  // Leading comment for field
  string title = 1;
}
`),
	}

	resWithSI, err := protoc.Compile(
		context.Background(),
		[]string{"doc.proto"},
		protoc.WithResolver(mem),
		protoc.WithIncludeSourceInfo(true),
		protoc.WithConcurrency(2),
	)
	if err != nil {
		t.Fatalf("Compile with source info failed: %v", err)
	}

	ds := resWithSI.DescriptorSet()
	if ds == nil || len(ds.GetFile()) != 1 {
		t.Fatal("Expected 1 file descriptor")
	}
	if ds.GetFile()[0].SourceCodeInfo == nil || len(ds.GetFile()[0].SourceCodeInfo.Location) == 0 {
		t.Error("Expected SourceCodeInfo locations with WithIncludeSourceInfo(true)")
	}

	resNoSI, err := protoc.Compile(
		context.Background(),
		[]string{"doc.proto"},
		protoc.WithResolver(mem),
		protoc.WithIncludeSourceInfo(false),
	)
	if err != nil {
		t.Fatalf("Compile without source info failed: %v", err)
	}
	dsNo := resNoSI.DescriptorSet()
	if dsNo.GetFile()[0].SourceCodeInfo != nil {
		t.Error("Expected nil SourceCodeInfo with WithIncludeSourceInfo(false)")
	}
}

// REQ-8.7: Files() returns protoreflect.FileDescriptor instances.
func TestProtoreflectFiles(t *testing.T) {
	mem := protoc.MemResolver{
		"types.proto": []byte(`syntax = "proto3";
package test.types;

message Person {
  string name = 1;
  int32 age = 2;
}
`),
	}

	res, err := protoc.Compile(
		context.Background(),
		[]string{"types.proto"},
		protoc.WithResolver(mem),
	)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	files := res.Files()
	if len(files) != 1 {
		t.Fatalf("Expected 1 protoreflect.FileDescriptor, got %d", len(files))
	}

	fd := files[0]
	if fd.Package() != protoreflect.FullName("test.types") {
		t.Errorf("Expected package test.types, got %s", fd.Package())
	}
	md := fd.Messages().ByName("Person")
	if md == nil {
		t.Fatal("Expected Person message descriptor")
	}
	if md.Fields().ByName("name") == nil {
		t.Error("Expected field name in Person")
	}
}

// REQ-8.9: CodeGeneratorRequest contains files_to_generate, parameters, and version.
func TestCodeGeneratorRequest(t *testing.T) {
	mem := protoc.MemResolver{
		"api.proto": []byte(`syntax = "proto3";
package test.api;

message Request {
  string query = 1;
}
`),
	}

	res, err := protoc.Compile(
		context.Background(),
		[]string{"api.proto"},
		protoc.WithResolver(mem),
	)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	req, err := res.CodeGeneratorRequest("paths=source_relative,custom_opt=123")
	if err != nil {
		t.Fatalf("CodeGeneratorRequest failed: %v", err)
	}

	if req.GetParameter() != "paths=source_relative,custom_opt=123" {
		t.Errorf("Parameter mismatch: got %q", req.GetParameter())
	}
	if len(req.GetFileToGenerate()) != 1 || req.GetFileToGenerate()[0] != "api.proto" {
		t.Errorf("FileToGenerate mismatch: %v", req.GetFileToGenerate())
	}
	if req.GetCompilerVersion() == nil || req.GetCompilerVersion().GetMajor() != 36 || req.GetCompilerVersion().GetMinor() != 2 {
		t.Errorf("CompilerVersion mismatch: %v", req.GetCompilerVersion())
	}
	if len(req.GetProtoFile()) != 1 {
		t.Errorf("Expected 1 ProtoFile, got %d", len(req.GetProtoFile()))
	}
	if len(req.GetSourceFileDescriptors()) != 1 {
		t.Errorf("Expected 1 SourceFileDescriptor, got %d", len(req.GetSourceFileDescriptors()))
	}
}

// REQ-8.10: Context cancellation aborts compilation immediately.
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mem := protoc.MemResolver{
		"dummy.proto": []byte(`syntax = "proto3"; message Dummy {}`),
	}

	_, err := protoc.Compile(ctx, []string{"dummy.proto"}, protoc.WithResolver(mem))
	if err == nil {
		t.Fatal("Expected cancelled context to fail compilation")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

// REQ-8.11: Thread safety: 10 concurrent goroutines calling Compile simultaneously.
func TestConcurrentCompilations(t *testing.T) {
	const numGoroutines = 10
	var wg sync.WaitGroup
	errs := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			fileName := fmt.Sprintf("tenant_%d.proto", idx)
			pkgName := fmt.Sprintf("tenant.pkg%d", idx)
			msgName := fmt.Sprintf("TenantMessage%d", idx)

			src := fmt.Sprintf(`syntax = "proto3";
package %s;
message %s {
  int64 id = 1;
}
`, pkgName, msgName)

			mem := protoc.MemResolver{
				fileName: []byte(src),
			}

			res, err := protoc.Compile(
				context.Background(),
				[]string{fileName},
				protoc.WithResolver(mem),
			)
			if err != nil {
				errs[idx] = err
				return
			}

			ds := res.DescriptorSet()
			if ds == nil || len(ds.GetFile()) != 1 {
				errs[idx] = fmt.Errorf("unexpected descriptor set: %v", ds)
				return
			}
			if ds.GetFile()[0].GetPackage() != pkgName {
				errs[idx] = fmt.Errorf("expected package %s, got %s", pkgName, ds.GetFile()[0].GetPackage())
				return
			}
			files := res.Files()
			if len(files) != 1 || files[0].Messages().ByName(protoreflect.Name(msgName)) == nil {
				errs[idx] = fmt.Errorf("expected message %s in files", msgName)
				return
			}
		}(i)
	}

	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("Goroutine %d failed: %v", i, err)
		}
	}
}
