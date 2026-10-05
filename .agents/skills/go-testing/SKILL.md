---
name: go-testing
description: "Use when writing, reviewing, or debugging Go tests in the github.com/easyp-tech/easyp CLI repository, including table tests, test doubles, and CLI state isolation."
argument-hint: "Describe the test you are writing or reviewing"
---

# Go Testing — EasyP CLI

Follow [AGENTS.md](../../../AGENTS.md), the mandatory [agent rules](../../../.spec/agent-rules.md), and [.spec/TESTING.md](../../../.spec/TESTING.md). EasyP tests exercise a CLI, engines, configuration, generation, migration, and Git-backed dependencies.

## Table Tests and Assertions

- Give new slice-based cases a descriptive <code>name string</code> field first, followed by inputs, optional setup, and expected outputs. Existing rule tests also use named map keys.
- Prefer lowercase scenario names such as <code>missing_manifest</code>, <code>invalid_policy</code>, or <code>canceled_context</code>; underscores improve readability.
- Use <code>t.Parallel()</code> at the top level and inside isolated subtests. Apply the process-state exceptions below before adding either call.
- Construct mutable dependencies inside each subtest. A setup callback should receive that case's dependency rather than capture shared mutable state. Give filesystem cases separate <code>t.TempDir()</code> directories.
- Use Testify <code>require</code> for fatal preconditions and <code>assert</code> for independent value checks. Check errors and nil pointers before accessing results.
- Use <code>require.ErrorIs</code> for sentinel errors, <code>require.ErrorAs</code> for typed details, and <code>require.ErrorContains</code> for meaningful diagnostics without a sentinel. Use <code>require.NoError</code> on success; do not invent sentinels to simplify tests.

This complete example uses the real <code>Validate.Action</code> API and follows [validate_test.go](../../../internal/api/validate_test.go). It can live in an <code>internal/api</code> test file. Each case has its own path, flag set, CLI context, and writer; it does not run the shared CLI parser setup.

~~~go
package api

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"

	"github.com/easyp-tech/easyp/internal/flags"
)

func TestValidateActionReportExample(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contents string
		wantHead string
		wantErr  error
	}{
		{
			name:     "valid_policy",
			contents: "version: v1\n",
			wantHead: "VALID: true\n",
		},
		{
			name:     "invalid_policy",
			contents: "linters:\n  unknown: true\n",
			wantHead: "VALID: false\nERRORS:\n",
			wantErr:  ErrHasValidateIssue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "easyp.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tt.contents), 0o600))
			var output bytes.Buffer
			set := flag.NewFlagSet("validate-config", flag.ContinueOnError)
			set.String(flags.Config.Name, "", "")
			set.String(flags.Format.Name, "", "")
			require.NoError(t, set.Set(flags.Config.Name, path))
			require.NoError(t, set.Set(flags.Format.Name, flags.TextFormat))
			ctx := cli.NewContext(&cli.App{Writer: &output}, set, nil)
			ctx.Context = t.Context()

			err := (Validate{}).Action(ctx)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, bytes.HasPrefix(output.Bytes(), []byte(tt.wantHead)), output.String())
		})
	}
}
~~~

## Context, Cwd, and Environment Isolation

Tests that change process cwd, environment, or global CLI/context hooks must remain sequential or run the affected behavior in an isolated subprocess. These are legitimate exceptions to parallel execution.

- With <code>t.Chdir</code> or <code>t.Setenv</code>, neither the test nor any ancestor may call <code>t.Parallel()</code>. Keep its state-dependent subtests sequential too. Prefer these helpers for automatic restoration.
- A private <code>t.TempDir()</code> does not isolate <code>os.Chdir</code>, <code>os.Setenv</code>, <code>os.Args</code>, standard streams, or package-global hooks. Avoid manual process mutation; when unavoidable, restore it with cleanup and check restoration errors.
- For parallel process-level scenarios, configure each child with <code>exec.Cmd.Dir</code> and <code>exec.Cmd.Env</code>, without changing the parent process. Test paths that call <code>os.Exit</code> or default CLI exit handling in subprocesses.
- Ordinary per-test contexts can run in parallel. Use <code>t.Context()</code> or a fresh cancellable child and clean it up. Do not share a mutable <code>*cli.Context</code>, <code>cli.App</code>, metadata map, or cancellation control between cases.
- Prefer absolute paths passed to APIs when a test does not need to exercise cwd behavior. See [get_v1_test.go](../../../internal/api/get_v1_test.go) for sequential cwd/environment cases and [resolve_test.go](../../../internal/modules/resolve_test.go) for isolated cancellation cases.

## urfave/cli v2 Parallel Tests

A fresh <code>cli.App</code> alone is insufficient. The library shares <code>cli.HelpFlag</code>, and parsing mutates flag state. Project globals in [internal/flags/flags.go](../../../internal/flags/flags.go) and some command constructors also reuse flag pointers.

For independent parallel parser tests:

1. Create an app, command tree, flags, writers, and metadata per case. Allocate fresh flag values and destinations too; a shallow copy can still share <code>GenericFlag.Value</code>, slices, or pointers.
2. Set <code>HideHelp: true</code> on the app and each command/subcommand in the test tree when help is irrelevant. <code>HideHelpCommand</code> alone does not remove the shared help flag. Use <code>HideVersion: true</code> on the app to avoid the shared version flag.
3. Do not reassign <code>cli.HelpFlag</code>, <code>cli.VersionFlag</code>, exit functions, or other package globals in a parallel test. Keep tests of real help/global behavior sequential or subprocess-isolated.
4. To test just an action, use a private standard-library <code>flag.FlagSet</code> and <code>cli.NewContext</code>, as above. This does not test command registration or flag parsing; cover those separately where relevant.

Follow [breaking_baseline_test.go](../../../internal/api/breaking_baseline_test.go) for fresh flags and help isolation, and [migrate_interactive_test.go](../../../internal/api/migrate_interactive_test.go) for a command tree with help disabled. Inspect each constructor before assuming it returns independent flags.

## Test Doubles and Mockery

Small handwritten consumer test doubles are appropriate. Put them in the consuming package's <code>_test.go</code> file, or a shared <code>helpers_test.go</code> when several files use them. Use names such as <code>mockRule</code> or <code>fakeSource</code> that identify the contract and behavior. Add an interface assertion when useful.

This complete double implements the actual [core.Rule](../../../internal/core/dom.go) interface:

~~~go
package core_test

import "github.com/easyp-tech/easyp/internal/core"

type mockRule struct {
	issues []core.Issue
	err    error
}

var _ core.Rule = (*mockRule)(nil)

func (m *mockRule) Message() string {
	return "example rule"
}

func (m *mockRule) Validate(_ core.ProtoInfo) ([]core.Issue, error) {
	return m.issues, m.err
}
~~~

Create a fresh double and any mutable slices/maps for each parallel case. Update doubles when their consumer interfaces change.

Optional Mockery generation is also supported through <code>task mock</code> and <code>task mocks</code> in [Taskfile.yml](../../../Taskfile.yml). Check the selected target and actual interface owner before generation; use [.spec/TESTING.md](../../../.spec/TESTING.md) for current tooling limitations. <code>Rule</code> and <code>CurrentProjectGitWalker</code> are in <code>internal/core</code>; <code>Console</code> is in [internal/adapters/console](../../../internal/adapters/console/new.go). Regenerate generated doubles rather than hand-editing them. Handwritten doubles do not require generation.

## Files, Packages, and Cleanup

- Colocate tests as <code>&lt;file&gt;_test.go</code>; use <code>Test&lt;Function&gt;</code> or <code>Test&lt;Function&gt;_&lt;scenario&gt;</code>, following nearby names for existing suites.
- Use same-package tests for unexported behavior and external test packages for exported contracts. Both are established here: [core/generate_path_test.go](../../../internal/core/generate_path_test.go) uses <code>package core</code>; [rules/file_lower_snake_case_test.go](../../../internal/rules/file_lower_snake_case_test.go) uses <code>package rules_test</code>.
- Mark helpers with <code>t.Helper()</code>. Use <code>t.Cleanup()</code> for resources and check close errors, following [rules/init_test.go](../../../internal/rules/init_test.go).
- Read fixtures from <code>testdata</code> and write generated test artifacts under temporary directories. Local Git tests can use temporary repositories without a public remote.
- For behavior changes, run affected package tests with <code>-race -count=1</code>, subject to the task's execution constraints. See [.spec/TESTING.md](../../../.spec/TESTING.md) for Task targets; documentation-only repairs do not require a full Go suite.

## Quick Checklist

- [ ] Cases are named; inputs, setup, and expected outputs are clear.
- [ ] Parallel cases have private dependencies, files, contexts, flags, and writers.
- [ ] Cwd/environment/global-state cases are sequential or subprocess-isolated.
- [ ] Parallel CLI parsers isolate application flags and the library's shared help/version flags.
- [ ] Fatal preconditions use <code>require</code>; error assertions match the actual error contract.
- [ ] Package choice follows the tested API; doubles match actual interfaces.
- [ ] Cleanup handles errors; relevant behavior tests have been run within the authorized scope.
