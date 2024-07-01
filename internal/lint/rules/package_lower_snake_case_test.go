package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestPackageLowerSnakeCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "PACKAGE_LOWER_SNAKE_CASE"

	rule := rules.PackageLowerSnakeCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestPackageLowerSnakeCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_package_lower_snake_case_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrPackageLowerSnakeCase,
		},
		"check_package_lower_snake_case_is_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.PackageLowerSnakeCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
