package core

import (
	"path/filepath"
	"strings"
	"unicode"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ============================================================================
// Types and Constants for Managed Mode
// ============================================================================

// FileOptionType represents the type of file option that can be managed.
type FileOptionType string

const (
	// Go options
	FileOptionGoPackage       FileOptionType = "go_package"
	FileOptionGoPackagePrefix FileOptionType = "go_package_prefix"

	// Java options
	FileOptionJavaPackage         FileOptionType = "java_package"
	FileOptionJavaPackagePrefix   FileOptionType = "java_package_prefix"
	FileOptionJavaPackageSuffix   FileOptionType = "java_package_suffix"
	FileOptionJavaMultipleFiles   FileOptionType = "java_multiple_files"
	FileOptionJavaOuterClassname  FileOptionType = "java_outer_classname"
	FileOptionJavaStringCheckUtf8 FileOptionType = "java_string_check_utf8"

	// C# options
	FileOptionCsharpNamespace       FileOptionType = "csharp_namespace"
	FileOptionCsharpNamespacePrefix FileOptionType = "csharp_namespace_prefix"

	// Ruby options
	FileOptionRubyPackage       FileOptionType = "ruby_package"
	FileOptionRubyPackageSuffix FileOptionType = "ruby_package_suffix"

	// PHP options
	FileOptionPhpNamespace               FileOptionType = "php_namespace"
	FileOptionPhpMetadataNamespace       FileOptionType = "php_metadata_namespace"
	FileOptionPhpMetadataNamespaceSuffix FileOptionType = "php_metadata_namespace_suffix"

	// Objective-C options
	FileOptionObjcClassPrefix FileOptionType = "objc_class_prefix"

	// Swift options
	FileOptionSwiftPrefix FileOptionType = "swift_prefix"

	// Optimization options
	FileOptionOptimizeFor FileOptionType = "optimize_for"

	// C++ options
	FileOptionCcEnableArenas FileOptionType = "cc_enable_arenas"
)

// FieldOptionType represents the type of field option that can be managed.
type FieldOptionType string

const (
	// JavaScript type option for int64/uint64 fields
	FieldOptionJsType FieldOptionType = "jstype"
)

// OptimizeMode represents the optimization mode for generated code.
type OptimizeMode string

const (
	OptimizeModeSpeed       OptimizeMode = "SPEED"
	OptimizeModeCodeSize    OptimizeMode = "CODE_SIZE"
	OptimizeModeLiteRuntime OptimizeMode = "LITE_RUNTIME"
)

// JSType represents the JavaScript type for int64/uint64 fields.
type JSType string

const (
	JSTypeNormal JSType = "JS_NORMAL"
	JSTypeString JSType = "JS_STRING"
	JSTypeNumber JSType = "JS_NUMBER"
)

// ManagedDisableRule defines a rule to disable managed mode for specific conditions.
type ManagedDisableRule struct {
	// Module disables managed mode for all files in the specified module.
	Module string
	// Path disables managed mode for files matching the specified path (directory or file).
	Path string
	// FileOption disables a specific file option from being modified.
	FileOption FileOptionType
	// FieldOption disables a specific field option from being modified.
	FieldOption FieldOptionType
	// Field disables a specific field (fully qualified name: package.Message.field).
	Field string
}

// ManagedOverrideRule defines a rule to override file or field options.
type ManagedOverrideRule struct {
	// FileOption specifies which file option to override.
	FileOption FileOptionType
	// FieldOption specifies which field option to override.
	FieldOption FieldOptionType
	// Value is the value to set for the option.
	Value any
	// Module applies this override only to files in the specified module.
	Module string
	// Path applies this override only to files matching the specified path.
	Path string
	// Field applies this override only to the specified field (fully qualified name).
	Field string
}

// ManagedModeConfig is the runtime configuration for managed mode.
type ManagedModeConfig struct {
	// Enabled activates managed mode.
	Enabled bool
	// Disable contains rules to disable managed mode for specific conditions.
	Disable []ManagedDisableRule
	// Override contains rules to override file and field options.
	Override []ManagedOverrideRule
}

// ============================================================================
// Config Methods for Rule Matching
// ============================================================================

// IsFileOptionDisabled checks if a file option is disabled for the given file.
func (c *ManagedModeConfig) IsFileOptionDisabled(filePath, module string, option FileOptionType) bool {
	for _, rule := range c.Disable {
		if rule.matchesFileOption(filePath, module, option) {
			return true
		}
	}
	return false
}

// IsFieldOptionDisabled checks if a field option is disabled for the given field.
func (c *ManagedModeConfig) IsFieldOptionDisabled(filePath, module string, option FieldOptionType, fieldName string) bool {
	for _, rule := range c.Disable {
		if rule.matchesFieldOption(filePath, module, option, fieldName) {
			return true
		}
	}
	return false
}

// GetFileOptionOverride returns the override value for a file option, or nil if not overridden.
// If multiple overrides match, the last matching rule wins (buf behavior).
func (c *ManagedModeConfig) GetFileOptionOverride(filePath, module string, option FileOptionType) any {
	var result any
	for _, rule := range c.Override {
		if rule.FileOption == option && rule.matchesFileContext(filePath, module) {
			result = rule.Value
		}
	}
	return result
}

// GetFieldOptionOverride returns the override value for a field option, or nil if not overridden.
// If multiple overrides match, the last matching rule wins (buf behavior).
func (c *ManagedModeConfig) GetFieldOptionOverride(filePath, module string, option FieldOptionType, fieldName string) any {
	var result any
	for _, rule := range c.Override {
		if rule.FieldOption == option && rule.matchesFieldContext(filePath, module, fieldName) {
			result = rule.Value
		}
	}
	return result
}

// ============================================================================
// Rule Matching Methods
// ============================================================================

// matchesFileOption checks if a disable rule matches the given file option context.
func (r *ManagedDisableRule) matchesFileOption(filePath, module string, fileOption FileOptionType) bool {
	// If file option is specified, it must match
	if r.FileOption != "" && r.FileOption != fileOption {
		return false
	}

	// Field option rules don't apply to file options
	if r.FieldOption != "" {
		return false
	}

	return r.matchesContext(filePath, module)
}

// matchesFieldOption checks if a disable rule matches the given field option context.
func (r *ManagedDisableRule) matchesFieldOption(filePath, module string, fieldOption FieldOptionType, fieldName string) bool {
	// If field option is specified, it must match
	if r.FieldOption != "" && r.FieldOption != fieldOption {
		return false
	}

	// File option rules don't apply to field options
	if r.FileOption != "" {
		return false
	}

	// If field is specified, it must match
	if r.Field != "" && r.Field != fieldName {
		return false
	}

	return r.matchesContext(filePath, module)
}

// matchesContext checks if module and path filters match.
// Uses exact module comparison to avoid false positives.
func (r *ManagedDisableRule) matchesContext(filePath, module string) bool {
	// Check module match - use exact comparison to avoid false positives
	// e.g., "googleapis" should NOT match "github.com/mycompany/super-googleapis-tools"
	if r.Module != "" && r.Module != module {
		return false
	}

	// Check path match - use prefix matching for paths
	if r.Path != "" && !strings.HasPrefix(filePath, r.Path) {
		return false
	}

	return true
}

// matchesFileContext checks if an override rule matches the given file context.
// Uses exact module comparison to avoid false positives.
func (r *ManagedOverrideRule) matchesFileContext(filePath, module string) bool {
	// Check module match - use exact comparison to avoid false positives
	if r.Module != "" && r.Module != module {
		return false
	}

	// Check path match - use prefix matching for paths
	if r.Path != "" && !strings.HasPrefix(filePath, r.Path) {
		return false
	}

	return true
}

// matchesFieldContext checks if an override rule matches the given field context.
func (r *ManagedOverrideRule) matchesFieldContext(filePath, module, fieldName string) bool {
	// Check field match
	if r.Field != "" && r.Field != fieldName {
		return false
	}

	return r.matchesFileContext(filePath, module)
}

// ============================================================================
// Option Handlers
// ============================================================================

// FileOptionHandler defines how to apply a file option.
type FileOptionHandler struct {
	// Option is the option type identifier.
	Option FileOptionType
	// HasDefault indicates if this option has a default value when managed mode is enabled.
	// Only options listed in buf documentation as having defaults should have this set to true.
	HasDefault bool
	// Apply applies the option value to the file descriptor.
	Apply func(fd *descriptorpb.FileDescriptorProto, value any, pkg string)
	// Default returns the default value for this option (if HasDefault is true).
	// This is called only when HasDefault is true and no override is specified.
	Default func(fd *descriptorpb.FileDescriptorProto, pkg string) any
	// AffectsOption specifies which base option this handler affects.
	// For example, go_package_prefix affects go_package.
	AffectsOption FileOptionType
}

// FieldOptionHandler defines how to apply a field option.
type FieldOptionHandler struct {
	// Option is the option type identifier.
	Option FieldOptionType
	// HasDefault indicates if this option has a default value when managed mode is enabled.
	HasDefault bool
	// Default returns the default value for this option (if HasDefault is true).
	Default func() any
	// AppliesToType checks if this option applies to the given field type.
	AppliesToType func(t descriptorpb.FieldDescriptorProto_Type) bool
	// Apply applies the option value to the field.
	Apply func(field *descriptorpb.FieldDescriptorProto, value any)
}

// fileOptionHandlers is the registry of all supported file options.
// According to buf documentation, the following options have defaults when managed mode is enabled:
// - java_multiple_files: true
// - java_outer_classname: PascalCase of file name + "Proto"
// - java_package_prefix: "com" (sets java_package to com.<proto_package>)
// - csharp_namespace: PascalCase of package with "." separator
// - ruby_package: PascalCase of package with "::" separator
// - php_namespace: PascalCase of package with "\" separator
// - objc_class_prefix: First letters of each package part, minimum 3 chars
// - cc_enable_arenas: true
//
// Options WITHOUT defaults (only applied when explicitly overridden):
// - go_package / go_package_prefix
// - java_package / java_package_suffix
// - java_string_check_utf8
// - optimize_for
// - swift_prefix
// - csharp_namespace_prefix
// - ruby_package_suffix
// - php_metadata_namespace / php_metadata_namespace_suffix
var fileOptionHandlers = []FileOptionHandler{
	// Go options - NO defaults in buf
	{
		Option:     FileOptionGoPackage,
		HasDefault: false,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.GoPackage = proto.String(v)
			}
		},
	},
	{
		Option:        FileOptionGoPackagePrefix,
		HasDefault:    false,
		AffectsOption: FileOptionGoPackage,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, pkg string) {
			if prefix, ok := value.(string); ok {
				// go_package_prefix sets go_package to <prefix>/<proto_package with dots replaced by slashes>
				pkgPath := strings.ReplaceAll(pkg, ".", "/")
				fd.Options.GoPackage = proto.String(prefix + "/" + pkgPath)
			}
		},
	},

	// Java options
	{
		Option:     FileOptionJavaPackage,
		HasDefault: false, // No default - only java_package_prefix has default
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.JavaPackage = proto.String(v)
			}
		},
	},
	{
		Option:        FileOptionJavaPackagePrefix,
		HasDefault:    true, // Default is "com"
		AffectsOption: FileOptionJavaPackage,
		Default: func(_ *descriptorpb.FileDescriptorProto, _ string) any {
			return "com" // buf default prefix
		},
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, pkg string) {
			// java_package_prefix sets java_package to <prefix>.<proto_package>
			if prefix, ok := value.(string); ok {
				fd.Options.JavaPackage = proto.String(prefix + "." + pkg)
			}
		},
	},
	{
		Option:        FileOptionJavaPackageSuffix,
		HasDefault:    false, // No default
		AffectsOption: FileOptionJavaPackage,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, pkg string) {
			// java_package_suffix sets java_package to <proto_package>.<suffix>
			if suffix, ok := value.(string); ok {
				fd.Options.JavaPackage = proto.String(pkg + "." + suffix)
			}
		},
	},
	{
		Option:     FileOptionJavaMultipleFiles,
		HasDefault: true,
		Default:    func(_ *descriptorpb.FileDescriptorProto, _ string) any { return true },
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(bool); ok {
				fd.Options.JavaMultipleFiles = proto.Bool(v)
			}
		},
	},
	{
		Option:     FileOptionJavaOuterClassname,
		HasDefault: true,
		Default: func(fd *descriptorpb.FileDescriptorProto, _ string) any {
			// Default: PascalCase of file name + "Proto"
			base := filepath.Base(fd.GetName())
			name := strings.TrimSuffix(base, filepath.Ext(base))
			return toPascalCase(name) + "Proto"
		},
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.JavaOuterClassname = proto.String(v)
			}
		},
	},
	{
		Option:     FileOptionJavaStringCheckUtf8,
		HasDefault: false, // No default
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(bool); ok {
				fd.Options.JavaStringCheckUtf8 = proto.Bool(v)
			}
		},
	},

	// C# options
	{
		Option:     FileOptionCsharpNamespace,
		HasDefault: true,
		Default: func(_ *descriptorpb.FileDescriptorProto, pkg string) any {
			// Default: PascalCase of package with "." separator
			// e.g., acme.weather.v1 -> Acme.Weather.V1
			return toPascalCaseWithSeparator(pkg, ".")
		},
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.CsharpNamespace = proto.String(v)
			}
		},
	},
	{
		Option:        FileOptionCsharpNamespacePrefix,
		HasDefault:    false, // No default
		AffectsOption: FileOptionCsharpNamespace,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, pkg string) {
			// csharp_namespace_prefix sets csharp_namespace to <prefix>.<PascalCase(package)>
			if prefix, ok := value.(string); ok {
				fd.Options.CsharpNamespace = proto.String(prefix + "." + toPascalCaseWithSeparator(pkg, "."))
			}
		},
	},

	// Ruby options
	{
		Option:     FileOptionRubyPackage,
		HasDefault: true,
		Default: func(_ *descriptorpb.FileDescriptorProto, pkg string) any {
			// Default: PascalCase of package with "::" separator
			// e.g., acme.weather.v1 -> Acme::Weather::V1
			return toPascalCaseWithSeparator(pkg, "::")
		},
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.RubyPackage = proto.String(v)
			}
		},
	},
	{
		Option:        FileOptionRubyPackageSuffix,
		HasDefault:    false, // No default
		AffectsOption: FileOptionRubyPackage,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, pkg string) {
			// ruby_package_suffix sets ruby_package to <PascalCase(package)>::<suffix>
			if suffix, ok := value.(string); ok {
				fd.Options.RubyPackage = proto.String(toPascalCaseWithSeparator(pkg, "::") + "::" + suffix)
			}
		},
	},

	// PHP options
	{
		Option:     FileOptionPhpNamespace,
		HasDefault: true,
		Default: func(_ *descriptorpb.FileDescriptorProto, pkg string) any {
			// Default: PascalCase of package with "\" separator
			// e.g., acme.weather.v1 -> Acme\Weather\V1
			return toPascalCaseWithSeparator(pkg, `\`)
		},
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.PhpNamespace = proto.String(v)
			}
		},
	},
	{
		Option:     FileOptionPhpMetadataNamespace,
		HasDefault: false, // No default
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.PhpMetadataNamespace = proto.String(v)
			}
		},
	},
	{
		Option:        FileOptionPhpMetadataNamespaceSuffix,
		HasDefault:    false, // No default
		AffectsOption: FileOptionPhpMetadataNamespace,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, pkg string) {
			// php_metadata_namespace_suffix sets php_metadata_namespace to <PascalCase(package)>\<suffix>
			if suffix, ok := value.(string); ok {
				fd.Options.PhpMetadataNamespace = proto.String(toPascalCaseWithSeparator(pkg, `\`) + `\` + suffix)
			}
		},
	},

	// Objective-C options
	{
		Option:     FileOptionObjcClassPrefix,
		HasDefault: true,
		Default: func(_ *descriptorpb.FileDescriptorProto, pkg string) any {
			// Default: First uppercase letter of each package part, minimum 3 chars
			// e.g., acme.weather.v1 -> AWV
			// If less than 3 chars, pad with 'X'
			// "GPB" is reserved by Google Protobuf, changed to "GPX"
			return generateObjcClassPrefix(pkg)
		},
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.ObjcClassPrefix = proto.String(v)
			}
		},
	},

	// Swift options - NO default in buf
	{
		Option:     FileOptionSwiftPrefix,
		HasDefault: false,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				fd.Options.SwiftPrefix = proto.String(v)
			}
		},
	},

	// Optimization options - NO default in buf
	{
		Option:     FileOptionOptimizeFor,
		HasDefault: false,
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(string); ok {
				switch OptimizeMode(v) {
				case OptimizeModeSpeed:
					fd.Options.OptimizeFor = descriptorpb.FileOptions_SPEED.Enum()
				case OptimizeModeCodeSize:
					fd.Options.OptimizeFor = descriptorpb.FileOptions_CODE_SIZE.Enum()
				case OptimizeModeLiteRuntime:
					fd.Options.OptimizeFor = descriptorpb.FileOptions_LITE_RUNTIME.Enum()
				}
			}
		},
	},

	// C++ options
	{
		Option:     FileOptionCcEnableArenas,
		HasDefault: true,
		Default:    func(_ *descriptorpb.FileDescriptorProto, _ string) any { return true },
		Apply: func(fd *descriptorpb.FileDescriptorProto, value any, _ string) {
			if v, ok := value.(bool); ok {
				fd.Options.CcEnableArenas = proto.Bool(v)
			}
		},
	},
}

// fieldOptionHandlers is the registry of all supported field options.
// According to buf documentation:
// - jstype: NO default - only applied when explicitly overridden
var fieldOptionHandlers = []FieldOptionHandler{
	{
		Option:     FieldOptionJsType,
		HasDefault: false, // No default in buf
		AppliesToType: func(t descriptorpb.FieldDescriptorProto_Type) bool {
			// jstype only applies to 64-bit integer types
			switch t {
			case descriptorpb.FieldDescriptorProto_TYPE_INT64,
				descriptorpb.FieldDescriptorProto_TYPE_UINT64,
				descriptorpb.FieldDescriptorProto_TYPE_SINT64,
				descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
				descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
				return true
			}
			return false
		},
		Apply: func(field *descriptorpb.FieldDescriptorProto, value any) {
			if v, ok := value.(string); ok {
				if field.Options == nil {
					field.Options = &descriptorpb.FieldOptions{}
				}
				switch JSType(v) {
				case JSTypeNormal:
					field.Options.Jstype = descriptorpb.FieldOptions_JS_NORMAL.Enum()
				case JSTypeString:
					field.Options.Jstype = descriptorpb.FieldOptions_JS_STRING.Enum()
				case JSTypeNumber:
					field.Options.Jstype = descriptorpb.FieldOptions_JS_NUMBER.Enum()
				}
			}
		},
	},
}

// fileOptionHandlerMap provides quick lookup by option type.
var fileOptionHandlerMap = func() map[FileOptionType]*FileOptionHandler {
	m := make(map[FileOptionType]*FileOptionHandler)
	for i := range fileOptionHandlers {
		m[fileOptionHandlers[i].Option] = &fileOptionHandlers[i]
	}
	return m
}()

// ============================================================================
// Main Apply Functions
// ============================================================================

// ApplyManagedMode applies managed mode settings to file descriptors.
func ApplyManagedMode(
	descriptors []*descriptorpb.FileDescriptorProto,
	config ManagedModeConfig,
	fileToModule map[string]string,
) error {
	if !config.Enabled {
		return nil
	}

	for _, fd := range descriptors {
		filePath := fd.GetName()
		module := fileToModule[filePath]
		pkg := fd.GetPackage()

		// Ensure Options is initialized
		if fd.Options == nil {
			fd.Options = &descriptorpb.FileOptions{}
		}

		// Apply all file options
		applyFileOptions(fd, config, filePath, module, pkg)

		// Apply field options to all messages
		applyFieldOptionsToMessages(fd.GetMessageType(), config, filePath, module, pkg)
	}

	return nil
}

// applyFileOptions applies all registered file options.
// The logic follows buf's managed mode behavior:
// 1. First, apply all overrides in order (last matching rule wins)
// 2. Then, apply defaults for options that have defaults and weren't overridden
//
// Important: Some options like go_package_prefix affect other options (go_package).
// When a prefix/suffix option is applied, it marks the base option as applied.
func applyFileOptions(
	fd *descriptorpb.FileDescriptorProto,
	config ManagedModeConfig,
	filePath, module, pkg string,
) {
	// Track which options have been applied via override
	appliedOptions := make(map[FileOptionType]bool)

	// First pass: apply overrides in order (last one wins)
	// This matches buf's behavior: "If multiple overrides for the same option apply
	// to a file or field, the last rule takes effect."
	for _, override := range config.Override {
		if override.FileOption == "" {
			continue
		}

		handler, exists := fileOptionHandlerMap[override.FileOption]
		if !exists {
			continue
		}

		// Check if this override matches the current file context
		if !override.matchesFileContext(filePath, module) {
			continue
		}

		// Check if this option is disabled
		if config.IsFileOptionDisabled(filePath, module, override.FileOption) {
			continue
		}

		// Apply the override
		handler.Apply(fd, override.Value, pkg)
		appliedOptions[override.FileOption] = true

		// If this handler affects another option (e.g., go_package_prefix affects go_package),
		// mark the affected option as applied too
		if handler.AffectsOption != "" {
			appliedOptions[handler.AffectsOption] = true
		}
	}

	// Second pass: apply defaults for options that have defaults and weren't overridden
	// According to buf documentation, only certain options have defaults
	for i := range fileOptionHandlers {
		handler := &fileOptionHandlers[i]

		// Skip if no default
		if !handler.HasDefault {
			continue
		}

		// Skip if this option was already applied
		if appliedOptions[handler.Option] {
			continue
		}

		// Skip if the option this handler affects was already applied
		// (e.g., if go_package was explicitly set, don't apply go_package_prefix default)
		if handler.AffectsOption != "" && appliedOptions[handler.AffectsOption] {
			continue
		}

		// Check if this option is disabled
		if config.IsFileOptionDisabled(filePath, module, handler.Option) {
			continue
		}

		// Apply default value
		defaultValue := handler.Default(fd, pkg)
		handler.Apply(fd, defaultValue, pkg)
		appliedOptions[handler.Option] = true

		// Mark affected option as applied
		if handler.AffectsOption != "" {
			appliedOptions[handler.AffectsOption] = true
		}
	}
}

// applyFieldOptionsToMessages recursively applies field options to all messages.
func applyFieldOptionsToMessages(
	messages []*descriptorpb.DescriptorProto,
	config ManagedModeConfig,
	filePath, module, parentPath string,
) {
	for _, msg := range messages {
		messagePath := parentPath + "." + msg.GetName()

		// Apply options to each field
		for _, field := range msg.GetField() {
			fieldPath := messagePath + "." + field.GetName()
			applyFieldOptions(field, config, filePath, module, fieldPath)
		}

		// Recursively process nested messages
		applyFieldOptionsToMessages(msg.GetNestedType(), config, filePath, module, messagePath)
	}
}

// applyFieldOptions applies all registered field options to a field.
// According to buf documentation, field options (like jstype) don't have defaults
// and are only applied when explicitly overridden.
func applyFieldOptions(
	field *descriptorpb.FieldDescriptorProto,
	config ManagedModeConfig,
	filePath, module, fieldPath string,
) {
	for i := range fieldOptionHandlers {
		handler := &fieldOptionHandlers[i]

		// Check if this handler applies to this field type
		if !handler.AppliesToType(field.GetType()) {
			continue
		}

		// Check if this option is disabled
		if config.IsFieldOptionDisabled(filePath, module, handler.Option, fieldPath) {
			continue
		}

		// Get override value - field options only apply when explicitly overridden
		override := config.GetFieldOptionOverride(filePath, module, handler.Option, fieldPath)
		if override != nil {
			handler.Apply(field, override)
			continue
		}

		// Apply default if this option has one
		if handler.HasDefault && handler.Default != nil {
			defaultValue := handler.Default()
			handler.Apply(field, defaultValue)
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// generateObjcClassPrefix generates the default Objective-C class prefix.
func generateObjcClassPrefix(pkg string) string {
	parts := strings.Split(pkg, ".")
	var prefix strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			prefix.WriteRune(unicode.ToUpper(rune(part[0])))
		}
	}

	result := prefix.String()

	// Pad with 'X' if less than 3 characters
	for len(result) < 3 {
		result += "X"
	}

	// Change "GPB" to "GPX" (reserved by Google Protobuf)
	if result == "GPB" {
		result = "GPX"
	}

	return result
}

// toPascalCase converts a string to PascalCase (e.g., "hello_world" -> "HelloWorld").
//
// This function is used to convert protobuf package names (typically lowercase with dots)
// to the naming conventions required by different programming languages:
//   - C# namespaces use PascalCase (e.g., "Acme.Weather.V1")
//   - Ruby modules use PascalCase (e.g., "Acme::Weather::V1")
//   - PHP namespaces use PascalCase (e.g., "Acme\Weather\V1")
//   - Java outer classnames use PascalCase (e.g., "TestProto")
//
// This follows the buf managed mode defaults, which are based on the official naming
// conventions for each language. Without this conversion, generated code would not
// conform to language-specific coding standards.
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' || r == '-' || r == '.' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// toPascalCaseWithSeparator converts a protobuf package name to PascalCase with a custom separator.
//
// This function is used to generate language-specific namespace/package names from protobuf
// package names. Protobuf packages are typically lowercase with dots (e.g., "acme.weather.v1"),
// but different languages require different formats:
//   - C#: PascalCase with "." separator -> "Acme.Weather.V1"
//   - Ruby: PascalCase with "::" separator -> "Acme::Weather::V1"
//   - PHP: PascalCase with "\" separator -> "Acme\Weather\V1"
//
// Examples:
//   - toPascalCaseWithSeparator("acme.weather.v1", ".")  -> "Acme.Weather.V1"
//   - toPascalCaseWithSeparator("acme.weather.v1", "::")  -> "Acme::Weather::V1"
//   - toPascalCaseWithSeparator("acme.weather.v1", `\`)   -> "Acme\Weather\V1"
//
// This follows the buf managed mode defaults, which are based on official naming conventions:
//   - C#: Microsoft C# Coding Conventions
//   - Ruby: Ruby Style Guide
//   - PHP: PSR-1 Basic Coding Standard
//
// Without this conversion, generated code would not conform to language-specific standards.
func toPascalCaseWithSeparator(pkg, separator string) string {
	parts := strings.Split(pkg, ".")
	for i, part := range parts {
		parts[i] = toPascalCase(part)
	}
	return strings.Join(parts, separator)
}
