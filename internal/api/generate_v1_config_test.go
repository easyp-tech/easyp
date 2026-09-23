package api

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
	v1 "github.com/easyp-tech/easyp/internal/config/v1"
)

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

			require.Len(t, first.Generate.Managed.Override, 2)
			require.Len(t, second.Generate.Managed.Override, 2)
			assert.Equal(t, firstPrefix, first.Generate.Managed.Override[1].Value)
			assert.Equal(t, secondPrefix, second.Generate.Managed.Override[1].Value)
			assert.Equal(t, []config.ManagedOverrideRule{{FileOption: "go_package_prefix", Value: "example.com/base"}}, gen.Generate.Managed.Override)
		})
	}
}

func TestPrepareV1GeneratorConfigPreservesManagedPriority(t *testing.T) {
	t.Parallel()

	prefix := "example.com/options"
	managedRule := config.ManagedOverrideRule{FileOption: "go_package_prefix", Value: "example.com/managed"}
	optionRule := config.ManagedOverrideRule{FileOption: "go_package_prefix", Value: prefix}
	tests := []struct {
		name          string
		prefix        string
		inherited     bool
		overrides     []config.ManagedOverrideRule
		wantOverrides []config.ManagedOverrideRule
	}{
		{
			name:          "explicit_option_overrides_managed_rule",
			prefix:        prefix,
			overrides:     []config.ManagedOverrideRule{managedRule},
			wantOverrides: []config.ManagedOverrideRule{managedRule, optionRule},
		},
		{
			name:          "inherited_option_precedes_managed_rule",
			prefix:        prefix,
			inherited:     true,
			overrides:     []config.ManagedOverrideRule{managedRule},
			wantOverrides: []config.ManagedOverrideRule{optionRule, managedRule},
		},
		{
			name: "omitted_overrides_remain_nil",
		},
		{
			name:          "explicit_empty_overrides_remain_empty",
			overrides:     []config.ManagedOverrideRule{},
			wantOverrides: []config.ManagedOverrideRule{},
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
			assert.Equal(t, tt.wantOverrides, cfg.Generate.Managed.Override)
		})
	}
}
