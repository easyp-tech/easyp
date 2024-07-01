package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestDirectorySamePackage_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "DIRECTORY_SAME_PACKAGE"

	rule := rules.DirectorySamePackage{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestDirectorySamePackage_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames []string
		wantErr   error
	}{
		"check_directory_same_package_is_invalid": {
			fileNames: []string{invalidAuthProto, invalidAuthProto2},
			wantErr:   lint.ErrDirectorySamePackage,
		},
		"check_directory_same_package_is_valid": {
			fileNames: []string{validAuthProto, validAuthProto2},
			wantErr:   nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.DirectorySamePackage{}
			for _, fileName := range tc.fileNames {
				err := rule.Validate(protos[fileName])
				if err != nil {
					r.ErrorIs(errors.Join(err...), tc.wantErr)
				}
			}
		})
	}
}
