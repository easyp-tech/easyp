package rules

import (
	"fmt"

	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"

	"github.com/easyp-tech/easyp/internal/core"
)

var _ core.Rule = (*CommentService)(nil)

// CommentService is a rule for checking service comments.
type CommentService struct{}

// Validate implements Rule.
func (c *CommentService) Validate(svc *unordered.Proto) []error {
	var res []error

	for _, service := range svc.ProtoBody.Services {
		if len(service.Comments) == 0 {
			res = append(res, &Error{
				Err: fmt.Errorf("%d:%d %s: %w", service.Meta.Pos.Line, service.Meta.Pos.Column, service.ServiceName, core.ErrServiceCommentIsEmpty),
			})
		}
	}

	if len(res) == 0 {
		return nil
	}

	return res
}
