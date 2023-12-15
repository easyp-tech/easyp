package core

import (
	"errors"
)

var (
	ErrInvalidRule = errors.New("invalid rule")
)

// Comment errors.
var (
	ErrServiceCommentIsEmpty = errors.New("service comment is empty")
	ErrRPCCommentIsEmpty     = errors.New("rpc comment is empty")
	ErrEnumCommentIsEmpty    = errors.New("enum comment is empty")
	ErrOneOfCommentIsEmpty   = errors.New("oneof comment is empty")
)
