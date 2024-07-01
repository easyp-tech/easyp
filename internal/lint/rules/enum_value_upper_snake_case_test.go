package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestEnumValueUpperSnakeCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "ENUM_VALUE_UPPER_SNAKE_CASE"

	rule := rules.EnumValueUpperSnakeCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestEnumValueUpperSnakeCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_enum_value_upper_snake_case_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrEnumValueUpperSnakeCase,
		},
		"check_enum_value_upper_snake_case_is_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.EnumValueUpperSnakeCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
