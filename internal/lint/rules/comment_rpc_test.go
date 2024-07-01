package rules_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/easyp-tech/easyp/internal/lint"
	"github.com/easyp-tech/easyp/internal/lint/rules"
)

func TestCommentRPC_Name(t *testing.T) {
	t.Parallel()

	assert := require.New(t)

	const expName = "COMMENT_RPC"

	rule := rules.CommentRPC{}
	name := rule.Name()

	assert.Equal(expName, name)
}

func TestCommentRPC_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"auth_rpc_comment_is_empty": {
			fileName: invalidAuthProto,
			wantErr:  lint.ErrRPCCommentIsEmpty,
		},
		"auth_rpc_comment_is_not_empty": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.CommentRPC{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
