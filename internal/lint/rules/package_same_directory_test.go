package rules_test

import (
	"errors"
	"testing"

	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestPackageSameDirectory_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames []string
		wantErr   error
	}{
		"check_package_same_directory_is_invalid": {
			fileNames: []string{invalidAuthProto2, invalidAuthProto4},
			wantErr:   rules.ErrPackageSameDirectory,
		},
		"check_package_package_same_is_valid": {
			fileNames: []string{validAuthProto, validAuthProto2},
			wantErr:   nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.PackageSameDirectory{}
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
