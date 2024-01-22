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
	ErrPackageIsNotDefined        = errors.New("package is not defined")
	ErrRPCCommentIsEmpty          = errors.New("rpc comment is empty")
	ErrEnumCommentIsEmpty         = errors.New("enum comment is empty")
	ErrOneOfCommentIsEmpty        = errors.New("oneof comment is empty")
	ErrMessageCommentIsEmpty      = errors.New("message comment is empty")
	ErrMessageFieldCommentIsEmpty = errors.New("message field comment is empty")
	ErrEnumValueCommentIsEmpty    = errors.New("enum value comment is empty")
	ErrEnumPascalCase             = errors.New("enum is not pascal case")
	ErrEnumValueUpperSnakeCase    = errors.New("enum value is not upper snake case")
	ErrRpcPascalCase              = errors.New("rpc is not pascal case")
	ErrServicePascalCase          = errors.New("service is not pascal case")
	ErrEnumFirstValueZero         = errors.New("enum first value is not zero")
	ErrDirectorySamePackage       = errors.New("different proto files in the same directory should have the same package")
	ErrMessagePascalCase          = errors.New("message is not pascal case")
	ErrMessageFieldLowerSnakeCase = errors.New("message field is not lower snake case")
	ErrOneofLowerSnakeCase        = errors.New("oneof is not lower snake case")
	ErrPackageLowerSnakeCase      = errors.New("package is not lower snake case")
)
