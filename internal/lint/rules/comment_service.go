package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentService)(nil)

// CommentService this rule checks that services have non-empty comments.
type CommentService struct{}

// Message implements lint.Rule.
func (c *CommentService) Message() string {
	return "service comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentService) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if len(service.Comments) == 0 {
			res = append(res, lint.BuildError(c, service.Meta.Pos, service.ServiceName))
		}
	}

	return res, nil
}
