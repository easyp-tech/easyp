package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestOneofLowerSnakeCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "ONEOF_LOWER_SNAKE_CASE"

	rule := rules.OneofLowerSnakeCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestOneofLowerSnakeCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_oneof_lower_snake_case_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrOneofLowerSnakeCase,
		},
		"check_oneof_lower_snake_case_is_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.OneofLowerSnakeCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
