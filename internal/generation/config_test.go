package generation

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
)

func TestPrepareV1GeneratorConfigPreservesPluginSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		plugin v1.Plugin
		want   core.PluginSource
	}{
		{
			name:   "named plugin",
			plugin: v1.Plugin{Name: "go", Out: "gen"},
			want:   core.PluginSource{Name: "go"},
		},
		{
			name:   "binary path",
			plugin: v1.Plugin{Path: "./tools/protoc-gen-custom", Out: "gen"},
			want:   core.PluginSource{Path: "./tools/protoc-gen-custom"},
		},
		{
			name:   "custom command",
			plugin: v1.Plugin{Command: []string{"sh", "./tools/run-plugin"}, Out: "gen"},
			want:   core.PluginSource{Command: []string{"sh", "./tools/run-plugin"}},
		},
		{
			name:   "remote plugin",
			plugin: v1.Plugin{Remote: "example.com/protoc-gen-custom", Version: "v1.2.3", Out: "gen"},
			want:   core.PluginSource{Remote: "example.com/protoc-gen-custom:v1.2.3"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			module := v1.Module{Name: "example.com/module", Roots: []string{"."}}
			gen := v1.Generate{Plugins: []v1.Plugin{tt.plugin}}

			cfg, err := prepareV1GeneratorConfig(filepath.Join(root, v1.GenerateFile), root, gen, module)

			require.NoError(t, err)
			require.Len(t, cfg.Plugins, 1)
			assert.Equal(t, tt.want, cfg.Plugins[0].Source)
		})
	}
}

func TestPrepareV1GeneratorConfigKeepsManagedOverridesIndependent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		capacity int
	}{
		{name: "append_reuses_capacity", capacity: 4},
		{name: "append_allocates", capacity: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			overrides := make([]config.ManagedOverrideRule, 1, tt.capacity)
			overrides[0] = config.ManagedOverrideRule{FileOption: "go_package_prefix", Value: "example.com/base"}
			gen := v1.Generate{}
			gen.Generate.Managed.Override = overrides
			firstPrefix, secondPrefix := "example.com/first", "example.com/second"
			gen.Options.Go.PackagePrefix = &firstPrefix
			module := v1.Module{Name: "example.com/module", Roots: []string{"."}}

			first, err := prepareV1GeneratorConfig(filepath.Join(root, v1.GenerateFile), root, gen, module)
			require.NoError(t, err)
			gen.Options.Go.PackagePrefix = &secondPrefix
			second, err := prepareV1GeneratorConfig(filepath.Join(root, v1.GenerateFile), root, gen, module)
			require.NoError(t, err)

			require.Len(t, first.ManagedModeConfig.Override, 2)
			require.Len(t, second.ManagedModeConfig.Override, 2)
			assert.Equal(t, firstPrefix, first.ManagedModeConfig.Override[1].Value)
			assert.Equal(t, secondPrefix, second.ManagedModeConfig.Override[1].Value)
			assert.Equal(t, []config.ManagedOverrideRule{{FileOption: "go_package_prefix", Value: "example.com/base"}}, gen.Generate.Managed.Override)
		})
	}
}

func TestPrepareV1GeneratorConfigPreservesManagedPriority(t *testing.T) {
	t.Parallel()

	prefix := "example.com/options"
	managedRule := config.ManagedOverrideRule{FileOption: "go_package_prefix", Value: "example.com/managed"}
	optionRule := core.ManagedOverrideRule{FileOption: "go_package_prefix", Value: prefix}
	tests := []struct {
		name          string
		prefix        string
		inherited     bool
		overrides     []config.ManagedOverrideRule
		wantOverrides []core.ManagedOverrideRule
	}{
		{
			name:          "explicit_option_overrides_managed_rule",
			prefix:        prefix,
			overrides:     []config.ManagedOverrideRule{managedRule},
			wantOverrides: []core.ManagedOverrideRule{{FileOption: "go_package_prefix", Value: "example.com/managed"}, optionRule},
		},
		{
			name:          "inherited_option_precedes_managed_rule",
			prefix:        prefix,
			inherited:     true,
			overrides:     []config.ManagedOverrideRule{managedRule},
			wantOverrides: []core.ManagedOverrideRule{optionRule, {FileOption: "go_package_prefix", Value: "example.com/managed"}},
		},
		{
			name: "omitted_overrides_remain_nil",
		},
		{
			name:          "explicit_empty_overrides_remain_empty",
			overrides:     []config.ManagedOverrideRule{},
			wantOverrides: []core.ManagedOverrideRule{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			gen := v1.Generate{InheritedGoPackagePrefix: tt.inherited}
			if tt.prefix != "" {
				prefix := tt.prefix
				gen.Options.Go.PackagePrefix = &prefix
			}
			gen.Generate.Managed.Override = tt.overrides
			module := v1.Module{Name: "example.com/module", Roots: []string{"."}}

			cfg, err := prepareV1GeneratorConfig(filepath.Join(root, v1.GenerateFile), root, gen, module)

			require.NoError(t, err)
			assert.Equal(t, tt.wantOverrides, cfg.ManagedModeConfig.Override)
		})
	}
}
