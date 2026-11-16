package linker

import (
	"fmt"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

// ExtKey uniquely identifies an extension by its target extendee and tag number.
type ExtKey struct {
	Extendee string // Canonical name with leading dot, e.g. ".google.protobuf.FieldOptions"
	Tag      int32  // Field tag number, e.g. 1052
}

// SymbolTable maintains the global symbol registry and indexes across all compilation units.
type SymbolTable struct {
	Symbols    []Symbol
	ByName      map[string]uint32 // FullName (without leading dot) -> SymbolID
	Extensions  map[ExtKey]uint32 // (Extendee, Tag) -> SymbolID
	FileIDs     map[string]uint32 // CanonicalPath -> FileID
	Files       []*parser.ASTFile
	FileSymbols [][]uint32 // FileID -> slice of SymbolIDs defined in that file
	Vis         *Visibility

	// MessageFields: Canonical message name (with leading dot) -> field name -> FieldInfo
	MessageFields map[string]map[string]FieldInfo

	// EnumValues: Canonical enum name (with leading dot) -> value name -> int32
	EnumValues map[string]map[string]int32

	// uses records, per file, the files defining symbols it resolved while
	// tracking is on. Tracking is not safe for concurrent resolution.
	uses     []map[uint32]struct{}
	tracking bool
}

// TrackUses starts recording which files each file's resolved names come from.
// Call it before ResolveTypes and stop it before resolving concurrently.
func (t *SymbolTable) TrackUses() {
	if t.uses == nil {
		t.uses = make([]map[uint32]struct{}, len(t.Files))
		for i := range t.uses {
			t.uses[i] = make(map[uint32]struct{})
		}
	}
	t.tracking = true
}

// StopTrackingUses keeps the recorded uses and stops recording.
func (t *SymbolTable) StopTrackingUses() { t.tracking = false }

// UsedFiles returns the IDs of files that define symbols fileID resolved while
// tracking was on, excluding packages. The order is unspecified.
func (t *SymbolTable) UsedFiles(fileID uint32) []uint32 {
	if int(fileID) >= len(t.uses) {
		return nil
	}
	ids := make([]uint32, 0, len(t.uses[fileID]))
	for id := range t.uses[fileID] {
		ids = append(ids, id)
	}
	return ids
}

// NewSymbolTable constructs and indexes a SymbolTable from topologically sorted AST files.
func NewSymbolTable(files []*parser.ASTFile, fileMap map[string]*parser.ASTFile) *SymbolTable {
	numFiles := len(files)
	table := &SymbolTable{
		Symbols:       make([]Symbol, 0, 1024),
		ByName:        make(map[string]uint32, 1024),
		Extensions:    make(map[ExtKey]uint32, 256),
		FileIDs:       make(map[string]uint32, numFiles),
		Files:         files,
		FileSymbols:   make([][]uint32, numFiles),
		Vis:           NewVisibility(numFiles),
		MessageFields: make(map[string]map[string]FieldInfo, 512),
		EnumValues:    make(map[string]map[string]int32, 256),
	}

	// 1. Assign FileIDs
	for i, f := range files {
		table.FileIDs[f.Path] = uint32(i)
	}

	// 2. Build visibility bitset and propagate public imports
	publicBits := make([][]uint64, numFiles)
	for i := range publicBits {
		publicBits[i] = make([]uint64, table.Vis.NumWords())
	}

	for i, f := range files {
		callerID := uint32(i)
		table.Vis.SetVisible(callerID, callerID)

		// Direct imports
		for _, imp := range f.Imports {
			impPath := imp.Path.String(f.Src)
			if targetAST, ok := fileMap[impPath]; ok {
				if targetID, ok := table.FileIDs[targetAST.Path]; ok {
					table.Vis.SetVisible(callerID, targetID)
				}
			}
		}
	}

	// Propagate transitive public imports in topological order
	for i, f := range files {
		callerID := uint32(i)
		for _, imp := range f.Imports {
			impPath := imp.Path.String(f.Src)
			if targetAST, ok := fileMap[impPath]; ok {
				if targetID, ok := table.FileIDs[targetAST.Path]; ok {
					// Merge target's publicly imported files into caller's visibility
					table.Vis.MergeBits(callerID, publicBits[targetID])

					// If this import is public, re-export target and target's public imports
					if imp.Kind == parser.ImportPublic {
						wordIdx := int(targetID) / 64
						publicBits[callerID][wordIdx] |= uint64(1) << (targetID % 64)
						for w := 0; w < table.Vis.NumWords(); w++ {
							publicBits[callerID][w] |= publicBits[targetID][w]
						}
					}
				}
			}
		}
	}

	// 3. Index symbols across all files
	for i, f := range files {
		fileID := uint32(i)
		table.indexFileSymbols(fileID, f)
	}

	return table
}

// ReindexFile replaces a single file's AST in the SymbolTable and re-indexes its symbols in O(file_symbols).
func (t *SymbolTable) ReindexFile(fileID uint32, f *parser.ASTFile) {
	if int(fileID) < len(t.Files) {
		for _, id := range t.FileSymbols[fileID] {
			if int(id) < len(t.Symbols) && t.Symbols[id].Kind != SymPackage {
				if curID, ok := t.ByName[t.Symbols[id].FullName]; ok && curID == id {
					delete(t.ByName, t.Symbols[id].FullName)
				}
			}
		}
		t.FileSymbols[fileID] = t.FileSymbols[fileID][:0]
		t.Files[fileID] = f
	}
	t.indexFileSymbols(fileID, f)
}

func (t *SymbolTable) addSymbol(kind SymbolKind, fileID uint32, parentID uint32, name string, fullName string, nodeIdx uint32) uint32 {
	symID := uint32(len(t.Symbols))
	sym := Symbol{
		ID:        symID,
		Kind:      kind,
		FileID:    fileID,
		ParentID:  parentID,
		Name:      name,
		FullName:  fullName,
		NodeIndex: nodeIdx,
	}
	t.Symbols = append(t.Symbols, sym)
	if int(fileID) < len(t.FileSymbols) && kind != SymPackage {
		t.FileSymbols[fileID] = append(t.FileSymbols[fileID], symID)
	}
	// Do not overwrite earlier symbols of same name (first wins in same scope)
	if existingID, exists := t.ByName[fullName]; !exists {
		t.ByName[fullName] = symID
	} else if kind != SymPackage && t.Symbols[existingID].Kind != SymPackage {
		f := t.Files[fileID]
		offset := uint32(0)
		switch kind {
		case SymMessage:
			if int(nodeIdx) < len(f.Messages) {
				offset = f.Messages[nodeIdx].Name.Start
			}
		case SymEnum:
			if int(nodeIdx) < len(f.Enums) {
				offset = f.Enums[nodeIdx].Name.Start
			}
		case SymService:
			if int(nodeIdx) < len(f.Services) {
				offset = f.Services[nodeIdx].Name.Start
			}
		}
		var line, col int
		if offset > 0 {
			line, col = f.LineCol(offset)
		}
		dispPath := f.DisplayPath
		if dispPath == "" {
			dispPath = f.Path
		}
		f.Errors = append(f.Errors, parser.Diagnostic{
			Path:    dispPath,
			Line:    line,
			Col:     col,
			Message: fmt.Sprintf("%q is already defined.", name),
		})
	}
	return symID
}

func joinScope(parent, name string) string {
	if parent == "" {
		return name
	}
	if name == "" {
		return parent
	}
	return parent + "." + name
}

func (t *SymbolTable) indexFileSymbols(fileID uint32, f *parser.ASTFile) {
	pkg := f.Package.String(f.Src)
	if pkg != "" {
		// Index package and prefixes
		parts := strings.Split(pkg, ".")
		cur := ""
		for _, part := range parts {
			cur = joinScope(cur, part)
			if _, ok := t.ByName[cur]; !ok {
				t.addSymbol(SymPackage, fileID, NoParent, part, cur, 0)
			}
		}
	}

	// Top-level Enums
	for i := f.TopLevelEnums.Start; i < f.TopLevelEnums.End(); i++ {
		enumIdx := f.ChildIndices[i]
		t.indexEnum(fileID, f, enumIdx, pkg, NoParent)
	}

	// Top-level Messages
	for i := f.TopLevelMessages.Start; i < f.TopLevelMessages.End(); i++ {
		msgIdx := f.ChildIndices[i]
		t.indexMessage(fileID, f, msgIdx, pkg, NoParent)
	}

	// Top-level Extensions
	for ei := range f.Extensions {
		ext := &f.Extensions[ei]
		extName := ext.Name.String(f.Src)
		extFullName := joinScope(pkg, extName)
		symID := t.addSymbol(SymExtension, fileID, NoParent, extName, extFullName, uint32(ei))
		t.Symbols[symID].TagNumber = ext.Number
		t.Symbols[symID].TypeKind = ext.TypeKind
		t.Symbols[symID].TypeName = ext.TypeName.String(f.Src)
		t.Symbols[symID].Label = ext.Label
		t.Symbols[symID].Extendee = ext.Extendee.String(f.Src)
	}

	// Services & Methods
	for si := range f.Services {
		svc := &f.Services[si]
		svcName := svc.Name.String(f.Src)
		svcFullName := joinScope(pkg, svcName)
		svcSymID := t.addSymbol(SymService, fileID, NoParent, svcName, svcFullName, uint32(si))

		for mi := svc.Methods.Start; mi < svc.Methods.End(); mi++ {
			methodIdx := f.ChildIndices[mi]
			method := &f.Methods[methodIdx]
			mName := method.Name.String(f.Src)
			mFullName := joinScope(svcFullName, mName)
			t.addSymbol(SymMethod, fileID, svcSymID, mName, mFullName, methodIdx)
		}
	}
}

func (t *SymbolTable) indexEnum(fileID uint32, f *parser.ASTFile, enumIdx uint32, parentScope string, parentSymID uint32) {
	en := &f.Enums[enumIdx]
	enName := en.Name.String(f.Src)
	enFullName := joinScope(parentScope, enName)
	canonicalEnum := "." + enFullName
	symID := t.addSymbol(SymEnum, fileID, parentSymID, enName, enFullName, enumIdx)

	if _, ok := t.EnumValues[canonicalEnum]; !ok {
		t.EnumValues[canonicalEnum] = make(map[string]int32, en.Values.Count)
	}

	for vi := en.Values.Start; vi < en.Values.End(); vi++ {
		valIdx := f.ChildIndices[vi]
		ev := &f.EnumValues[valIdx]
		evName := ev.Name.String(f.Src)
		t.EnumValues[canonicalEnum][evName] = ev.Number

		// In protobuf C++ scoping, enum values are in the parent scope of the enum
		valFullName := joinScope(parentScope, evName)
		valSymID := t.addSymbol(SymEnumValue, fileID, symID, evName, valFullName, valIdx)
		t.Symbols[valSymID].TagNumber = ev.Number

		// Also register under enum name (e.g. MyEnum.VAL)
		scopedName := joinScope(enFullName, evName)
		if scopedName != valFullName {
			scopedSymID := t.addSymbol(SymEnumValue, fileID, symID, evName, scopedName, valIdx)
			t.Symbols[scopedSymID].TagNumber = ev.Number
		}
	}
}

func (t *SymbolTable) indexMessage(fileID uint32, f *parser.ASTFile, msgIdx uint32, parentScope string, parentSymID uint32) {
	m := &f.Messages[msgIdx]
	mName := m.Name.String(f.Src)
	mFullName := joinScope(parentScope, mName)
	canonicalMsg := "." + mFullName
	msgSymID := t.addSymbol(SymMessage, fileID, parentSymID, mName, mFullName, msgIdx)

	if _, ok := t.MessageFields[canonicalMsg]; !ok {
		t.MessageFields[canonicalMsg] = make(map[string]FieldInfo, m.Fields.Count)
	}

	// Index fields
	for fi := m.Fields.Start; fi < m.Fields.End(); fi++ {
		fldIdx := f.ChildIndices[fi]
		fld := &f.Fields[fldIdx]
		fldName := fld.Name.String(f.Src)
		fldFullName := joinScope(mFullName, fldName)
		fldSymID := t.addSymbol(SymField, fileID, msgSymID, fldName, fldFullName, fldIdx)
		t.Symbols[fldSymID].TagNumber = fld.Number
		t.Symbols[fldSymID].TypeKind = fld.TypeKind
		t.Symbols[fldSymID].TypeName = fld.TypeName.String(f.Src)
		t.Symbols[fldSymID].Label = fld.Label

		t.MessageFields[canonicalMsg][fldName] = FieldInfo{
			Name:      fldName,
			TagNumber: fld.Number,
			TypeKind:  fld.TypeKind,
			TypeName:  fld.TypeName.String(f.Src),
			Label:     fld.Label,
		}
	}

	// Index nested enums
	for nei := m.NestedEnums.Start; nei < m.NestedEnums.End(); nei++ {
		enumIdx := f.ChildIndices[nei]
		t.indexEnum(fileID, f, enumIdx, mFullName, msgSymID)
	}

	// Index nested messages
	for nmi := m.NestedMessages.Start; nmi < m.NestedMessages.End(); nmi++ {
		subMsgIdx := f.ChildIndices[nmi]
		t.indexMessage(fileID, f, subMsgIdx, mFullName, msgSymID)
	}

	// Index nested extensions
	for exi := m.Extensions.Start; exi < m.Extensions.End(); exi++ {
		extIdx := f.ChildIndices[exi]
		ext := &f.Extensions[extIdx]
		extName := ext.Name.String(f.Src)
		extFullName := joinScope(mFullName, extName)
		extSymID := t.addSymbol(SymExtension, fileID, msgSymID, extName, extFullName, extIdx)
		t.Symbols[extSymID].TagNumber = ext.Number
		t.Symbols[extSymID].TypeKind = ext.TypeKind
		t.Symbols[extSymID].TypeName = ext.TypeName.String(f.Src)
		t.Symbols[extSymID].Label = ext.Label
		t.Symbols[extSymID].Extendee = ext.Extendee.String(f.Src)
	}
}
