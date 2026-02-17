package core

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderInitConfig(t *testing.T) {
	data := InitTemplateData{
		LintGroups: []LintGroup{
			{
				Name:  "Minimal",
				Rules: []string{"DIRECTORY_SAME_PACKAGE", "PACKAGE_DEFINED"},
			},
			{
				Name:  "Basic",
				Rules: []string{"ENUM_FIRST_VALUE_ZERO", "FIELD_LOWER_SNAKE_CASE"},
			},
		},
		EnumZeroValueSuffix: "_NONE",
		ServiceSuffix:       "API",
		AgainstGitRef:       "master",
	}

	var buf bytes.Buffer
	err := renderInitConfig(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	t.Run("contains group comments", func(t *testing.T) {
		require.Contains(t, output, "# Minimal")
		require.Contains(t, output, "# Basic")
	})

	t.Run("contains all rules", func(t *testing.T) {
		require.Contains(t, output, "- DIRECTORY_SAME_PACKAGE")
		require.Contains(t, output, "- PACKAGE_DEFINED")
		require.Contains(t, output, "- ENUM_FIRST_VALUE_ZERO")
		require.Contains(t, output, "- FIELD_LOWER_SNAKE_CASE")
	})

	t.Run("contains enum_zero_value_suffix", func(t *testing.T) {
		require.Contains(t, output, "enum_zero_value_suffix: _NONE")
	})

	t.Run("contains service_suffix", func(t *testing.T) {
		require.Contains(t, output, "service_suffix: API")
	})

	t.Run("contains breaking section", func(t *testing.T) {
		require.Contains(t, output, "against_git_ref: master")
	})

	t.Run("contains documentation link", func(t *testing.T) {
		require.Contains(t, output, "https://easyp.tech")
	})

	t.Run("is valid yaml structure", func(t *testing.T) {
		// Verify basic YAML structure by checking indentation
		lines := strings.Split(output, "\n")
		foundLint := false
		foundBreaking := false
		for _, line := range lines {
			if strings.TrimSpace(line) == "lint:" {
				foundLint = true
			}
			if strings.TrimSpace(line) == "breaking:" {
				foundBreaking = true
			}
		}
		require.True(t, foundLint, "should contain lint: section")
		require.True(t, foundBreaking, "should contain breaking: section")
	})
}

func TestRenderInitConfig_EmptyGroups(t *testing.T) {
	data := InitTemplateData{
		LintGroups:          nil,
		EnumZeroValueSuffix: "_UNSPECIFIED",
		ServiceSuffix:       "Service",
		AgainstGitRef:       "develop",
	}

	var buf bytes.Buffer
	err := renderInitConfig(&buf, data)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "enum_zero_value_suffix: _UNSPECIFIED")
	require.Contains(t, output, "service_suffix: Service")
	require.Contains(t, output, "against_git_ref: develop")
}
