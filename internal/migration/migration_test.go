package migration

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	v1 "github.com/easyp-tech/easyp/internal/config/v1"
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestBuildAndApply(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		legacy string
	}{
		{name: "empty_rule_set", legacy: "lint: {}\n"},
		{name: "single_rule", legacy: "lint:\n  use: [SERVICE_SUFFIX]\n"},
		{name: "default_group", legacy: "lint:\n  use: [DEFAULT]\n  except: [SERVICE_SUFFIX]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", tt.legacy)
			require.NoError(t, os.Chmod(filepath.Join(root, "easyp.yaml"), 0o640))
			plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
			require.NoError(t, err)
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			assert.Len(t, entries, 1, "preview must not create backups or staging files")
			var policy v1.Policy
			require.NoError(t, yaml.Unmarshal(outputContent(t, plan, "easyp.yaml"), &policy))
			assert.Equal(t, "MINIMAL", policy.Linters.Default)
			require.NotNil(t, policy.Linters.AllowCommentIgnores)
			assert.False(t, *policy.Linters.AllowCommentIgnores)
			cfg, err := policy.LintConfig()
			require.NoError(t, err)
			actual, _, err := rules.New(cfg)
			require.NoError(t, err)
			legacy, _, err := parseLegacy([]byte(tt.legacy))
			require.NoError(t, err)
			expected, _, err := rules.New(legacy.Lint)
			require.NoError(t, err)
			assert.ElementsMatch(t, ruleNames(expected), ruleNames(actual))
			require.NoError(t, plan.Apply())
			backup, err := os.ReadFile(filepath.Join(root, "easyp.yaml.v0.bak"))
			require.NoError(t, err)
			assert.Equal(t, tt.legacy, string(backup))
			for _, name := range []string{"easyp.yaml", "easyp.yaml.v0.bak"} {
				info, err := os.Stat(filepath.Join(root, name))
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
			}
			second, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true})
			require.NoError(t, err)
			assert.True(t, second.AlreadyV1())
			require.NoError(t, second.Apply())
			assert.Equal(t, tt.legacy, string(mustRead(t, root, "easyp.yaml.v0.bak")))
		})
	}
}

func TestSettingsTransfer(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := `lint:
  use: [MINIMAL, BASIC, DEFAULT]
  except: [PACKAGE_DIRECTORY_MATCH]
  ignore: [generated]
  ignore_only:
    DEFAULT: [old]
  enum_zero_value_suffix: ZERO
  service_suffix: API
  allow_comment_ignores: true
breaking:
  ignore: [old]
  use: [FILE]
  against_git_ref: release
generate:
  inputs:
    - directory: {root: proto, path: .}
  plugins:
    - command: [./generator, "${TOKEN}"]
      out: ../generated
      opts: {paths: source_relative, token: "${TOKEN}", repeated: [a, b]}
      with_imports: true
    - remote: registry.example.com/go:v1.2.3
      out: gen
  managed:
    enabled: true
    disable:
      - module: example.com/dependency
        file_option: go_package
    override:
      - file_option: go_package_prefix
        value: "${GO_PREFIX}"
        path: foo
`
	writeFixture(t, root, "easyp.yaml", legacy)
	writeFixture(t, root, "proto/foo/a.proto", "syntax = \"proto3\";\n")
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
	require.NoError(t, err)
	var policy v1.Policy
	require.NoError(t, yaml.Unmarshal(outputContent(t, plan, "easyp.yaml"), &policy))
	assert.Equal(t, "ZERO", policy.LinterSettings["ENUM_ZERO_VALUE_SUFFIX"]["suffix"])
	assert.Equal(t, "API", policy.LinterSettings["SERVICE_SUFFIX"]["suffix"])
	assert.True(t, *policy.Linters.AllowCommentIgnores)
	assert.Equal(t, "git:release", policy.Breaking.Baseline)
	assert.Equal(t, []string{"FILE"}, policy.Breaking.Categories)
	assert.Equal(t, []string{"old"}, policy.Breaking.Ignore)
	assert.Contains(t, policy.Issues.ExcludeRules, v1.IssueExcludeRule{Path: "generated"})
	var gen v1.Generate
	require.NoError(t, yaml.Unmarshal(outputContent(t, plan, "easyp.gen.yaml"), &gen))
	require.Len(t, gen.Plugins, 2)
	assert.Equal(t, []string{"./generator", "${TOKEN}"}, gen.Plugins[0].Command)
	assert.Equal(t, "../generated", gen.Plugins[0].Out)
	assert.Equal(t, []string{"${TOKEN}"}, gen.Plugins[0].Opts["token"])
	assert.Equal(t, []string{"a", "b"}, gen.Plugins[0].Opts["repeated"])
	assert.True(t, gen.Plugins[0].WithImports)
	assert.Equal(t, "registry.example.com/go", gen.Plugins[1].Remote)
	assert.Equal(t, "v1.2.3", gen.Plugins[1].Version)
	assert.Equal(t, "${GO_PREFIX}", gen.Generate.Managed.Override[0].Value)
	assert.Equal(t, "foo", gen.Generate.Managed.Override[0].Path)
	assert.Equal(t, []string{"example.com/acme/api"}, gen.Generate.Modules)
	module, err := v1.ParseModule(bytes.NewReader(outputContent(t, plan, "protobuf.mod")))
	require.NoError(t, err)
	assert.Equal(t, []string{"proto"}, module.Roots)
}

func TestRejectedInputsDoNotWrite(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, input, want string }{
		{name: "multi_document", input: "lint: {}\n---\nlint: {}\n", want: "one YAML document"},
		{name: "null", input: "null\n", want: "null"},
		{name: "nested_null", input: "generate: {plugins: null}\n", want: "null"},
		{name: "empty", input: "", want: "empty"},
		{name: "alias", input: "lint: &lint {}\nbreaking: *lint\n", want: "alias"},
		{name: "duplicate", input: "lint: {}\nlint: {}\n", want: "duplicate"},
		{name: "nested_boundary", input: "generate:\n  inputs: [{directory: {root: ., path: selected}}]\n", want: "scope"},
		{name: "external", input: "generate:\n  inputs: [{directory: {root: ../external, path: .}}]\n", want: "manual"},
		{name: "placeholder", input: "generate:\n  inputs: [{directory: '${ROOT}'}]\n", want: "placeholder"},
		{name: "git_slice", input: "generate:\n  inputs: [{git_repo: {url: example.com/acme/deps, sub_directory: api}}]\n", want: "manual"},
		{name: "unsupported_ref", input: "deps: [example.com/acme/deps@main]\n", want: "full Git commit"},
		{name: "unpinned_remote", input: "generate:\n  inputs: [{directory: .}]\n  plugins: [{remote: registry/go:latest, out: gen}]\n", want: "semantic version"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			writeFixture(t, root, "easyp.yaml", tt.input)
			writeFixture(t, root, "selected/a.proto", "syntax = \"proto3\";\n")
			writeFixture(t, root, "other.proto", "syntax = \"proto3\";\n")
			if tt.name == "nested_boundary" {
				writeFixture(t, root, "selected/protobuf.mod", "module example.test/nested\n")
			}
			_, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
			require.ErrorContains(t, err, tt.want)
			assert.Equal(t, tt.input, string(mustRead(t, root, "easyp.yaml")))
			for _, name := range []string{"protobuf.mod", "easyp.gen.yaml", "easyp.yaml.v0.bak"} {
				_, err := os.Stat(filepath.Join(root, name))
				assert.ErrorIs(t, err, os.ErrNotExist)
			}
		})
	}
}

func TestPlanConflictsAndChanges(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, existing, changed string }{
		{name: "output_conflict", existing: "easyp.gen.yaml"},
		{name: "backup_conflict", existing: "easyp.yaml.v0.bak"},
		{name: "native_manifest", existing: "protobuf.mod"},
		{name: "source_changed", changed: "easyp.yaml"},
		{name: "output_created", changed: "easyp.gen.yaml"},
		{name: "proto_added", changed: "new.proto"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			legacy := "generate:\n  inputs: [{directory: .}]\n"
			writeFixture(t, root, "easyp.yaml", legacy)
			if tt.existing != "" {
				writeFixture(t, root, tt.existing, "module example.com/other/api\n")
			}
			plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
			if tt.existing != "" {
				require.Error(t, err)
				assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml")))
				return
			}
			require.NoError(t, err)
			writeFixture(t, root, tt.changed, "changed\n")
			require.Error(t, plan.Apply())
			_, err = os.Stat(filepath.Join(root, "protobuf.mod"))
			assert.ErrorIs(t, err, os.ErrNotExist)
		})
	}
}

func TestProtoAddedDuringStaging(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	legacy := "generate:\n  inputs: [{directory: .}]\n"
	writeFixture(t, root, "easyp.yaml", legacy)
	plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
	require.NoError(t, err)
	originalStage := plan.tx.stage
	plan.tx.stage = func(fsroot *os.Root, name string, content []byte, mode os.FileMode) error {
		writeFixture(t, root, "added.proto", "syntax = \"proto3\";\n")
		return originalStage(fsroot, name, content, mode)
	}
	require.ErrorContains(t, plan.Apply(), "source selection changed")
	assert.Equal(t, legacy, string(mustRead(t, root, "easyp.yaml")))
	_, err = os.Stat(filepath.Join(root, "protobuf.mod"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	entries, err := filepath.Glob(filepath.Join(root, ".easyp-migrate-*"))
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestManagedModuleSelectorSafety(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, selector     string
		enabled, wantError bool
	}{
		{name: "new_local_identity", selector: "example.com/acme/api", enabled: true, wantError: true},
		{name: "unknown_selector", selector: "${MODULE}", enabled: true, wantError: true},
		{name: "dependency_selector", selector: "example.com/acme/dep", enabled: true},
		{name: "disabled_mode", selector: "example.com/acme/api", enabled: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			enabled := "false"
			if tt.enabled {
				enabled = "true"
			}
			legacy := "generate:\n  inputs: [{directory: .}]\n  managed:\n    enabled: " + enabled + "\n    override:\n      - {module: '" + tt.selector + "', file_option: go_package_prefix, value: api}\n"
			writeFixture(t, root, "easyp.yaml", legacy)
			_, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api"})
			if tt.wantError {
				require.ErrorContains(t, err, "managed module selector")
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNativeNoOpValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, policy, gen, manifest, lock string
		wantError                         bool
	}{
		{name: "valid_option_list", gen: "version: v1\nplugins: [{name: go, out: gen, opts: [paths=source_relative]}]\n"},
		{name: "bad_plugin", gen: "version: v1\nplugins: [{name: go}]\n", wantError: true},
		{name: "unknown_preset", policy: "version: v1\nlinters: {default: MADE_UP}\n", wantError: true},
		{name: "missing_lock", manifest: "module example.com/acme/api\nrequire example.com/acme/dep v2.0.0\n", wantError: true},
		{name: "mismatched_lock", manifest: "module example.com/acme/api\nrequire example.com/acme/dep v2.0.0\n", lock: "version: 1\nmodules:\n  - {source: example.com/acme/dep, version: v1.0.0, commit: '" + testCommit + "', hash: '" + testHash + "'}\n", wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if tt.policy == "" {
				tt.policy = "version: v1\n"
			}
			if tt.gen == "" {
				tt.gen = "version: v1\nplugins: []\n"
			}
			if tt.manifest == "" {
				tt.manifest = "module example.com/acme/api\n"
			}
			writeFixture(t, root, "easyp.yaml", tt.policy)
			writeFixture(t, root, "easyp.gen.yaml", tt.gen)
			writeFixture(t, root, "protobuf.mod", tt.manifest)
			if tt.lock != "" {
				writeFixture(t, root, "protobuf.lock", tt.lock)
			}
			plan, err := Build(context.Background(), Options{Dir: root, Module: "example.com/acme/api", ResolveLock: true})
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.True(t, plan.AlreadyV1())
			require.NoError(t, plan.Apply())
			assert.Equal(t, tt.gen, string(mustRead(t, root, "easyp.gen.yaml")))
		})
	}
}

func writeFixture(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

func mustRead(t *testing.T, root, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, name))
	require.NoError(t, err)
	return data
}

func outputContent(t *testing.T, plan *Plan, name string) []byte {
	t.Helper()
	for _, output := range plan.Outputs() {
		if output.Name == name {
			return output.Content
		}
	}
	t.Fatalf("missing candidate %s", name)
	return nil
}

func ruleNames(input []core.Rule) []string {
	var names []string
	for _, rule := range input {
		names = append(names, core.GetRuleName(rule))
	}
	return names
}
