// Package protoast is the lint view of a single .proto file. It is built from
// easyp-protoc's syntax tree (internal/easypc/protoc/syntax) without linking,
// so a file lints even when its imports cannot be resolved.
//
// Text fields hold the source as written with whitespace and comments removed
// (types keep a leading dot, numbers keep their sign). Positions follow protoc:
// 1-based lines, 1-based columns counted in bytes with tabs advancing to the
// next multiple of 8, and 0-based byte offsets.
package protoast

import (
	"fmt"
	"strings"
)

// Position is a location in the source file.
type Position struct {
	Filename string
	Offset   int
	Line     int
	Column   int
}

// FilePosition is where file-level findings are reported: line 1, column 1.
func FilePosition(path string) Position {
	return Position{Filename: path, Line: 1, Column: 1}
}

func (p Position) String() string {
	name := p.Filename
	if name == "" {
		name = "<input>"
	}
	return fmt.Sprintf("%s:%d:%d", name, p.Line, p.Column)
}

// Comment is a single // line or /* */ block comment.
type Comment struct {
	Raw string   // as written, including the comment markers
	Pos Position // first character
	End Position // last character
}

// Lines returns the comment text without its markers, split on newlines.
func (c *Comment) Lines() []string {
	raw := c.Raw
	if strings.HasPrefix(raw, "/*") {
		raw = strings.TrimSuffix(strings.TrimPrefix(raw, "/*"), "*/")
	} else {
		raw = strings.TrimPrefix(raw, "//")
	}
	return strings.Split(raw, "\n")
}

// Decl is the part every declaration shares.
//
// Comments between the previous declaration and this one are split into
// Leading (the block directly above, no blank line in between) and Detached
// (earlier blocks). Trailing holds the first comment on the line where the
// declaration ends or, for blocks, on the line of its opening brace.
type Decl struct {
	Pos      Position // first token
	End      Position // last token: ';' or '}'
	NamePos  Position // the declared name, for named declarations
	Leading  []*Comment
	Detached []*Comment
	Trailing []*Comment
}

// HasCommentAbove reports whether any comment precedes the declaration.
func (d *Decl) HasCommentAbove() bool { return len(d.Leading)+len(d.Detached) > 0 }

// File is one parsed .proto file.
type File struct {
	Path   string
	Syntax string // "proto2", "proto3" or "editions"; "proto2" when unspecified
	// SyntaxSpecified reports whether the file has a syntax or edition statement.
	SyntaxSpecified bool
	Packages        []*Package
	Imports         []*Import
	Options         []*Option
	Messages        []*Message
	Enums           []*Enum
	Services        []*Service
	Extends         []*Extend

	// Comments lists every comment in source order.
	Comments []*Comment
	owners   map[*Comment]*Decl
}

// PackageName returns the first package name, or "" when the file has none.
func (f *File) PackageName() string {
	if len(f.Packages) == 0 {
		return ""
	}
	return f.Packages[0].Name
}

// Owner returns the declaration a comment annotates: the one it precedes, or
// the one it trails. It returns nil for comments that annotate nothing, such
// as comments at the end of the file or before a closing brace.
func (f *File) Owner(c *Comment) *Decl { return f.owners[c] }

type Package struct {
	Decl
	Name string
}

type Import struct {
	Decl
	Path     string // unquoted
	Location string // the string literal as written, with quotes
	Public   bool
	Weak     bool
}

// Option is an option statement or one entry of a [...] option list.
type Option struct {
	Decl
	Name  string // e.g. "(google.api.http).post" or "go_package"
	Value string // literal as written; adjacent string literals are merged into one
}

type Message struct {
	Decl
	Name     string
	Fields   []*Field // plain fields; map, group and oneof fields are listed separately
	Maps     []*Field
	Groups   []*Field
	Oneofs   []*Oneof
	Messages []*Message
	Enums    []*Enum
	Options  []*Option
	Extends  []*Extend
}

type Field struct {
	Decl
	Label   string // "repeated", "optional", "required" or ""
	Type    string // value type for map fields; the group name for groups
	KeyType string // map fields only
	Name    string // for groups, the field name: the group name in lowercase
	Number  string
	Group   bool
	Options []*Option
}

type Oneof struct {
	Decl
	Name    string
	Fields  []*Field
	Options []*Option
}

type Enum struct {
	Decl
	Name    string
	Values  []*EnumValue
	Options []*Option
}

type EnumValue struct {
	Decl
	Name      string
	Number    string
	NumberPos Position
	Options   []*Option
}

type Service struct {
	Decl
	Name    string
	RPCs    []*RPC
	Options []*Option
}

type RPC struct {
	Decl
	Name     string
	Request  RPCType
	Response RPCType
	Options  []*Option
}

type RPCType struct {
	Type   string
	Stream bool
	Pos    Position // the type name, after any stream keyword
}

type Extend struct {
	Decl
	Type   string
	Fields []*Field
}

// Outline returns the parts of f another file can refer to: packages,
// top-level message and enum names, and the fields of top-level extends. It
// drops comments and bodies, so it is cheap to keep for every imported file.
func (f *File) Outline() *File {
	out := &File{Path: f.Path, Syntax: f.Syntax}
	for _, p := range f.Packages {
		out.Packages = append(out.Packages, &Package{Decl: Decl{Pos: p.Pos, End: p.End}, Name: p.Name})
	}
	for _, m := range f.Messages {
		out.Messages = append(out.Messages, &Message{Decl: Decl{Pos: m.Pos, End: m.End}, Name: m.Name})
	}
	for _, e := range f.Enums {
		out.Enums = append(out.Enums, &Enum{Decl: Decl{Pos: e.Pos, End: e.End}, Name: e.Name})
	}
	for _, e := range f.Extends {
		ext := &Extend{Decl: Decl{Pos: e.Pos, End: e.End}, Type: e.Type}
		for _, fld := range e.Fields {
			ext.Fields = append(ext.Fields, &Field{Decl: Decl{Pos: fld.Pos, End: fld.End}, Name: fld.Name, Type: fld.Type})
		}
		out.Extends = append(out.Extends, ext)
	}
	return out
}
