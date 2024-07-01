package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestEnumFirstValueZero_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "ENUM_FIRST_VALUE_ZERO"

	rule := rules.EnumFirstValueZero{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestEnumFirstValueZero_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_enum_first_value_zero_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrEnumFirstValueZero,
		},
		"check_enum_first_value_zero_is_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.EnumFirstValueZero{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
