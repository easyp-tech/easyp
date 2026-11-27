package compilerbackend

import (
	"context"
	"fmt"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/easyp-tech/easyp/internal/parallel"
)

// DescriptorSet owns linked descriptor messages in target/import traversal order.
// Its messages belong to this compilation and may be transformed by the caller.
type DescriptorSet struct {
	Files        []*descriptorpb.FileDescriptorProto
	Dependencies []string
}

// CompileDescriptorSet returns validated descriptor messages with dependencies
// ordered before the files that import them. Set Request.ResolveExtensions to
// also decode user-defined option extensions into typed Go fields.
func CompileDescriptorSet(ctx context.Context, req Request) (DescriptorSet, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	set, err := compileDescriptors(ctx, req)
	if err != nil {
		return DescriptorSet{}, err
	}
	byName := make(map[string]*descriptorpb.FileDescriptorProto, len(set.File))
	for _, file := range set.File {
		byName[file.GetName()] = file
	}
	var descriptors DescriptorSet
	processed := make(map[string]bool, len(set.File))
	var visit func(string, bool) error
	visit = func(name string, dependency bool) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if processed[name] {
			return nil
		}
		file := byName[name]
		if file == nil {
			return fmt.Errorf("compiled source %q has no descriptor", name)
		}
		for _, imported := range file.GetDependency() {
			if err := visit(imported, true); err != nil {
				return err
			}
		}
		descriptors.Files = append(descriptors.Files, file)
		processed[name] = true
		if dependency {
			descriptors.Dependencies = append(descriptors.Dependencies, name)
		}
		return nil
	}
	for _, target := range req.Files {
		if err := visit(filepath.ToSlash(filepath.Clean(target)), false); err != nil {
			return DescriptorSet{}, fmt.Errorf("visit: %w", err)
		}
	}
	return descriptors, nil
}

func compileDescriptors(ctx context.Context, req Request) (*descriptorpb.FileDescriptorSet, error) {
	result, err := compileEasyprotoc(ctx, req)
	if err != nil {
		return nil, err
	}
	wire := result.WireBytes()
	set, err := decodeDescriptorSet(ctx, wire, req.MaxParallelism, proto.UnmarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("decodeDescriptorSet: %w", err)
	}
	registry, err := protodesc.NewFiles(set)
	if err != nil {
		return nil, fmt.Errorf("NewFiles: %w", err)
	}
	if req.ResolveExtensions && hasExtensions(set) {
		// Use the checked registry to decode user-defined option fields.
		unmarshal := proto.UnmarshalOptions{Resolver: descriptorExtensions{dynamicpb.NewTypes(registry)}}
		set, err = decodeDescriptorSet(ctx, wire, req.MaxParallelism, unmarshal)
		if err != nil {
			return nil, fmt.Errorf("decode custom options: %w", err)
		}
	}
	return set, ctx.Err()
}

type descriptorExtensions struct {
	*dynamicpb.Types
}

func (r descriptorExtensions) FindExtensionByName(name protoreflect.FullName) (protoreflect.ExtensionType, error) {
	extension, err := r.Types.FindExtensionByName(name)
	if err == protoregistry.NotFound {
		return protoregistry.GlobalTypes.FindExtensionByName(name)
	}
	return extension, err
}

func (r descriptorExtensions) FindExtensionByNumber(message protoreflect.FullName, number protoreflect.FieldNumber) (protoreflect.ExtensionType, error) {
	extension, err := r.Types.FindExtensionByNumber(message, number)
	if err == protoregistry.NotFound {
		return protoregistry.GlobalTypes.FindExtensionByNumber(message, number)
	}
	return extension, err
}

func hasExtensions(set *descriptorpb.FileDescriptorSet) bool {
	for _, file := range set.GetFile() {
		if len(file.GetExtension()) > 0 || hasNestedExtensions(file.GetMessageType()) {
			return true
		}
	}
	return false
}

func hasNestedExtensions(messages []*descriptorpb.DescriptorProto) bool {
	for _, message := range messages {
		if len(message.GetExtension()) > 0 || hasNestedExtensions(message.GetNestedType()) {
			return true
		}
	}
	return false
}

// decodeDescriptorSet splits the compiler's repeated file field and decodes
// each message independently, retaining the original wire order.
func decodeDescriptorSet(ctx context.Context, wire []byte, limit int, unmarshal proto.UnmarshalOptions) (*descriptorpb.FileDescriptorSet, error) {
	var fields [][]byte
	for remaining := wire; len(remaining) > 0; {
		number, kind, n := protowire.ConsumeTag(remaining)
		if n < 0 || number != 1 || kind != protowire.BytesType {
			fields = nil
			break
		}
		value, width := protowire.ConsumeBytes(remaining[n:])
		if width < 0 {
			fields = nil
			break
		}
		fields = append(fields, value)
		remaining = remaining[n+width:]
	}
	if fields == nil && len(wire) > 0 {
		var set descriptorpb.FileDescriptorSet
		if err := unmarshal.Unmarshal(wire, &set); err != nil {
			return nil, fmt.Errorf("Unmarshal: %w", err)
		}
		return &set, ctx.Err()
	}
	set := &descriptorpb.FileDescriptorSet{File: make([]*descriptorpb.FileDescriptorProto, len(fields))}
	err := parallel.Run(ctx, len(fields), limit, func(index int) error {
		file := &descriptorpb.FileDescriptorProto{}
		if err := unmarshal.Unmarshal(fields[index], file); err != nil {
			return fmt.Errorf("Unmarshal: descriptor %d: %w", index, err)
		}
		set.File[index] = file
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Run: %w", err)
	}
	return set, nil
}
