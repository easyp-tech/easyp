package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
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

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if len(service.Comments) == 0 {
			res = core.AppendIssue(res, c, service.Meta.Pos, service.ServiceName, service.Comments)
		}
	}

	return res, nil
}
