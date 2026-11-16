package linker

import (
	"fmt"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

// ResolveSymbol performs lexical outward search to resolve a symbol reference.
// If name starts with a '.', it performs a fully-qualified lookup.
// Otherwise, it walks outward from current scope through enclosing scopes to the root.
func (t *SymbolTable) ResolveSymbol(callerFileID uint32, scope string, name string) (*Symbol, bool) {
	return t.resolveSymbol(callerFileID, scope, name, false)
}

// ResolveType resolves like ResolveSymbol in protoc's LOOKUP_TYPES mode: a
// single-component name that finds a non-type symbol keeps searching outward.
func (t *SymbolTable) ResolveType(callerFileID uint32, scope string, name string) (*Symbol, bool) {
	return t.resolveSymbol(callerFileID, scope, name, true)
}

func (t *SymbolTable) resolveSymbol(callerFileID uint32, scope string, name string, typesOnly bool) (*Symbol, bool) {
	sym, ok := t.lookupSymbol(callerFileID, scope, name, typesOnly)
	if ok && t.tracking && sym.Kind != SymPackage {
		t.uses[callerFileID][sym.FileID] = struct{}{}
	}
	return sym, ok
}

func (t *SymbolTable) lookupSymbol(callerFileID uint32, scope string, name string, typesOnly bool) (*Symbol, bool) {
	// 1. Fully-qualified reference
	if strings.HasPrefix(name, ".") {
		canonical := name[1:]
		if symID, ok := t.ByName[canonical]; ok {
			sym := &t.Symbols[symID]
			if sym.Kind == SymPackage || t.Vis.IsVisible(callerFileID, sym.FileID) {
				return sym, true
			}
		}
		return nil, false
	}

	// 2. Relative reference
	firstName := name
	rest := ""
	if dot := strings.IndexByte(name, '.'); dot >= 0 {
		firstName = name[:dot]
		rest = name[dot+1:]
	}

	currentScope := strings.TrimPrefix(scope, ".")
	for {
		candFirst := joinScope(currentScope, firstName)
		if symID, ok := t.ByName[candFirst]; ok {
			sym := &t.Symbols[symID]
			// protoc skips a first component that cannot contain nested names
			// (field, enum value, method, extension) and keeps searching outward.
			if rest != "" && !sym.Kind.isAggregate() {
				symID, ok = 0, false
			}
			if rest == "" && typesOnly && sym.Kind != SymMessage && sym.Kind != SymEnum {
				symID, ok = 0, false
			}
			if ok && (sym.Kind == SymPackage || t.Vis.IsVisible(callerFileID, sym.FileID)) {
				// First component resolved in this scope
				if rest == "" {
					return sym, true
				}
				// Verify full relative path
				candFull := joinScope(currentScope, name)
				if fullID, ok := t.ByName[candFull]; ok {
					fullSym := &t.Symbols[fullID]
					if fullSym.Kind == SymPackage || t.Vis.IsVisible(callerFileID, fullSym.FileID) {
						return fullSym, true
					}
				}
				// C++ shadowing rule: candidate prefix matched here, outer scopes shadowed
				return nil, false
			}
		}

		if currentScope == "" {
			break
		}
		if lastDot := strings.LastIndexByte(currentScope, '.'); lastDot >= 0 {
			currentScope = currentScope[:lastDot]
		} else {
			currentScope = ""
		}
	}

	return nil, false
}

// ResolveTypes iterates across all files in the compilation and resolves field types,
// service method input/output types, and extension extendees.
func (t *SymbolTable) ResolveTypes() {
	for i, f := range t.Files {
		t.ResolveFile(uint32(i), f)
	}
}

// ResolveFile resolves field types, service method input/output types, and extension extendees for a single file.
func (t *SymbolTable) ResolveFile(fileID uint32, f *parser.ASTFile) {
	pkg := f.Package.String(f.Src)

	// 1. Resolve top-level extensions
	for ei := range f.Extensions {
		ext := &f.Extensions[ei]
		extFullName := joinScope(pkg, ext.Name.String(f.Src))
		if extSymID, ok := t.ByName[extFullName]; ok {
			t.resolveExtension(fileID, f, ext, extSymID, pkg)
		}
	}

	// 2. Resolve messages (fields, nested extensions, nested messages)
	for mi := f.TopLevelMessages.Start; mi < f.TopLevelMessages.End(); mi++ {
		msgIdx := f.ChildIndices[mi]
		t.resolveMessage(fileID, f, msgIdx, pkg)
	}

	// 3. Resolve service methods
	for si := range f.Services {
		svc := &f.Services[si]
		svcFullName := joinScope(pkg, svc.Name.String(f.Src))
		for mi := svc.Methods.Start; mi < svc.Methods.End(); mi++ {
			methodIdx := f.ChildIndices[mi]
			method := &f.Methods[methodIdx]

			inStr := method.InputType.String(f.Src)
			if sym, ok := t.resolveMessageRef(fileID, f, svcFullName, inStr, sourceSpan(method.InputTypeSpan, method.InputType)); ok {
				method.InputType = f.InternString("." + sym.FullName)
			}

			outStr := method.OutputType.String(f.Src)
			if sym, ok := t.resolveMessageRef(fileID, f, svcFullName, outStr, sourceSpan(method.OutputTypeSpan, method.OutputType)); ok {
				method.OutputType = f.InternString("." + sym.FullName)
			}
		}
	}
}

func (t *SymbolTable) resolveExtension(fileID uint32, f *parser.ASTFile, ext *parser.FieldNode, extSymID uint32, scope string) {
	// Extendee resolution
	extendeeStr := ext.Extendee.String(f.Src)
	if sym, ok := t.resolveMessageRef(fileID, f, scope, extendeeStr, sourceSpan(ext.ExtendeeSpan, ext.Extendee)); ok {
		canonical := "." + sym.FullName
		ext.Extendee = f.InternString(canonical)
		t.Symbols[extSymID].Extendee = canonical
		t.Extensions[ExtKey{Extendee: canonical, Tag: ext.Number}] = extSymID
	} else {
		canonical := extendeeStr
		if !strings.HasPrefix(canonical, ".") {
			canonical = "." + canonical
		}
		ext.Extendee = f.InternString(canonical)
		t.Symbols[extSymID].Extendee = canonical
		t.Extensions[ExtKey{Extendee: canonical, Tag: ext.Number}] = extSymID
	}

	// Extension type resolution
	if ext.TypeKind == parser.TypeUnresolved || ext.TypeKind == parser.TypeMessage || ext.TypeKind == parser.TypeEnum {
		typeStr := ext.TypeName.String(f.Src)
		if sym, ok := t.resolveTypeRef(fileID, f, scope, typeStr, sourceSpan(ext.TypeSpan, ext.TypeName)); ok {
			if sym.Kind == SymEnum {
				ext.TypeKind = parser.TypeEnum
			} else {
				ext.TypeKind = parser.TypeMessage
			}
			canonical := "." + sym.FullName
			ext.TypeName = f.InternString(canonical)
			t.Symbols[extSymID].TypeKind = ext.TypeKind
			t.Symbols[extSymID].TypeName = canonical
		}
	}
}

func (t *SymbolTable) resolveMessage(fileID uint32, f *parser.ASTFile, msgIdx uint32, parentScope string) {
	m := &f.Messages[msgIdx]
	mFullName := joinScope(parentScope, m.Name.String(f.Src))
	canonicalMsg := "." + mFullName

	// Fields
	for fi := m.Fields.Start; fi < m.Fields.End(); fi++ {
		fldIdx := f.ChildIndices[fi]
		fld := &f.Fields[fldIdx]

		if fld.TypeKind == parser.TypeUnresolved || fld.TypeKind == parser.TypeMessage || fld.TypeKind == parser.TypeEnum {
			typeStr := fld.TypeName.String(f.Src)
			if sym, ok := t.resolveTypeRef(fileID, f, mFullName, typeStr, sourceSpan(fld.TypeSpan, fld.TypeName)); ok {
				if sym.Kind == SymEnum {
					fld.TypeKind = parser.TypeEnum
				} else {
					fld.TypeKind = parser.TypeMessage
				}
				canonical := "." + sym.FullName
				fld.TypeName = f.InternString(canonical)
			}

			// Update FieldInfo
			fldName := fld.Name.String(f.Src)
			if info, ok := t.MessageFields[canonicalMsg][fldName]; ok {
				info.TypeKind = fld.TypeKind
				info.TypeName = fld.TypeName.String(f.Src)
				t.MessageFields[canonicalMsg][fldName] = info
			}
		}
	}

	// Nested messages
	for nmi := m.NestedMessages.Start; nmi < m.NestedMessages.End(); nmi++ {
		subIdx := f.ChildIndices[nmi]
		t.resolveMessage(fileID, f, subIdx, mFullName)
	}

	// Nested extensions
	for exi := m.Extensions.Start; exi < m.Extensions.End(); exi++ {
		extIdx := f.ChildIndices[exi]
		ext := &f.Extensions[extIdx]
		extFullName := joinScope(mFullName, ext.Name.String(f.Src))
		if extSymID, ok := t.ByName[extFullName]; ok {
			t.resolveExtension(fileID, f, ext, extSymID, mFullName)
		}
	}
}

// resolveTypeRef resolves a field or extension type, reporting protoc's
// diagnostics when the name is undefined or does not name a message or enum.
func (t *SymbolTable) resolveTypeRef(fileID uint32, f *parser.ASTFile, scope, name string, src parser.TextSpan) (*Symbol, bool) {
	sym, ok := t.ResolveType(fileID, scope, name)
	if !ok {
		t.addError(f, src.Start, fmt.Sprintf("%q is not defined.", writtenName(f, src, name)))
		return nil, false
	}
	if sym.Kind != SymMessage && sym.Kind != SymEnum {
		t.addError(f, src.Start, fmt.Sprintf("%q is not a type.", writtenName(f, src, name)))
		return nil, false
	}
	return sym, true
}

// resolveMessageRef resolves an extendee or rpc input/output type, which must be a message.
func (t *SymbolTable) resolveMessageRef(fileID uint32, f *parser.ASTFile, scope, name string, src parser.TextSpan) (*Symbol, bool) {
	sym, ok := t.ResolveSymbol(fileID, scope, name)
	if !ok {
		t.addError(f, src.Start, fmt.Sprintf("%q is not defined.", writtenName(f, src, name)))
		return nil, false
	}
	if sym.Kind != SymMessage {
		t.addError(f, src.Start, fmt.Sprintf("%q is not a message type.", writtenName(f, src, name)))
		return nil, false
	}
	return sym, true
}

func (t *SymbolTable) addError(f *parser.ASTFile, offset uint32, msg string) {
	path := f.DisplayPath
	if path == "" {
		path = f.Path
	}
	var line, col int
	if offset > 0 {
		line, col = f.LineCol(offset)
	}
	f.Errors = append(f.Errors, parser.Diagnostic{Path: path, Line: line, Col: col, Message: msg})
}

// sourceSpan returns the first non-empty span, preferring the original source span.
func sourceSpan(spans ...parser.TextSpan) parser.TextSpan {
	for _, span := range spans {
		if !span.IsEmpty() {
			return span
		}
	}
	return parser.TextSpan{}
}

// writtenName returns the reference as written in the source, falling back to
// the stored name; protoc quotes the written form in its diagnostics.
func writtenName(f *parser.ASTFile, src parser.TextSpan, name string) string {
	if src.IsEmpty() || int(src.End) > len(f.Src) {
		return name
	}
	return strings.TrimSpace(src.String(f.Src))
}
