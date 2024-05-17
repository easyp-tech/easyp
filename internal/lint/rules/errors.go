package rules

import (
	"errors"
)

var (
	ErrInvalidRule = errors.New("invalid rule")
)

// Comment errors.
var (
	// Minimal
	ErrDirectorySamePackage        = errors.New("different proto files in the same directory should have the same package")
	ErrPackageIsNotDefined         = errors.New("package is not defined")
	ErrPackageIsNotMatchedWithPath = errors.New("package is not matched with the path")
	ErrPackageSameDirectory        = errors.New("different proto files in the same package should be in the same directory")

	// Basic
	ErrEnumFirstValueZero           = errors.New("enum first value is not zero")
	ErrEnumNoAllowAlias             = errors.New("enum no allow alias")
	ErrEnumPascalCase               = errors.New("enum is not pascal case")
	ErrEnumValueUpperSnakeCase      = errors.New("enum value is not upper snake case")
	ErrMessageFieldLowerSnakeCase   = errors.New("message field is not lower snake case")
	ErrImportIsPublic               = errors.New("import is public")
	ErrImportIsWeak                 = errors.New("import is weak")
	ErrImportIsNotUsed              = errors.New("import is not used")
	ErrMessagePascalCase            = errors.New("message is not pascal case")
	ErrOneofLowerSnakeCase          = errors.New("oneof is not lower snake case")
	ErrPackageLowerSnakeCase        = errors.New("package is not lower snake case")
	ErrPackageSameCSharpNamespace   = errors.New("different proto files in the same package should have the same csharp_namespace")
	ErrPackageSameGoPackage         = errors.New("different proto files in the same package should have the same go_package")
	ErrPackageSameJavaMultipleFiles = errors.New("different proto files in the same package should have the same java_multiple_files")
	ErrPackageSameJavaPackage       = errors.New("different proto files in the same package should have the same java_package")
	ErrPackageSamePhpNamespace      = errors.New("different proto files in the same package should have the same php_namespace")
	ErrPackageSameRubyPackage       = errors.New("different proto files in the same package should have the same ruby_package")
	ErrPackageSameSwiftPrefix       = errors.New("different proto files in the same package should have the same swift_prefix")
	ErrRpcPascalCase                = errors.New("rpc is not pascal case")
	ErrServicePascalCase            = errors.New("service is not pascal case")

	// Default
	ErrEnumValuePrefix          = errors.New("enum value prefix is not valid")
	ErrEnumZeroValueSuffix      = errors.New("enum zero value suffix is not valid")
	ErrFileLowerSnakeCase       = errors.New("file is not lower snake case")
	ErrRPCRequestResponseUnique = errors.New("rpc request and response should be unique")
	ErrRPCRequestStandardName   = errors.New("rpc request should have suffix 'Request'")
	ErrRPCResponseStandardName  = errors.New("rpc response should have suffix 'Response'")
	ErrPackageVersionSuffix     = errors.New("package version suffix is not valid")
	ErrProtoValidate            = errors.New("validate proto") // TODO: This, rule, is not implemented yet
	ErrServiceSuffix            = errors.New("service name should have suffix")

	// Comments
	ErrEnumCommentIsEmpty         = errors.New("enum comment is empty")
	ErrEnumValueCommentIsEmpty    = errors.New("enum value comment is empty")
	ErrMessageFieldCommentIsEmpty = errors.New("message field comment is empty")
	ErrMessageCommentIsEmpty      = errors.New("message comment is empty")
	ErrOneOfCommentIsEmpty        = errors.New("oneof comment is empty")
	ErrRPCCommentIsEmpty          = errors.New("rpc comment is empty")
	ErrServiceCommentIsEmpty      = errors.New("service comment is empty")

	// UNARY_RPC
	ErrRPCClientStreaming = errors.New("rpc client streaming is forbidden")
	ErrRPCServerStreaming = errors.New("rpc server streaming is forbidden")

	// Uncategorized
	ErrPackageNoImportCycle = errors.New("package has import cycle") // TODO: This, rule, is not implemented yet
)
