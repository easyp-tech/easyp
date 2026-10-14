package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKindOf(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want FileKind
	}{
		{path: "easyp.yaml", want: PolicyFileKind},
		{path: "nested/strict.easyp.yaml", want: PolicyFileKind},
		{path: "easyp.gen.yaml", want: GenerateFileKind},
		{path: "/abs/private.easyp.gen.yaml", want: GenerateFileKind},
		{path: "protobuf.mod", want: ModuleFileKind},
		{path: "protobuf.lock", want: LockFileKind},
		{path: "strict.yaml", want: UnknownFileKind},
		{path: "myeasyp.yaml", want: UnknownFileKind},
		{path: "easyp.yaml.v0.bak", want: UnknownFileKind},
		{path: "private.protobuf.mod", want: UnknownFileKind},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, KindOf(tt.path))
		})
	}
}

func TestRequireKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		want    FileKind
		wantErr string
	}{
		{name: "matching canonical name", path: "easyp.yaml", want: PolicyFileKind},
		{name: "matching profile", path: "private.easyp.gen.yaml", want: GenerateFileKind},
		{name: "unknown name is accepted", path: "strict.yaml", want: PolicyFileKind},
		{name: "generation file for policy", path: "easyp.gen.yaml", want: PolicyFileKind, wantErr: "easyp.gen.yaml is a generation config; this command expects a policy config"},
		{name: "policy profile for generation", path: "strict.easyp.yaml", want: GenerateFileKind, wantErr: "strict.easyp.yaml is a policy config; this command expects a generation config"},
		{name: "manifest for policy", path: "protobuf.mod", want: PolicyFileKind, wantErr: "is a module manifest"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := RequireKind(tt.path, tt.want)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}
