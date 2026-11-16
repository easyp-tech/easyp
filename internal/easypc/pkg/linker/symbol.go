package linker

import (
	"fmt"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

// SymbolKind identifies the kind of a symbol in the protobuf symbol table.
type SymbolKind uint8

const (
	SymPackage SymbolKind = iota
	SymMessage
	SymEnum
	SymEnumValue
	SymField
	SymExtension
	SymService
	SymMethod
)

// isAggregate reports whether the symbol can contain nested names, matching protoc's Symbol::IsAggregate.
func (k SymbolKind) isAggregate() bool {
	return k == SymPackage || k == SymMessage || k == SymEnum || k == SymService
}

func (k SymbolKind) String() string {
	switch k {
	case SymPackage:
		return "package"
	case SymMessage:
		return "message"
	case SymEnum:
		return "enum"
	case SymEnumValue:
		return "enum_value"
	case SymField:
		return "field"
	case SymExtension:
		return "extension"
	case SymService:
		return "service"
	case SymMethod:
		return "method"
	default:
		return fmt.Sprintf("SymbolKind(%d)", k)
	}
}

const NoParent uint32 = 0xFFFFFFFF

// Symbol represents a declared protobuf entity in the global symbol table.
type Symbol struct {
	ID        uint32
	Kind      SymbolKind
	FileID    uint32
	ParentID  uint32
	Name      string // Local identifier name, e.g. "LaunchStage"
	FullName  string // Fully qualified name without leading dot, e.g. "google.api.LaunchStage"

	NodeIndex uint32 // Index in ASTFile's corresponding slab

	// Type metadata for Fields & Extensions
	TypeKind  parser.FieldTypeKind
	TypeName  string // Canonical type name with leading dot (e.g. ".google.api.LaunchStage")
	Label     parser.FieldLabel
	TagNumber int32
	Extendee  string // Target message with leading dot (e.g. ".google.protobuf.MethodOptions")
	IsPacked  bool
}

// FieldInfo stores field metadata for message options and aggregate field evaluation.
type FieldInfo struct {
	Name       string
	TagNumber  int32
	TypeKind   parser.FieldTypeKind
	TypeName   string // Canonical type name with leading dot
	Label      parser.FieldLabel
	IsPacked   bool
}
