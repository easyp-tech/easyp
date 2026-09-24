package generation

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"testing"

	pluginv1 "github.com/easyp-tech/service/api/generator/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/logger"
)

type testV1RemotePlugin struct {
	pluginv1.UnimplementedServiceAPIServer
	called chan string
}

func (s *testV1RemotePlugin) GenerateCode(_ context.Context, request *pluginv1.GenerateCodeRequest) (*pluginv1.GenerateCodeResponse, error) {
	select {
	case s.called <- request.GetPluginName():
	default:
		return nil, fmt.Errorf("unexpected repeated GenerateCode call for %q", request.GetPluginName())
	}
	return &pluginv1.GenerateCodeResponse{CodeGeneratorResponse: &pluginpb.CodeGeneratorResponse{
		File: []*pluginpb.CodeGeneratorResponse_File{{Name: proto.String("remote.txt"), Content: proto.String("ok")}},
	}}, nil
}

func TestGenerateV1SendsPinnedRemotePluginVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		version     string
		wantRequest string
	}{
		{name: "release", version: "v1.2.3", wantRequest: "python:v1.2.3"},
		{name: "prerelease", version: "v2.0.0-rc.1", wantRequest: "python:v2.0.0-rc.1"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			server := grpc.NewServer()
			remote := &testV1RemotePlugin{called: make(chan string, 1)}
			pluginv1.RegisterServiceAPIServer(server, remote)
			serveErr := make(chan error, 1)
			go func() { serveErr <- server.Serve(listener) }()
			t.Cleanup(func() {
				server.Stop()
				assert.NoError(t, <-serveErr)
			})

			root := t.TempDir()
			writeV1GenerateFixture(t, root, "item.proto", "syntax = \"proto3\";\npackage item.v1;\nmessage Item {}\n")
			configText := fmt.Sprintf("version: v1\nplugins:\n  - remote: http://%s/python\n    version: %s\n    out: gen\n", listener.Addr(), tt.version)
			gen, err := v1.ParseGenerate(strings.NewReader(configText))
			require.NoError(t, err)
			module := v1.Module{Name: "example.com/root", Roots: []string{"."}}

			err = generateV1ModuleWithRoots(t.Context(), logger.NewNop(), Request{}, filepath.Join(root, "easyp.gen.yaml"), root, gen, module, nil)

			require.NoError(t, err)
			require.Len(t, remote.called, 1)
			assert.Equal(t, tt.wantRequest, <-remote.called)
			assert.FileExists(t, filepath.Join(root, "gen", "remote.txt"))
		})
	}
}
