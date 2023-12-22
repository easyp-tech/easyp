package core

import (
	"errors"
)

var (
	ErrInvalidRule = errors.New("invalid rule")
)

// Comment errors.
var (
	ErrServiceCommentIsEmpty      = errors.New("service comment is empty")
	ErrRPCCommentIsEmpty          = errors.New("rpc comment is empty")
	ErrEnumCommentIsEmpty         = errors.New("enum comment is empty")
	ErrOneOfCommentIsEmpty        = errors.New("oneof comment is empty")
	ErrMessageCommentIsEmpty      = errors.New("message comment is empty")
	ErrMessageFieldCommentIsEmpty = errors.New("message field comment is empty")
	ErrEnumPascalCase             = errors.New("enum is not pascal case")
	ErrEnumValueUpperSnakeCase    = errors.New("enum value is not upper snake case")
)
