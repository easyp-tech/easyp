package linker

import (
	"fmt"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

// Validate reports the semantic errors protoc reports after names resolve:
// proto3 restrictions, use of reserved field numbers and names, default JSON
// name conflicts, and extension numbers outside the extendee's ranges or used
// twice. Errors go to each file's Errors.
func (t *SymbolTable) Validate() {
	for _, f := range t.Files {
		validateFile(f)
	}
	t.validateExtensions()
}

func validateFile(f *parser.ASTFile) {
	proto3 := f.Syntax == parser.SyntaxProto3
	for mi := range f.Messages {
		m := &f.Messages[mi]
		if m.IsMapEntry {
			continue
		}
		fields := children(f, m.Fields)
		for _, idx := range fields {
			fld := &f.Fields[idx]
			if proto3 {
				switch {
				case !fld.DefaultOpt.IsEmpty() || !fld.DefaultVal.IsEmpty():
					pos := fld.DefaultOpt.Start
					if fld.DefaultOpt.IsEmpty() {
						pos = fld.Span.Start
					}
					semanticError(f, pos, "Explicit default values are not allowed in proto3.")
				case fld.Label == parser.LabelRequired:
					semanticError(f, fld.LabelSpan.Start, "Required fields are not allowed in proto3.")
				case fld.TypeKind == parser.TypeGroup:
					semanticError(f, fld.GroupKeywordSpan.Start, "Groups are not supported in proto3 syntax.")
				}
			}
			for _, ri := range children(f, m.ReservedRanges) {
				r := f.ReservedRanges[ri]
				if fld.Number >= r.Start && fld.Number < r.End {
					semanticError(f, fld.NumberSpan.Start, fmt.Sprintf("Field %q uses reserved number %d.", fld.Name.String(f.Src), fld.Number))
				}
			}
			for _, ni := range children(f, m.ReservedNames) {
				if f.ReservedNames[ni].String(f.Src) == fld.Name.String(f.Src) {
					semanticError(f, fld.Name.Start, fmt.Sprintf("Field name %q is reserved.", fld.Name.String(f.Src)))
				}
			}
		}
		if proto3 {
			for _, decl := range m.Decls {
				if decl.Kind == parser.DeclExtensionRange {
					semanticError(f, f.ExtensionRangeDecls[decl.Index].Span.Start, "Extension ranges are not allowed in proto3.")
					break
				}
			}
			validateJSONNames(f, fields)
		}
	}
	if proto3 {
		for ei := range f.Enums {
			values := children(f, f.Enums[ei].Values)
			if len(values) > 0 && f.EnumValues[values[0]].Number != 0 {
				semanticError(f, f.EnumValues[values[0]].NumberSpan.Start, "The first enum value must be zero for open enums.")
			}
		}
	}
}

// validateJSONNames reports fields whose default JSON names collide.
func validateJSONNames(f *parser.ASTFile, fields []uint32) {
	seen := make(map[string]string, len(fields))
	for _, idx := range fields {
		fld := &f.Fields[idx]
		name := fld.Name.String(f.Src)
		jsonName := parser.JSONName(name)
		if first, ok := seen[jsonName]; ok && first != name {
			semanticError(f, fld.Name.Start, fmt.Sprintf("The default JSON name of field %q (%q) conflicts with the default JSON name of field %q.", name, jsonName, first))
			continue
		}
		seen[jsonName] = name
	}
}

// validateExtensions checks every extension number against the extendee's
// extension ranges and against earlier extensions of the same extendee.
func (t *SymbolTable) validateExtensions() {
	type extRef struct{ file, node uint32 }
	names := make(map[extRef]string)
	for _, sym := range t.Symbols {
		if sym.Kind == SymExtension {
			names[extRef{sym.FileID, sym.NodeIndex}] = sym.FullName
		}
	}
	used := make(map[ExtKey]string)
	for fileID, f := range t.Files {
		for ei := range f.Extensions {
			ext := &f.Extensions[ei]
			extendee := ext.Extendee.String(f.Src)
			symID, ok := t.ByName[strings.TrimPrefix(extendee, ".")]
			if !ok || t.Symbols[symID].Kind != SymMessage {
				continue
			}
			target := t.Symbols[symID]
			tf := t.Files[target.FileID]
			declared := false
			for _, ri := range children(tf, tf.Messages[target.NodeIndex].ExtensionRanges) {
				r := tf.ExtensionRanges[ri]
				if ext.Number >= r.Start && ext.Number < r.End {
					declared = true
					break
				}
			}
			if !declared {
				semanticError(f, ext.NumberSpan.Start, fmt.Sprintf("%q does not declare %d as an extension number.", target.FullName, ext.Number))
				continue
			}
			key := ExtKey{Extendee: extendee, Tag: ext.Number}
			if first, ok := used[key]; ok {
				semanticError(f, ext.NumberSpan.Start, fmt.Sprintf("Extension number %d has already been used in %q by extension %q.", ext.Number, target.FullName, first))
				continue
			}
			used[key] = names[extRef{uint32(fileID), uint32(ei)}]
		}
	}
}

func children(f *parser.ASTFile, s parser.Span32) []uint32 {
	if s.IsEmpty() {
		return nil
	}
	return f.ChildIndices[s.Start:s.End()]
}

func semanticError(f *parser.ASTFile, offset uint32, msg string) {
	line, col := f.LineCol(offset)
	path := f.DisplayPath
	if path == "" {
		path = f.Path
	}
	f.Errors = append(f.Errors, parser.Diagnostic{Path: path, Line: line, Col: col, Message: msg, Semantic: true})
}
