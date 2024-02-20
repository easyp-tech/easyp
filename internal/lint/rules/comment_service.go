package rules

import (
	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentService)(nil)

// CommentService is a rule for checking service comments.
type CommentService struct{}

// Validate implements core.Rule.
func (c *CommentService) Validate(protoInfo lint.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if len(service.Comments) == 0 {
			res = append(res, buildError(service.Meta.Pos, service.ServiceName, lint.ErrServiceCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
