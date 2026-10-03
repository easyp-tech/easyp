package v1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBSRDependencyValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		dependency BSRDependency
		wantErr    string
	}{
		{name: "unversioned module", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis", Config: "buf.yaml"}},
		{name: "label and locked pin", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis", Reference: "stable", Commit: strings.Repeat("a", 32), Digest: "shake256:" + strings.Repeat("a", 128), Config: "proto/buf.yaml"}},
		{name: "empty module", dependency: BSRDependency{Config: "buf.yaml"}, wantErr: "invalid BSR module"},
		{name: "reference in identity", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis:stable", Config: "buf.yaml"}, wantErr: "invalid BSR module"},
		{name: "invalid reference", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis", Reference: "two words", Config: "buf.yaml"}, wantErr: "invalid BSR reference"},
		{name: "Git commit is not BSR commit", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis", Commit: strings.Repeat("a", 40), Config: "buf.yaml"}, wantErr: "invalid BSR commit"},
		{name: "conflicting commit", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis", Reference: strings.Repeat("a", 32), Commit: strings.Repeat("b", 32), Config: "buf.yaml"}, wantErr: "conflicting BSR commit"},
		{name: "invalid digest", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis", Digest: "b5:not-hex", Config: "buf.yaml"}, wantErr: "invalid BSR digest"},
		{name: "missing origin", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis"}, wantErr: "invalid BSR config path"},
		{name: "origin outside module", dependency: BSRDependency{Module: "buf.build/googleapis/googleapis", Config: "../buf.yaml"}, wantErr: "invalid BSR config path"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.dependency.Validate()

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestLockBSRBindings(t *testing.T) {
	t.Parallel()
	binding := "      - dependency: {module: buf.build/googleapis/googleapis, config: buf.yaml}\n        git: {module: github.com/googleapis/googleapis, version: v1.0.0}\n        resolution: compatibility_snapshot\n"
	tests := []struct {
		name    string
		binding string
		wantErr string
	}{
		{name: "compatibility snapshot", binding: binding},
		{name: "duplicate origin", binding: binding + binding, wantErr: "duplicate BSR binding"},
		{name: "missing Git version", binding: strings.Replace(binding, ", version: v1.0.0", "", 1), wantErr: "unpinned BSR Git target"},
		{name: "unknown policy", binding: strings.Replace(binding, "compatibility_snapshot", "implicit_latest", 1), wantErr: "invalid BSR resolution"},
		{name: "invalid target version", binding: strings.Replace(binding, "version: v1.0.0", "version: main", 1), wantErr: "invalid version"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := "version: 1\nmodules:\n  - source: github.com/acme/parent\n    version: v1.0.0\n    commit: " + strings.Repeat("a", 40) + "\n    hash: h1:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n    bsr:\n" + tt.binding

			lock, err := ParseLock(strings.NewReader(raw))

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Len(t, lock.Modules, 1)
			require.Len(t, lock.Modules[0].BSR, 1)
			assert.Equal(t, "buf.build/googleapis/googleapis", lock.Modules[0].BSR[0].Dependency.Module)
			assert.Equal(t, Requirement{Module: "github.com/googleapis/googleapis", Version: "v1.0.0"}, lock.Modules[0].BSR[0].Git)
		})
	}
}
