package rules_test

import (
	"errors"
	"testing"

	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestCheckNoLint(t *testing.T) {
	t.Parallel()
	t.Skip()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"no_lint": {
			fileName: noLint,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.MessagePascalCase{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
