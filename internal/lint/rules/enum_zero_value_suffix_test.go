package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestEnumZeroValueSuffix_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "ENUM_ZERO_VALUE_SUFFIX"

	rule := rules.EnumZeroValueSuffix{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestEnumZeroValueSuffix_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrEnumZeroValueSuffix,
		},
		"valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.EnumZeroValueSuffix{
				Suffix: "NONE",
			}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
