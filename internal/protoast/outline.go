package protoast

import (
	"github.com/easyp-tech/easyprotoc/parser"
	"github.com/easyp-tech/easyprotoc/protoc/syntax"
)

// OutlineOf returns what Outline returns, read straight from the syntax tree
// without building the full model or attaching comments. Declarations carry
// no positions.
func OutlineOf(f *syntax.File) *File {
	a := f.AST
	b := &builder{syn: f, ast: a}
	out := &File{Path: a.Path, Syntax: syntaxName(a.Syntax)}
	for _, ref := range a.Decls {
		switch ref.Kind {
		case parser.DeclSyntax:
			out.SyntaxSpecified = true
		case parser.DeclPackage:
			out.Packages = append(out.Packages, &Package{Name: b.tokens(a.Package)})
		case parser.DeclMessage:
			out.Messages = append(out.Messages, &Message{Name: f.Text(a.Messages[ref.Index].Name)})
		case parser.DeclEnum:
			out.Enums = append(out.Enums, &Enum{Name: f.Text(a.Enums[ref.Index].Name)})
		case parser.DeclExtend:
			n := &a.Extends[ref.Index]
			ext := &Extend{Type: b.tokens(n.ExtendeeSpan)}
			for _, idx := range b.children(n.FieldIndices) {
				fld := &a.Extensions[idx]
				ext.Fields = append(ext.Fields, &Field{Name: f.Text(fld.Name), Type: b.tokens(fld.TypeSpan)})
			}
			out.Extends = append(out.Extends, ext)
		}
	}
	return out
}
