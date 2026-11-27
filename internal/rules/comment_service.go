package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
	"github.com/easyp-tech/easyp/internal/protoast"
)

var _ core.Rule = (*CommentService)(nil)

// CommentService this rule checks that services have non-empty comments.
type CommentService struct{}

// Message implements lint.Rule.
func (c *CommentService) Message() string {
	return "service comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentService) Validate(protoInfo core.ProtoInfo) ([]core.Issue, error) {
	var res []core.Issue
	protoInfo.File.Walk(protoast.Visitor{Service: func(service *protoast.Service) {
		if !leadingCommentDocuments(&service.Decl) {
			res = core.AppendIssue(res, c, service.Pos, service.Name)
		}
	}})

	return res, nil
}
