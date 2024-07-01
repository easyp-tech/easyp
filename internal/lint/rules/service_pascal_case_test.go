package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestServicePascalCase_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "SERVICE_PASCAL_CASE"

	rule := rules.ServicePascalCase{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestServicePascalCase_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"check_service_pascal_case_is_invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrServicePascalCase,
		},
		"check_service_pascal_case_is_valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.ServicePascalCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
