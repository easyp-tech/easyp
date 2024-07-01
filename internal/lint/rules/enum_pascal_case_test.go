package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestEnumPascalCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "ENUM_PASCAL_CASE"

	rule := rules.EnumPascalCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestEnumPascalCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_enum_pascal_case_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrEnumPascalCase,
		},
		"check_enum_pascal_case_is_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.EnumPascalCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
