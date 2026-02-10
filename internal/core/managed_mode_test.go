package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestApplyManagedMode_Defaults(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// Check Java defaults
	assert.Equal(t, "com.acme.weather.v1", fd.Options.GetJavaPackage())
	assert.True(t, fd.Options.GetJavaMultipleFiles())
	assert.Equal(t, "TestProto", fd.Options.GetJavaOuterClassname())

	// Check C# defaults
	assert.Equal(t, "Acme.Weather.V1", fd.Options.GetCsharpNamespace())

	// Check Ruby defaults
	assert.Equal(t, "Acme::Weather::V1", fd.Options.GetRubyPackage())

	// Check PHP defaults
	assert.Equal(t, `Acme\Weather\V1`, fd.Options.GetPhpNamespace())

	// Check Objective-C defaults
	assert.Equal(t, "AWV", fd.Options.GetObjcClassPrefix())

	// Check C++ defaults
	assert.True(t, fd.Options.GetCcEnableArenas())

	// Go options should NOT be set (no default)
	assert.Nil(t, fd.Options.GoPackage)
}

func TestApplyManagedMode_GoPackagePrefix(t *testing.T) {
	tests := []struct {
		name         string
		fileName     string
		protoPackage string
		goPackage    string
	}{
		{
			name:         "two_dots_in_package",
			fileName:     "acme/weather/v1/weather.proto",
			protoPackage: "acme.weather.v1",
			goPackage:    "github.com/acme/weather/gen/go/acme/weather/v1;weatherv1",
		},
		{
			name:         "one_dot_in_package",
			fileName:     "task/v1/task.proto",
			protoPackage: "task.v1",
			goPackage:    "github.com/acme/weather/gen/go/task/v1;taskv1",
		},
		{
			name:         "no_dots_in_package",
			fileName:     "common/v1/common.proto",
			protoPackage: "common",
			goPackage:    "github.com/acme/weather/gen/go/common/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fd := &descriptorpb.FileDescriptorProto{
				Name:    strPtr(tt.fileName),
				Package: strPtr(tt.protoPackage),
				Options: &descriptorpb.FileOptions{},
			}

			config := ManagedModeConfig{
				Enabled: true,
				Override: []ManagedOverrideRule{
					{
						FileOption: FileOptionGoPackagePrefix,
						Value:      "github.com/acme/weather/gen/go",
					},
				},
			}

			fileToModule := map[string]string{
				tt.fileName: "",
			}

			err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
			require.NoError(t, err)

			assert.Equal(t, tt.goPackage, fd.Options.GetGoPackage())
		})
	}
}

func TestApplyManagedMode_JavaPackageSuffix(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FileOption: FileOptionJavaPackageSuffix,
				Value:      "grpc",
			},
		},
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// java_package_suffix should override java_package_prefix default
	expected := "acme.weather.v1.grpc"
	assert.Equal(t, expected, fd.Options.GetJavaPackage())
}

func TestApplyManagedMode_DisableOption(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Disable: []ManagedDisableRule{
			{
				FileOption: FileOptionJavaPackagePrefix,
			},
		},
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// java_package should NOT be set because java_package_prefix is disabled
	assert.Nil(t, fd.Options.JavaPackage)
}

func TestApplyManagedMode_DisableForModule(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Disable: []ManagedDisableRule{
			{
				Module: "buf.build/googleapis/googleapis",
			},
		},
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "buf.build/googleapis/googleapis",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// No options should be set for googleapis module
	assert.Nil(t, fd.Options.JavaPackage)
}

func TestApplyManagedMode_OverrideForModule(t *testing.T) {
	fd1 := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("acme/weather/v1/weather.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	fd2 := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("acme/pet/v1/pet.proto"),
		Package: strPtr("acme.pet.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FileOption: FileOptionGoPackagePrefix,
				Value:      "github.com/acme/default/gen/go",
			},
			{
				FileOption: FileOptionGoPackagePrefix,
				Module:     "buf.build/acme/petapis",
				Value:      "github.com/acme/pet/gen/go",
			},
		},
	}

	fileToModule := map[string]string{
		"acme/weather/v1/weather.proto": "",
		"acme/pet/v1/pet.proto":         "buf.build/acme/petapis",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd1, fd2}, config, fileToModule)
	require.NoError(t, err)

	// fd1 should use default prefix
	expected1 := "github.com/acme/default/gen/go/acme/weather/v1;weatherv1"
	assert.Equal(t, expected1, fd1.Options.GetGoPackage())

	// fd2 should use module-specific prefix
	expected2 := "github.com/acme/pet/gen/go/acme/pet/v1;petv1"
	assert.Equal(t, expected2, fd2.Options.GetGoPackage())
}

func TestApplyManagedMode_LastRuleWins(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("acme/weather/v1/weather.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FileOption: FileOptionGoPackage,
				Value:      "first/value",
			},
			{
				FileOption: FileOptionGoPackagePrefix,
				Value:      "second/prefix",
			},
		},
	}

	fileToModule := map[string]string{
		"acme/weather/v1/weather.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// Last rule (go_package_prefix) should win
	expected := "second/prefix/acme/weather/v1;weatherv1"
	assert.Equal(t, expected, fd.Options.GetGoPackage())
}

func TestApplyManagedMode_CsharpNamespacePrefix(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FileOption: FileOptionCsharpNamespacePrefix,
				Value:      "Data",
			},
		},
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	expected := "Data.Acme.Weather.V1"
	assert.Equal(t, expected, fd.Options.GetCsharpNamespace())
}

func TestApplyManagedMode_OptimizeFor(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FileOption: FileOptionOptimizeFor,
				Value:      string(OptimizeModeCodeSize),
			},
		},
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	assert.Equal(t, descriptorpb.FileOptions_CODE_SIZE, fd.Options.GetOptimizeFor())
}

func TestApplyManagedMode_Disabled(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: false,
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// No options should be set when managed mode is disabled
	assert.Nil(t, fd.Options.JavaPackage)
}

func TestGenerateObjcClassPrefix(t *testing.T) {
	tests := map[string]struct {
		pkg      string
		expected string
	}{
		"acme.weather.v1": {
			pkg:      "acme.weather.v1",
			expected: "AWV",
		},
		"foo.bar": {
			pkg:      "foo.bar",
			expected: "FBX", // Less than 3 chars, padded with X
		},
		"a": {
			pkg:      "a",
			expected: "AXX", // Single char, padded with XX
		},
		"google.protobuf.bar": {
			pkg:      "google.protobuf.bar",
			expected: "GPX", // GPB is reserved, changed to GPX
		},
		"empty": {
			pkg:      "",
			expected: "XXX", // Empty package
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := generateObjcClassPrefix(tt.pkg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"hello_world": {
			input:    "hello_world",
			expected: "HelloWorld",
		},
		"hello-world": {
			input:    "hello-world",
			expected: "HelloWorld",
		},
		"hello.world": {
			input:    "hello.world",
			expected: "HelloWorld",
		},
		"helloWorld": {
			input:    "helloWorld",
			expected: "HelloWorld",
		},
		"HelloWorld": {
			input:    "HelloWorld",
			expected: "HelloWorld",
		},
		"empty": {
			input:    "",
			expected: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := toPascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToPascalCaseWithSeparator(t *testing.T) {
	tests := map[string]struct {
		pkg       string
		separator string
		expected  string
	}{
		"acme.weather.v1_.": {
			pkg:       "acme.weather.v1",
			separator: ".",
			expected:  "Acme.Weather.V1",
		},
		"acme.weather.v1_::": {
			pkg:       "acme.weather.v1",
			separator: "::",
			expected:  "Acme::Weather::V1",
		},
		"acme.weather.v1_\\": {
			pkg:       "acme.weather.v1",
			separator: `\`,
			expected:  `Acme\Weather\V1`,
		},
		"foo_.": {
			pkg:       "foo",
			separator: ".",
			expected:  "Foo",
		},
		"empty_.": {
			pkg:       "",
			separator: ".",
			expected:  "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := toPascalCaseWithSeparator(tt.pkg, tt.separator)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestApplyManagedMode_FieldOptions(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: strPtr("TestMessage"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   strPtr("id"),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_INT64.Enum(),
						Number: int32Ptr(1),
					},
					{
						Name:   strPtr("name"),
						Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
						Number: int32Ptr(2),
					},
				},
			},
		},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FieldOption: FieldOptionJsType,
				Value:       string(JSTypeString),
			},
		},
	}

	fileToModule := map[string]string{
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// Check that jstype was applied to int64 field
	idField := fd.MessageType[0].Field[0]
	require.NotNil(t, idField.Options)
	assert.Equal(t, descriptorpb.FieldOptions_JS_STRING, idField.Options.GetJstype())

	// Check that jstype was NOT applied to string field (wrong type)
	nameField := fd.MessageType[0].Field[1]
	if nameField.Options != nil {
		assert.Nil(t, nameField.Options.Jstype)
	}
}

func TestApplyManagedMode_ExternalModules(t *testing.T) {
	// Files from external modules should only be processed if there's an explicit rule
	localFile := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("api/v1/service.proto"),
		Package: strPtr("api.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	externalFile := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("google/api/annotations.proto"),
		Package: strPtr("google.api"),
		Options: &descriptorpb.FileOptions{
			GoPackage: strPtr("google.golang.org/genproto/googleapis/api/annotations"),
		},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FileOption: FileOptionGoPackagePrefix,
				Value:      "github.com/example/ec-code/gen/go",
			},
		},
	}

	fileToModule := map[string]string{
		"api/v1/service.proto":         "",                                 // Local file
		"google/api/annotations.proto": "github.com/googleapis/googleapis", // External module
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{localFile, externalFile}, config, fileToModule)
	require.NoError(t, err)

	// Local file should get go_package from managed mode
	assert.Equal(t, "github.com/example/ec-code/gen/go/api/v1;apiv1", localFile.Options.GetGoPackage())

	// External file should keep its original go_package (no rule for this module)
	assert.Equal(t, "google.golang.org/genproto/googleapis/api/annotations", externalFile.Options.GetGoPackage())
}

func TestApplyManagedMode_ExternalModuleWithRule(t *testing.T) {
	// External module file should be processed if there's an explicit rule
	externalFile := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("google/api/annotations.proto"),
		Package: strPtr("google.api"),
		Options: &descriptorpb.FileOptions{},
	}

	config := ManagedModeConfig{
		Enabled: true,
		Override: []ManagedOverrideRule{
			{
				FileOption: FileOptionGoPackagePrefix,
				Value:      "github.com/example/ec-code/gen/go",
				Module:     "github.com/googleapis/googleapis", // Explicit rule for this module
			},
		},
	}

	fileToModule := map[string]string{
		"google/api/annotations.proto": "github.com/googleapis/googleapis",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{externalFile}, config, fileToModule)
	require.NoError(t, err)

	// External file should get go_package from managed mode because there's a rule for its module
	// Package: google.api -> take last 2 segments -> googleapi
	assert.Equal(t, "github.com/example/ec-code/gen/go/google/api;googleapi", externalFile.Options.GetGoPackage())
}

// Test cleanPackageName function
func TestCleanPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple package with dot",
			input:    "task.v1",
			expected: "taskv1",
		},
		{
			name:     "multiple dots",
			input:    "acme.weather.v1",
			expected: "acmeweatherv1",
		},
		{
			name:     "no dots",
			input:    "simple",
			expected: "simple",
		},
		{
			name:     "with underscores",
			input:    "my_package.v1",
			expected: "my_packagev1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPackageName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions
func TestMatchesPath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		rulePath string
		expected bool
	}{
		// Directory path ending with "/"
		{
			name:     "directory path matches file in directory",
			filePath: "internal/cms/as.proto",
			rulePath: "internal/cms/",
			expected: true,
		},
		{
			name:     "directory path matches file in subdirectory",
			filePath: "internal/cms/v1/service.proto",
			rulePath: "internal/cms/",
			expected: true,
		},
		{
			name:     "directory path does not match file outside",
			filePath: "internal/cmsv2/file.proto",
			rulePath: "internal/cms/",
			expected: false,
		},
		// Exact file path ending with ".proto"
		{
			name:     "exact file path matches",
			filePath: "internal/cms/as.proto",
			rulePath: "internal/cms/as.proto",
			expected: true,
		},
		{
			name:     "exact file path does not match different file",
			filePath: "internal/cms/node.proto",
			rulePath: "internal/cms/as.proto",
			expected: false,
		},
		// Prefix path (no trailing "/" or ".proto")
		{
			name:     "prefix path matches file in directory",
			filePath: "internal/cms/as.proto",
			rulePath: "internal/cms",
			expected: true,
		},
		{
			name:     "prefix path matches file with similar prefix (buf behavior)",
			filePath: "internal/cmsv2/file.proto",
			rulePath: "internal/cms",
			expected: true, // buf uses prefix matching, not directory-aware
		},
		{
			name:     "prefix path does not match different prefix",
			filePath: "internal/svc/service.proto",
			rulePath: "internal/cms",
			expected: false,
		},
		// Edge cases
		{
			name:     "empty rule path matches all",
			filePath: "any/path/file.proto",
			rulePath: "",
			expected: true,
		},
		{
			name:     "exact match with same prefix",
			filePath: "internal/cms",
			rulePath: "internal/cms",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := ManagedOverrideRule{
				Path: tt.rulePath,
			}
			result := rule.matchesFileContext(tt.filePath, "")
			assert.Equal(t, tt.expected, result, "filePath: %s, rulePath: %s", tt.filePath, tt.rulePath)
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
