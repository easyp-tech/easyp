package rules_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestError_Error(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err error
		exp string
	}{
		"success": {
			err: lint.ErrServiceSuffix,
			exp: "10:1:file_name: " + lint.ErrServiceSuffix.Error(),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert := require.New(t)

			err := rules.BuildError(meta.Position{
				Filename: "file_name",
				Offset:   0,
				Line:     10,
				Column:   1,
			}, "file_name", tc.err)

			assert.Equal(tc.exp, err.Error())
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err error
		exp error
	}{
		"success": {
			err: lint.ErrServiceSuffix,
			exp: lint.ErrServiceSuffix,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert := require.New(t)

			err := rules.BuildError(meta.Position{
				Filename: "file_name",
				Offset:   0,
				Line:     10,
				Column:   1,
			}, "file_name", tc.err)

			var unwrapped error
			if err.(interface{ Unwrap() error }).Unwrap() != nil {
				unwrapped = err.(interface{ Unwrap() error }).Unwrap()
			}

			assert.Equal(tc.exp, unwrapped)
		})
	}
}
