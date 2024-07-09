package rules

import (
	"reflect"

	"github.com/easyp-tech/easyp/internal/lint"
)

var _ lint.Rule = (*CommentService)(nil)

// CommentService this rule checks that services have non-empty comments.
type CommentService struct{}

// Name implements lint.Rule.
func (c *CommentService) Name() string {
	return toUpperSnakeCase(reflect.TypeOf(c).Elem().Name())
}

// Message implements lint.Rule.
func (c *CommentService) Message() string {
	return "service comments must not be empty"
}

// Validate implements lint.Rule.
func (c *CommentService) Validate(protoInfo lint.ProtoInfo) ([]lint.Issue, error) {
	var res []lint.Issue

	for _, service := range protoInfo.Info.ProtoBody.Services {
		if len(service.Comments) == 0 {
			res = append(res, lint.BuildError(service.Meta.Pos, service.ServiceName, c.Message()))
		}
	}

	return res, nil
}
