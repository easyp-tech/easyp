package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentService)(nil)

// CommentService is a rule for checking service comments.
type CommentService struct{}

// Validate implements core.Rule.
func (c *CommentService) Validate(protoInfo core.ProtoInfo) []error {
	var res []error

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if len(service.Comments) == 0 {
			res = append(res, buildError(service.Meta.Pos, service.ServiceName, core.ErrServiceCommentIsEmpty))
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
