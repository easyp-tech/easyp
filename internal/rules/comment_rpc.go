package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*CommentRPC)(nil)

// CommentRPC this rule checks that RPCs have non-empty comments.
type CommentRPC struct{}

// Message implements lint.Rule.
func (c *CommentRPC) Message() string {
	return "rpc comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentRPC) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{RPC: func(_ *protoast.Service, rpc *protoast.RPC) {
		if !leadingCommentDocuments(&rpc.Decl) {
			res = core.AppendIssue(res, c, rpc.Pos, rpc.Name)
		}
	}})

	return res, nil
}
