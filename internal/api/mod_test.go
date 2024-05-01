package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getEasypPath_EnvNotSet(t *testing.T) {
	t.Setenv(envEasypPath, "")

	userHomeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	expectedResult := filepath.Join(userHomeDir, defaultEasypPath)

	result, err := getEasypPath()
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func Test_getEasypPath_EnvSet(t *testing.T) {
	expectedResult := "/tmp/test"
	t.Setenv(envEasypPath, expectedResult)

	result, err := getEasypPath()
	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}
