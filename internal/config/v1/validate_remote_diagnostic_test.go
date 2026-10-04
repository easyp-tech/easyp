package v1

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/config"
)

func TestValidateFilePrefersEmbeddedRemoteVersionDiagnostic(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), GenerateFile)
	require.NoError(t, os.WriteFile(path, []byte("version: v1\nplugins:\n  - remote: localhost:8080/go:v1.2.3\n    out: gen\n"), 0o644))

	issues, err := ValidateFile(path)

	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, "v1_validation", issues[0].Code)
	assert.Equal(t, config.SeverityError, issues[0].Severity)
	assert.Contains(t, issues[0].Message, "split remote into remote:")
	assert.Contains(t, issues[0].Message, `remote: "localhost:8080/go"`)
	assert.Contains(t, issues[0].Message, `version: "v1.2.3"`)
}
