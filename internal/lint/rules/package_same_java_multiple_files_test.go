package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestPackageSameJavaMultipleFiles_Message(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expMessage = "all files in the same package must have the same java_multiple_files option"

	rule := rules.PackageSameJavaMultipleFiles{}
	message := rule.Message()

	assert.Equal(expMessage, message)
}

func TestPackageSameJavaMultipleFiles_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileNames  []string
		wantIssues *lint.Issue
		wantErr    error
	}{
		"invalid": {
			fileNames: []string{invalidAuthProto5, invalidAuthProto6},
			wantIssues: &lint.Issue{
				Position: meta.Position{
					Filename: "",
					Offset:   0,
					Line:     0,
					Column:   0,
				},
				SourceName: "",
				Message:    "",
			},
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

			rule := rules.PackageSameJavaMultipleFiles{}
			var got []error
			for _, fileName := range tc.fileNames {
				_, err := rule.Validate(protos[fileName])
				got = append(got, err)
			}

			r.ErrorIs(errors.Join(got...), tc.wantErr)
		})
	}
}
