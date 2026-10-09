package core

import (
	"fmt"
	"io"
	"reflect"

	"github.com/yoheimuta/go-protoparser/v4"
	"github.com/yoheimuta/go-protoparser/v4/interpret/unordered"
	"github.com/yoheimuta/go-protoparser/v4/parser"
)

// readProtoFileWithDirectives retains comments at EOF and before closing braces.
// The unordered interpreter cannot represent standalone comment nodes, so collect
// their positions from the ordered AST before adapting its declaration bodies.
func readProtoFileWithDirectives(reader io.Reader) (*unordered.Proto, []directiveComment, error) {
	parsed, err := protoparser.Parse(reader, protoparser.WithBodyIncludingComments(true))
	if err != nil {
		return nil, nil, fmt.Errorf("Parse: %w", err)
	}
	comments := collectDirectiveComments(parsed)
	stripStandaloneComments(reflect.ValueOf(parsed))
	proto, err := unordered.InterpretProto(parsed)
	if err != nil {
		return nil, nil, fmt.Errorf("InterpretProto: %w", err)
	}
	return proto, comments, nil
}

// stripStandaloneComments removes only comment nodes in declaration bodies.
// Attached Comments/InlineComment fields are retained for existing lint rules.
func stripStandaloneComments(value reflect.Value) {
	if !value.IsValid() {
		return
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if !value.IsNil() {
			stripStandaloneComments(value.Elem())
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			if field.CanInterface() {
				stripStandaloneComments(field)
			}
		}
	case reflect.Slice:
		if value.Type() == reflect.TypeFor[[]parser.Visitee]() {
			kept := 0
			for i := 0; i < value.Len(); i++ {
				element := value.Index(i)
				if _, comment := element.Interface().(*parser.Comment); comment {
					continue
				}
				stripStandaloneComments(element)
				value.Index(kept).Set(element)
				kept++
			}
			value.Slice(kept, value.Len()).Clear()
			value.SetLen(kept)
			return
		}
		for i := 0; i < value.Len(); i++ {
			stripStandaloneComments(value.Index(i))
		}
	}
}
