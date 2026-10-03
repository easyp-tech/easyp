package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDoesNotPromptWithoutTerminal(t *testing.T) {
	t.Parallel()
	p := terminalInitializationPrompter{}
	_, err := p.Input(t.Context(), "module", "")
	require.ErrorContains(t, err, "noninteractive mode; pass --module")
	confirmed, err := p.Confirm(t.Context(), "overwrite", true)
	require.ErrorContains(t, err, "without terminal input and output")
	assert.False(t, confirmed)
}

func TestInitRejectsInjectedIdentityBeforeFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	err := initializeV1(t.Context(), root, "example.com/app\nrequire example.com/other", &initPrompter{})
	require.ErrorContains(t, err, "expected one module identity")
	assert.NoFileExists(t, root+"/protobuf.mod")
	assert.NoFileExists(t, root+"/easyp.yaml")
}
