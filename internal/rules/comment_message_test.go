package rules_test

import (
	"errors"
	"testing"

	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/rules"
)

func TestCommentMessage_Validate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		fileName string
		wantErr  error
	}{
		"auth_message_comment_is_empty": {
			fileName: invalidAuthProto,
			wantErr:  core.ErrMessageCommentIsEmpty,
		},
		"auth_message_comment_is_not_empty": {
			fileName: validAuthProto,
			wantErr:  nil,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, protos := start(t)

			rule := rules.CommentMessage{}
			err := rule.Validate(protos[tc.fileName])
			r.ErrorIs(errors.Join(err...), tc.wantErr)
		})
	}
}
