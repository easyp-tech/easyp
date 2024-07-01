package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestFieldLowerSnakeCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "FIELD_LOWER_SNAKE_CASE"

	rule := rules.FieldLowerSnakeCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestMessageFieldLowerSnakeCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_message_field_lower_snake_case_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrMessageFieldLowerSnakeCase,
		},
		"check_message_field_lower_snake_case_is_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.FieldLowerSnakeCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
