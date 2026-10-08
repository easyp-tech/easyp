package moduleconfig

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGitDependencyRepeatedMergeAliasesRemainBounded(t *testing.T) {
	t.Parallel()
	const fixtureEnvironment = "EASYP_TEST_REPEATED_MERGE_FIXTURE"
	if directory := os.Getenv(fixtureEnvironment); directory != "" {
		_, err := ReadGitDependency(directory, "example.test/dependency")
		require.ErrorContains(t, err, "excessive aliasing")
		return
	}

	directory := t.TempDir()
	var fixture strings.Builder
	fixture.WriteString("version: v1\nbase0: &base0 {x: harmless}\n")
	for level := range 28 {
		fmt.Fprintf(&fixture, "base%d: &base%d {<<: [*base%d, *base%d]}\n", level+1, level+1, level, level)
	}
	fixture.WriteString("<<: *base28\n")
	require.NoError(t, os.WriteFile(filepath.Join(directory, "buf.yaml"), []byte(fixture.String()), 0o644))

	// Run the hostile DAG in a killable child. A regression must fail promptly
	// instead of leaving an exponentially expanding goroutine in the test process.
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestGitDependencyRepeatedMergeAliasesRemainBounded$", "-test.count=1")
	command.Env = append(os.Environ(), fixtureEnvironment+"="+directory)
	output, err := command.CombinedOutput()
	require.NoError(t, ctx.Err(), "root metadata preflight exceeded its bounded test window")
	require.NoError(t, err, "%s", output)
}

func TestDependencyYAMLMergeFieldPrecedence(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, raw, want string
	}{
		{name: "direct field wins", raw: "<<: [{path: first}, {path: second}]\npath: direct\n", want: "direct"},
		{name: "first merge wins", raw: "<<: [{path: first}, {path: second}]\n", want: "first"},
		{name: "later merge supplies absent field", raw: "<<: [{other: ignored}, {path: second}]\n", want: "second"},
		{name: "absent merged field", raw: "<<: [{other: first}, {other: second}]\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var document yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(tt.raw), &document))
			value := dependencyYAMLField(document.Content[0], "path")
			if tt.want == "" {
				assert.Nil(t, value)
				return
			}
			require.NotNil(t, value)
			assert.Equal(t, tt.want, value.Value)
		})
	}
}

func TestGitDependencyMergeCycleRetainsDecoderCause(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	raw := "version: v1\nbase: &base {<<: *base}\n<<: *base\n"
	require.NoError(t, os.WriteFile(filepath.Join(directory, "buf.yaml"), []byte(raw), 0o644))
	_, err := ReadGitDependency(directory, "example.test/dependency")
	require.ErrorContains(t, err, "contains itself")
}
