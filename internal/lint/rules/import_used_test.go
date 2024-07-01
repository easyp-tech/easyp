package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestImportUsed_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "IMPORT_USED"

	rule := rules.ImportUsed{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestImportUsed_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"all imports are used": {
			fileName: importUsed,
			wantErr:  nil,
		},
		"not used imports": {
			fileName: importNotUsed,
			wantErr:  lint.ErrImportIsNotUsed,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.ImportUsed{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
