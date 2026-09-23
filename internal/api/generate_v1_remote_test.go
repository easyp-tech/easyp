package api

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
	pluginv1 "github.com/easyp-tech/service/api/generator/v1"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type testV1RemotePlugin struct {
	pluginv1.UnimplementedServiceAPIServer
	called chan string
}

func (s *testV1RemotePlugin) GenerateCode(_ context.Context, request *pluginv1.GenerateCodeRequest) (*pluginv1.GenerateCodeResponse, error) {
	s.called <- request.GetPluginName()
	return &pluginv1.GenerateCodeResponse{CodeGeneratorResponse: &pluginpb.CodeGeneratorResponse{
		File: []*pluginpb.CodeGeneratorResponse_File{{Name: proto.String("remote.txt"), Content: proto.String("ok")}},
	}}, nil
}

func TestGenerateV1SendsPinnedRemotePluginVersion(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	remote := &testV1RemotePlugin{called: make(chan string, 1)}
	pluginv1.RegisterServiceAPIServer(server, remote)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "protobuf.mod"), []byte("module example.com/root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "item.proto"), []byte("syntax = \"proto3\";\npackage item.v1;\nmessage Item {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	configText := fmt.Sprintf("version: v1\nplugins:\n  - remote: http://%s/python\n    version: v1.2.3\n    out: gen\n", listener.Addr())
	gen, err := v1.ParseGenerate(strings.NewReader(configText))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("EASYPPATH", filepath.Join(t.TempDir(), "cache"))
	ctx := cli.NewContext(&cli.App{Metadata: map[string]any{}}, flag.NewFlagSet("test", flag.ContinueOnError), nil)
	ctx.Context = context.Background()
	if err := generateSelectedV1Module(ctx, logger.NewNop(), filepath.Join(root, "easyp.gen.yaml"), root, v1ModuleSelection{directory: root}, gen); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-remote.called:
		if got != "python:v1.2.3" {
			t.Fatalf("remote plugin name = %q", got)
		}
	default:
		t.Fatal("remote plugin was not called")
	}
	if _, err := os.Stat(filepath.Join(root, "gen", "remote.txt")); err != nil {
		t.Fatal(err)
	}
}
