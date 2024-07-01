package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestEnumNoAllowAlias_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "ENUM_NO_ALLOW_ALIAS"

	rule := rules.EnumNoAllowAlias{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestEnumNoAllowAlias_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"enum_no_allow_alias_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrEnumNoAllowAlias,
		},
		"enum_no_allow_alias_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.EnumNoAllowAlias{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
