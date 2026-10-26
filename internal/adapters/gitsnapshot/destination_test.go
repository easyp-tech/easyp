package gitsnapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDestinationValidatorRetainsHostCollisionChecks(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(directory, "Case", "api"), 0o755))
	validator := NewDestinationValidator(directory)
	require.NoError(t, validator.Validate("Case/api/first.proto"))
	require.NoError(t, validator.Validate("Case/api/second.proto"))
	_, caseErr := os.Stat(filepath.Join(directory, "case"))
	err := validator.Validate("case/api/other.proto")
	if caseErr == nil {
		require.ErrorIs(t, err, ErrPathCollision)
	} else {
		require.NoError(t, err)
	}
	require.NoError(t, os.RemoveAll(filepath.Join(directory, "Case")))
	require.NoError(t, os.MkdirAll(filepath.Join(directory, "case", "api"), 0o755))
	validator.Reset()
	err = validator.Validate("Case/api/other.proto")
	if caseErr == nil {
		require.ErrorIs(t, err, ErrPathCollision)
	} else {
		require.NoError(t, err)
	}
}

func BenchmarkDestinationValidation(b *testing.B) {
	directory := b.TempDir()
	for i := range 1000 {
		require.NoError(b, os.MkdirAll(filepath.Join(directory, "google", fmt.Sprintf("api%04d", i), "v1"), 0o755))
	}
	for _, name := range []string{"stateless", "snapshot"} {
		b.Run(name, func(b *testing.B) {
			validator := NewDestinationValidator(directory)
			b.ReportAllocs()
			for i := range b.N {
				path := fmt.Sprintf("google/api%04d/v1/file.proto", i%1000)
				var err error
				if name == "stateless" {
					err = ValidateDestination(directory, path)
				} else {
					err = validator.Validate(path)
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
