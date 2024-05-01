package rules_test

import (
	"errors"
	"testing"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestPackageSameCSharpNamespace_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames []string
		wantErr   error
	}{
		"invalid": {
			fileNames: []string{invalidAuthProto5, invalidAuthProto6},
			wantErr:   lint.ErrPackageSameCSharpNamespace,
		},
		"valid": {
			fileNames: []string{validAuthProto, validAuthProto2},
			wantErr:   nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.PackageSameCSharpNamespace{}
			var got []error
			for _, fileName := range tc.fileNames {
				err := rule.Validate(protos[fileName])
				if len(err) > 0 {
					got = append(got, err...)
				}
			}

			r.ErrorIs(errors.Join(got...), tc.wantErr)
		})
	}
}
