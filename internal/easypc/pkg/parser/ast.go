package parser

import (
	"fmt"
	"math"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
)

// DeclKind identifies the type of an AST declaration for declaration ordering.
type DeclKind uint8

const (
	DeclSyntax DeclKind = iota
	DeclPackage
	DeclImport
	DeclOption
	DeclMessage
	DeclEnum
	DeclService
	DeclExtend
	DeclField
	DeclOneof
	DeclExtensionRange
	DeclReserved
	DeclEnumValue
	DeclMethod
)

// DeclRef records a declaration's kind and its index in the corresponding slab.
type DeclRef struct {
	Kind  DeclKind
	Index uint32
}

// NodeID represents a 32-bit compact index referencing a node within an arena slab.
type NodeID uint32

const NoNode NodeID = math.MaxUint32

// Span32 represents a contiguous index range [Start, Start+Count) in an arena slice.
type Span32 struct {
	Start uint32
	Count uint32
}

func (s Span32) IsEmpty() bool { return s.Count == 0 }
func (s Span32) End() uint32   { return s.Start + s.Count }

// TextSpan represents byte offsets in the source buffer [Start, End).
type TextSpan struct {
	Start uint32
	End   uint32
}

func (t TextSpan) IsEmpty() bool           { return t.Start == t.End }
func (t TextSpan) Len() uint32             { return t.End - t.Start }
func (t TextSpan) Bytes(src []byte) []byte { return src[t.Start:t.End] }
func (t TextSpan) String(src []byte) string {
	if t.IsEmpty() {
		return ""
	}
	return string(src[t.Start:t.End])
}

// SyntaxKind identifies the protobuf syntax specification.
type SyntaxKind uint8

const (
	SyntaxUnspecified SyntaxKind = iota // Default is proto2
	SyntaxProto2
	SyntaxProto3
	SyntaxEditions
)

// Edition represents protobuf edition numbers.
type Edition uint32

const (
	EditionUnknown Edition = 0
	EditionLegacy  Edition = 900
	EditionProto2  Edition = 998
	EditionProto3  Edition = 999
	Edition2023    Edition = 1000
	Edition2024    Edition = 1001
	Edition2026    Edition = 1002
)

// ImportKind identifies the import modifier.
type ImportKind uint8

const (
	ImportDefault ImportKind = iota
	ImportPublic
	ImportWeak
)

// ImportNode records an import statement.
type ImportNode struct {
	Path         TextSpan   // Unquoted import path
	Kind         ImportKind // Default, Public, Weak
	Span         TextSpan   // Entire statement span
	ModifierSpan TextSpan   // "public" or "weak" token span
}

// FieldLabel identifies field cardinality.
type FieldLabel uint8

const (
	LabelUnspecified FieldLabel = 0
	LabelOptional    FieldLabel = 1
	LabelRequired    FieldLabel = 2
	LabelRepeated    FieldLabel = 3
)

// FieldTypeKind identifies field primitive types or unresolved named types.
type FieldTypeKind uint8

const (
	TypeUnresolved FieldTypeKind = 0 // Named message or enum type
	TypeDouble     FieldTypeKind = 1
	TypeFloat      FieldTypeKind = 2
	TypeInt64      FieldTypeKind = 3
	TypeUint64     FieldTypeKind = 4
	TypeInt32      FieldTypeKind = 5
	TypeFixed64    FieldTypeKind = 6
	TypeFixed32    FieldTypeKind = 7
	TypeBool       FieldTypeKind = 8
	TypeString     FieldTypeKind = 9
	TypeGroup      FieldTypeKind = 10
	TypeMessage    FieldTypeKind = 11
	TypeBytes      FieldTypeKind = 12
	TypeUint32     FieldTypeKind = 13
	TypeEnum       FieldTypeKind = 14
	TypeSfixed32   FieldTypeKind = 15
	TypeSfixed64   FieldTypeKind = 16
	TypeSint32     FieldTypeKind = 17
	TypeSint64     FieldTypeKind = 18
)

// FieldNode records a message or extension field declaration.
type FieldNode struct {
	Name           TextSpan
	TypeName       TextSpan // Raw type identifier (e.g. "int32", "string", "MyMessage", ".foo.Bar")
	Number         int32    // Field tag number
	Label          FieldLabel
	TypeKind       FieldTypeKind
	OneofIndex     int32  // 0-based index in message.Oneofs, or -1 if none
	Proto3Optional bool   // True if proto3 explicit optional
	Span           TextSpan
	Options        Span32 // Index range in ASTFile.Options

	CustomJSONName TextSpan // Explicit [json_name = "..."], or empty
	DefaultVal     TextSpan // Explicit [default = "..."], or empty
	DefaultOpt     TextSpan // "default = ..." item span as written, or empty
	Extendee       TextSpan // Target message if this is an extension
	ExtendeeSpan   TextSpan // Original source extendee span

	LabelSpan         TextSpan // "optional", "required", "repeated" token span
	TypeSpan          TextSpan // source type span (e.g. "map<string, int32>" for map fields)
	NumberSpan        TextSpan // field number token span
	OptionsSpan       TextSpan // [...] options bracket span
	CustomJSONNameOpt TextSpan // "json_name = ..." item span
	CustomJSONNameVal TextSpan // "..." string value span
	DefaultValSpan    TextSpan // default value token span
	GroupKeywordSpan  TextSpan // "group" keyword token span

	CustomOptionsWire []byte
}

// ExtendNode records an extend declaration block.
type ExtendNode struct {
	Extendee      TextSpan
	ExtendeeSpan  TextSpan
	Span          TextSpan
	OpenBraceSpan TextSpan
	FieldIndices  Span32 // in ASTFile.ChildIndices
}

// MessageNode records a message declaration.
type MessageNode struct {
	Name            TextSpan
	Span            TextSpan
	OpenBraceSpan   TextSpan
	Fields          Span32 // Index range in ASTFile.Fields
	NestedMessages  Span32 // Index range in ASTFile.Messages
	NestedEnums     Span32 // Index range in ASTFile.Enums
	Oneofs          Span32 // Index range in ASTFile.Oneofs
	ExtensionRanges Span32 // Index range in ASTFile.ExtensionRanges
	Extensions      Span32 // Index range in ASTFile.Extensions (nested extend blocks)
	ReservedRanges  Span32 // Index range in ASTFile.ReservedRanges
	ReservedNames   Span32 // Index range in ASTFile.ReservedNames
	Options         Span32 // Index range in ASTFile.Options
	Decls           []DeclRef

	IsMapEntry bool // Synthetic map entry
	IsGroup    bool // Synthetic group message

	CustomOptionsWire []byte
}

// OneofNode records a oneof declaration.
type OneofNode struct {
	Name          TextSpan
	Span          TextSpan
	OpenBraceSpan TextSpan
	Options       Span32 // Index range in ASTFile.Options
	Decls         []DeclRef
	IsSynthetic   bool

	CustomOptionsWire []byte
}

// EnumNode records an enum declaration.
type EnumNode struct {
	Name           TextSpan
	Span           TextSpan
	OpenBraceSpan  TextSpan
	Values         Span32 // Index range in ASTFile.EnumValues
	ReservedRanges Span32 // Index range in ASTFile.ReservedRanges
	ReservedNames  Span32 // Index range in ASTFile.ReservedNames
	Options        Span32 // Index range in ASTFile.Options
	Decls          []DeclRef

	CustomOptionsWire []byte
}

// EnumValueNode records a single enum value declaration.
type EnumValueNode struct {
	Name        TextSpan
	Number      int32
	NumberSpan  TextSpan
	Span        TextSpan
	OptionsSpan TextSpan
	Options     Span32 // Index range in ASTFile.Options

	CustomOptionsWire []byte
}

// ServiceNode records a service declaration.
type ServiceNode struct {
	Name          TextSpan
	Span          TextSpan
	OpenBraceSpan TextSpan
	Methods       Span32 // Index range in ASTFile.Methods
	Options       Span32 // Index range in ASTFile.Options
	Decls         []DeclRef

	CustomOptionsWire []byte
}

// MethodNode records an rpc method declaration.
type MethodNode struct {
	Name                TextSpan
	InputType           TextSpan
	OutputType          TextSpan
	InputTypeSpan       TextSpan
	OutputTypeSpan      TextSpan
	ClientStreaming     bool
	ServerStreaming     bool
	ClientStreamingSpan TextSpan
	ServerStreamingSpan TextSpan
	Span                TextSpan
	OpenBraceSpan       TextSpan
	Options             Span32 // Index range in ASTFile.Options
	HasBody             bool   // True if rpc method was defined with a body { ... }

	CustomOptionsWire []byte
}

// ExtensionRangeNode records an extension range declaration.
type ExtensionRangeNode struct {
	Start       int32
	End         int32  // Exclusive: protoc convention is end + 1
	StartSpan   TextSpan
	EndSpan     TextSpan
	OptionsSpan TextSpan
	Options     Span32 // Index range in ASTFile.Options
	Span        TextSpan

	CustomOptionsWire []byte
}

// ExtensionRangeDeclNode records an extensions statement block.
type ExtensionRangeDeclNode struct {
	Span         TextSpan
	RangeIndices Span32 // in ASTFile.ChildIndices
}

// ReservedRangeNode records a reserved range declaration.
type ReservedRangeNode struct {
	Start     int32
	End       int32 // Exclusive for messages (e.g. start=4, end=5), inclusive for enums
	Span      TextSpan
	StartSpan TextSpan
	EndSpan   TextSpan
}

// ReservedDeclNode records a reserved statement.
type ReservedDeclNode struct {
	Span        TextSpan
	IsNames     bool
	ItemIndices Span32 // in ASTFile.ChildIndices
}

// OptionValKind identifies the type of an option value literal.
type OptionValKind uint8

const (
	OptionValIdent OptionValKind = iota
	OptionValInt
	OptionValFloat
	OptionValString
	OptionValAggregate
	OptionValList
)

// OptionNode records an option declaration.
type OptionNode struct {
	NameParts  Span32 // Index range in ASTFile.OptionNameParts
	ValKind    OptionValKind
	ValSpan    TextSpan
	IntVal     uint64
	IsNegative bool
	FloatVal   float64
	SubOptions Span32 // Index range in ASTFile.Options if ValKind == OptionValAggregate
	Span       TextSpan
	NameSpan   TextSpan // Option name as written, e.g. "(google.api.http).post"
	ValRawSpan TextSpan // Value tokens as written: quotes, sign and every concatenated literal
}

// OptionNamePartNode records a single component of an option path.
type OptionNamePartNode struct {
	Name        TextSpan
	IsExtension bool // True if in parentheses: (foo.bar)
}

// Diagnostic records a compilation diagnostic with source position.
type Diagnostic struct {
	Path    string
	Line    int
	Col     int
	Message string
	IsWarn  bool
	// Semantic marks checks on a syntactically valid file (duplicate names or
	// numbers, field number ranges, option values); tools that only need the
	// syntax tree may ignore them.
	Semantic bool
}

func (d Diagnostic) String() string {
	if d.Line <= 0 {
		if d.IsWarn {
			return fmt.Sprintf("%s: warning: %s", d.Path, d.Message)
		}
		return fmt.Sprintf("%s: %s", d.Path, d.Message)
	}
	if d.IsWarn {
		return fmt.Sprintf("%s:%d:%d: warning: %s", d.Path, d.Line, d.Col, d.Message)
	}
	return fmt.Sprintf("%s:%d:%d: %s", d.Path, d.Line, d.Col, d.Message)
}

// ASTFile is the root container for a parsed .proto file.
// All slices are contiguous pointer-free slabs, avoiding GC traversal.
type ASTFile struct {
	Path        string             // Canonical path relative to import root
	DisplayPath string             // User-facing display path for diagnostics
	Src         []byte             // Retained source buffer (and interned synthetic bytes)
	SrcLen      uint32             // Length of the original source; spans at or past it are synthetic
	Tokens      *lexer.TokenStream // Retained lexer token stream for sourceinfo
	Syntax      SyntaxKind         // proto2, proto3, editions
	Edition     Edition            // e.g. Edition2023
	SyntaxSpan  TextSpan           // Entire syntax/edition statement span
	Package     TextSpan           // Package identifier
	PackageSpan TextSpan           // Entire package statement span

	TopLevelMessages Span32 // Range in ChildIndices pointing to Messages
	TopLevelEnums    Span32 // Range in ChildIndices pointing to Enums
	FileOptions      Span32 // Range in ChildIndices pointing to Options
	ChildIndices     []uint32

	Imports             []ImportNode
	Messages            []MessageNode
	Fields              []FieldNode
	Enums               []EnumNode
	EnumValues          []EnumValueNode
	Services            []ServiceNode
	Methods             []MethodNode
	Oneofs              []OneofNode
	ExtensionRanges     []ExtensionRangeNode
	ExtensionRangeDecls []ExtensionRangeDeclNode
	ReservedRanges      []ReservedRangeNode
	ReservedNames       []TextSpan
	ReservedNameSpans   []TextSpan
	ReservedDecls       []ReservedDeclNode
	Extensions          []FieldNode // Top-level and nested extension fields
	Extends             []ExtendNode
	Options             []OptionNode
	OptionNameParts     []OptionNamePartNode

	Decls []DeclRef // Source order of top-level declarations

	SourceInfo any // *sourceinfo.FileSourceInfo
	SourceInfoWire []byte // Serialized SourceCodeInfo bytes

	CustomOptionsWire []byte

	Errors []Diagnostic
}

func (f *ASTFile) InternString(s string) TextSpan {
	start := uint32(len(f.Src))
	f.Src = append(f.Src, s...)
	return TextSpan{Start: start, End: uint32(len(f.Src))}
}

// LineCol returns 1-based line and 1-based column with tab stop 8 for offset.
func (f *ASTFile) LineCol(offset uint32) (int, int) {
	if f.Tokens != nil {
		return f.Tokens.LineCol(f.Src, offset)
	}
	if offset > uint32(len(f.Src)) {
		offset = uint32(len(f.Src))
	}
	line := 1
	col0 := 0
	for i := 0; i < int(offset); i++ {
		b := f.Src[i]
		if b == '\n' {
			line++
			col0 = 0
		} else if b == '\t' {
			col0 = (col0/8 + 1) * 8
		} else {
			col0++
		}
	}
	return line, col0 + 1
}
