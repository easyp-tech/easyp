package bsr

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

func TestStaticResolver(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		module  string
		git     v1.Requirement
		wantErr bool
	}{
		{name: "googleapis", module: "buf.build/googleapis/googleapis", git: v1.Requirement{Module: "github.com/googleapis/googleapis", Version: "03a91044136a014466d4293eb1fe91f2b02075d2"}},
		{name: "gateway", module: "buf.build/grpc-ecosystem/grpc-gateway", git: v1.Requirement{Module: "github.com/grpc-ecosystem/grpc-gateway", Version: "v2.31.0+incompatible"}},
		{name: "federation", module: "buf.build/mercari/grpc-federation", git: v1.Requirement{Module: "github.com/mercari/grpc-federation", Version: "v1.27.0"}},
		{name: "PGV", module: "buf.build/envoyproxy/protoc-gen-validate", git: v1.Requirement{Module: "github.com/envoyproxy/protoc-gen-validate", Version: "v1.3.3"}},
		{name: "protovalidate", module: "buf.build/bufbuild/protovalidate", git: v1.Requirement{Module: "github.com/bufbuild/protovalidate", Version: "v1.2.2"}},
		{name: "unknown module", module: "buf.build/acme/missing", wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dependency := v1.BSRDependency{Module: tt.module, Reference: "stable", Commit: strings.Repeat("a", 32), Digest: "b5:" + strings.Repeat("a", 128), Config: "proto/buf.yaml"}

			binding, err := (StaticResolver{}).Resolve(t.Context(), dependency)

			if tt.wantErr {
				var unsupported *UnsupportedModuleError
				require.ErrorAs(t, err, &unsupported)
				assert.Equal(t, tt.module, unsupported.Module)
				assert.Contains(t, err.Error(), "unsupported BSR module: "+tt.module)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, dependency, binding.Dependency)
			assert.Equal(t, tt.git, binding.Git)
			assert.Equal(t, v1.BSRCompatibilitySnapshot, binding.Resolution)
			require.NoError(t, binding.Validate())
		})
	}
}

func TestStaticResolverCancellation(t *testing.T) {
	t.Parallel()
	tests := []struct{ name string }{{name: "cancelled before resolution"}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			_, err := (StaticResolver{}).Resolve(ctx, v1.BSRDependency{Module: "buf.build/googleapis/googleapis", Config: "buf.yaml"})

			require.ErrorIs(t, err, context.Canceled)
		})
	}
}
