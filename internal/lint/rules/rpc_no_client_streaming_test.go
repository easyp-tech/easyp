package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestRPCNoClientStreaming_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "RPC_NO_CLIENT_STREAMING"

	rule := rules.RPCNoClientStreaming{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestRPCNoClientStreaming_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"invalid": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrRPCClientStreaming,
		},
		"valid": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.RPCNoClientStreaming{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
