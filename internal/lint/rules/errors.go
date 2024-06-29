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

	errMapping = map[string]error{
		// Minimal
		DIRECTORY_SAME_PACKAGE:  ErrDirectorySamePackage,
		PACKAGE_DEFINED:         ErrPackageIsNotDefined,
		PACKAGE_DIRECTORY_MATCH: ErrPackageIsNotMatchedWithPath,
		PACKAGE_SAME_DIRECTORY:  ErrPackageSameDirectory,

		// Basic
		ENUM_FIRST_VALUE_ZERO:            ErrEnumFirstValueZero,
		ENUM_NO_ALLOW_ALIAS:              ErrEnumNoAllowAlias,
		ENUM_PASCAL_CASE:                 ErrEnumPascalCase,
		ENUM_VALUE_UPPER_SNAKE_CASE:      ErrEnumValueUpperSnakeCase,
		FIELD_LOWER_SNAKE_CASE:           ErrMessageFieldLowerSnakeCase,
		IMPORT_NO_PUBLIC:                 ErrImportIsPublic,
		IMPORT_NO_WEAK:                   ErrImportIsWeak,
		IMPORT_USED:                      ErrMessagePascalCase,
		MESSAGE_PASCAL_CASE:              ErrMessagePascalCase,
		ONEOF_LOWER_SNAKE_CASE:           ErrOneofLowerSnakeCase,
		PACKAGE_LOWER_SNAKE_CASE:         ErrPackageLowerSnakeCase,
		PACKAGE_SAME_CSHARP_NAMESPACE:    ErrPackageSameCSharpNamespace,
		PACKAGE_SAME_GO_PACKAGE:          ErrPackageSameGoPackage,
		PACKAGE_SAME_JAVA_MULTIPLE_FILES: ErrPackageSameJavaMultipleFiles,
		PACKAGE_SAME_JAVA_PACKAGE:        ErrPackageSameJavaPackage,
		PACKAGE_SAME_PHP_NAMESPACE:       ErrPackageSamePhpNamespace,
		PACKAGE_SAME_RUBY_PACKAGE:        ErrPackageSameRubyPackage,
		PACKAGE_SAME_SWIFT_PREFIX:        ErrPackageSameSwiftPrefix,
		RPC_PASCAL_CASE:                  ErrRpcPascalCase,
		SERVICE_PASCAL_CASE:              ErrServicePascalCase,

		// Default
		ENUM_VALUE_PREFIX:           ErrEnumValuePrefix,
		ENUM_ZERO_VALUE_SUFFIX:      ErrEnumZeroValueSuffix,
		FILE_LOWER_SNAKE_CASE:       ErrFileLowerSnakeCase,
		RPC_REQUEST_RESPONSE_UNIQUE: ErrRPCRequestResponseUnique,
		RPC_REQUEST_STANDARD_NAME:   ErrRPCRequestStandardName,
		RPC_RESPONSE_STANDARD_NAME:  ErrRPCResponseStandardName,
		PACKAGE_VERSION_SUFFIX:      ErrPackageVersionSuffix,
		PROTOVALIDATE:               ErrProtoValidate,
		SERVICE_SUFFIX:              ErrServiceSuffix,

		// Comments
		COMMENT_ENUM:       ErrEnumCommentIsEmpty,
		COMMENT_ENUM_VALUE: ErrEnumValueCommentIsEmpty,
		COMMENT_FIELD:      ErrMessageFieldLowerSnakeCase,
		COMMENT_MESSAGE:    ErrMessageCommentIsEmpty,
		COMMENT_ONEOF:      ErrOneOfCommentIsEmpty,
		COMMENT_RPC:        ErrRPCCommentIsEmpty,
		COMMENT_SERVICE:    ErrServiceCommentIsEmpty,

		// UNARY_RPC
		RPC_NO_CLIENT_STREAMING: ErrRPCClientStreaming,
		RPC_NO_SERVER_STREAMING: ErrRPCServerStreaming,

		// Uncategorized
		PACKAGE_NO_IMPORT_CYCLE: ErrPackageNoImportCycle,
	}
)
