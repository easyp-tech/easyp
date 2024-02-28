package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentService)(nil)

// CommentService this rule checks that services have non-empty comments.
type CommentService struct{}

// Validate implements lint.Rule.
func (c *CommentService) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if len(service.Comments) == 0 {
			res = append(res, BuildError(service.Meta.Pos, service.ServiceName, lint.ErrServiceCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
