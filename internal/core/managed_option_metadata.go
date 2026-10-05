package core

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// ManagedOptionMetadata describes the configuration values accepted by a managed handler.
type ManagedOptionMetadata struct {
	Name      string
	ValueType string
	Enum      []string
}

// ManagedFileOptionMetadata returns fresh metadata for the registered file handlers.
// Prefix/suffix handlers accept the same value type as the option they affect.
func ManagedFileOptionMetadata() []ManagedOptionMetadata {
	fields := (&descriptorpb.FileOptions{}).ProtoReflect().Descriptor().Fields()
	options := make([]ManagedOptionMetadata, 0, len(fileOptionHandlers))
	for _, handler := range fileOptionHandlers {
		option := handler.Option
		if handler.AffectsOption != "" {
			option = handler.AffectsOption
		}
		options = append(options, managedOptionMetadata(string(handler.Option), fields.ByName(protoreflect.Name(option))))
	}
	return options
}

// ManagedFieldOptionMetadata returns fresh metadata for the registered field handlers.
func ManagedFieldOptionMetadata() []ManagedOptionMetadata {
	fields := (&descriptorpb.FieldOptions{}).ProtoReflect().Descriptor().Fields()
	options := make([]ManagedOptionMetadata, 0, len(fieldOptionHandlers))
	for _, handler := range fieldOptionHandlers {
		options = append(options, managedOptionMetadata(string(handler.Option), fields.ByName(protoreflect.Name(handler.Option))))
	}
	return options
}

func managedOptionMetadata(name string, field protoreflect.FieldDescriptor) ManagedOptionMetadata {
	metadata := ManagedOptionMetadata{Name: name, ValueType: "string"}
	switch field.Kind() {
	case protoreflect.BoolKind:
		metadata.ValueType = "boolean"
	case protoreflect.EnumKind:
		values := field.Enum().Values()
		for i := 0; i < values.Len(); i++ {
			metadata.Enum = append(metadata.Enum, string(values.Get(i).Name()))
		}
	}
	return metadata
}
