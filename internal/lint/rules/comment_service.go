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
			res = AppendError(res, COMMENT_SERVICE, service.Meta.Pos, service.ServiceName, service.Comments)
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
