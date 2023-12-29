package rules_test

import (
	"errors"
	"testing"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestMessageFieldLowerSnakeCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_message_field_lower_snake_case_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  core.ErrMessageFieldLowerSnakeCase,
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

			rule := rules.MessageFieldLowerSnakeCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
