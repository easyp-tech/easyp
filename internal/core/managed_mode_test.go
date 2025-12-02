package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestApplyManagedMode_Defaults(t *testing.T) {
	// Create a test file descriptor
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
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
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
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	expected := "github.com/acme/weather/gen/go/acme/weather/v1"
	assert.Equal(t, expected, fd.Options.GetGoPackage())
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
		Name:    strPtr("test/v1/test.proto"),
		Package: strPtr("acme.weather.v1"),
		Options: &descriptorpb.FileOptions{},
	}

	fd2 := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("pet/v1/pet.proto"),
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
		"test/v1/test.proto": "",
		"pet/v1/pet.proto":   "buf.build/acme/petapis",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd1, fd2}, config, fileToModule)
	require.NoError(t, err)

	// fd1 should use default prefix
	expected1 := "github.com/acme/default/gen/go/acme/weather/v1"
	assert.Equal(t, expected1, fd1.Options.GetGoPackage())

	// fd2 should use module-specific prefix
	expected2 := "github.com/acme/pet/gen/go/acme/pet/v1"
	assert.Equal(t, expected2, fd2.Options.GetGoPackage())
}

func TestApplyManagedMode_LastRuleWins(t *testing.T) {
	fd := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test/v1/test.proto"),
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
		"test/v1/test.proto": "",
	}

	err := ApplyManagedMode([]*descriptorpb.FileDescriptorProto{fd}, config, fileToModule)
	require.NoError(t, err)

	// Last rule (go_package_prefix) should win
	expected := "second/prefix/acme/weather/v1"
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

// Helper functions
func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}
