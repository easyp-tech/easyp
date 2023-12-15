package rules

import (
	"github.com/easyp-tech/easyp/internal/core"
)

var (
	Rules = map[string]core.Rule{
		"COMMENT_ENUM":    &CommentEnum{},
		"COMMENT_ONEOF":   &CommentOneOf{},
		"COMMENT_RPC":     &CommentRPC{},
		"COMMENT_SERVICE": &CommentService{},
	}
)
