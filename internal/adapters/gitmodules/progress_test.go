package gitmodules

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/logger"
)

func TestMigrationDiagnosticsSurviveHistoricalPinScope(t *testing.T) {
	t.Parallel()
	remote, commit := diagnosticRepository(t)
	var output bytes.Buffer
	log := logger.New(slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	cache := New(t.TempDir()).WithLogger(log).WithMigrationPins(nil).(*Cache)
	_, err := cache.FetchMigrationImports(t.Context(), remote, commit, "")
	require.NoError(t, err)
	data := output.String()
	for _, phase := range []string{"migrate dependency", "git", "cache lock", "snapshot", "source inspection", "legacy archive proof"} {
		assert.Contains(t, data, `"phase":"`+phase+`"`)
	}
	assert.Contains(t, data, remote)
	assert.Contains(t, data, commit)
	assert.Contains(t, data, `"regular_files":`)
	assert.Contains(t, data, `"elapsed":`)
	assert.Contains(t, data, `"status":"complete"`)
}
