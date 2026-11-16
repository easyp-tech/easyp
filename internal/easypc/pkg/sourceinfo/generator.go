package sourceinfo

import (
	"sort"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/lexer"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

type generator struct {
	file      *parser.ASTFile
	fileID    uint32
	symTable  *linker.SymbolTable
	comments  *CommentTable
	locations []Location
}

func slicesClone(p []int32) []int32 {
	if len(p) == 0 {
		return nil
	}
	c := make([]int32, len(p))
	copy(c, p)
	return c
}

func (g *generator) addLoc(path []int32, span parser.TextSpan) {
	s, sLen := MakeSpan(g.file.Tokens, g.file.Src, span)
	g.locations = append(g.locations, Location{
		Path:    slicesClone(path),
		Span:    s,
		SpanLen: sLen,
	})
}

func (g *generator) addLocWithComments(path []int32, span parser.TextSpan, leadingStart, trailingEnd uint32) {
	lead, detached := g.comments.TakeLeading(leadingStart)
	trail := g.comments.TakeTrailing(trailingEnd)
	s, sLen := MakeSpan(g.file.Tokens, g.file.Src, span)
	g.locations = append(g.locations, Location{
		Path:                    slicesClone(path),
		Span:                    s,
		SpanLen:                 sLen,
		LeadingComments:         lead,
		TrailingComments:        trail,
		LeadingDetachedComments: detached,
	})
}

// Generate constructs the SourceCodeInfo locations for an ASTFile.
func Generate(f *parser.ASTFile, fileID uint32, symTable *linker.SymbolTable) *FileSourceInfo {
	if f == nil || f.Tokens == nil {
		return &FileSourceInfo{}
	}

	comments := BuildCommentTable(f.Tokens, f.Src)
	g := &generator{
		file:      f,
		fileID:    fileID,
		symTable:  symTable,
		comments:  comments,
		locations: make([]Location, 0, 128),
	}

	// 1. Root location path: []
	var rootStart, rootEnd uint32
	foundStart := false
	for i := 0; i < f.Tokens.Len(); i++ {
		k := f.Tokens.Kinds[i]
		if k != lexer.KindLineComment && k != lexer.KindBlockComment && k != lexer.KindError {
			rootStart = f.Tokens.Starts[i]
			foundStart = true
			break
		}
	}
	if !foundStart && f.Tokens.Len() > 0 {
		rootStart = f.Tokens.Starts[0]
	}
	foundEnd := false
	for i := f.Tokens.Len() - 1; i >= 0; i-- {
		k := f.Tokens.Kinds[i]
		if k != lexer.KindLineComment && k != lexer.KindBlockComment && k != lexer.KindError {
			rootEnd = f.Tokens.Starts[i] + f.Tokens.Lens[i]
			foundEnd = true
			break
		}
	}
	if !foundEnd && f.Tokens.Len() > 0 {
		rootEnd = f.Tokens.Starts[f.Tokens.Len()-1] + f.Tokens.Lens[f.Tokens.Len()-1]
	}
	rootSpan := parser.TextSpan{Start: rootStart, End: rootEnd}
	g.addLoc(nil, rootSpan)

	// 2. Syntax / Edition (path: [12])
	if !f.SyntaxSpan.IsEmpty() {
		g.addLocWithComments([]int32{12}, f.SyntaxSpan, f.SyntaxSpan.Start, f.SyntaxSpan.End)
	}

	// 3. Top-level declarations in original source order
	var depIdx, pubDepIdx, weakDepIdx int32
	var msgIdx, enumIdx, svcIdx, extIdx int32
	fileOptCounts := make(map[int32]int32)
	var fileUninterpIdx int32

	for _, decl := range f.Decls {
		switch decl.Kind {
		case parser.DeclPackage:
			if !f.PackageSpan.IsEmpty() {
				g.addLocWithComments([]int32{2}, f.PackageSpan, f.PackageSpan.Start, f.PackageSpan.End)
			}
		case parser.DeclImport:
			imp := &f.Imports[decl.Index]
			g.addLocWithComments([]int32{3, depIdx}, imp.Span, imp.Span.Start, imp.Span.End)
			depIdx++
			if imp.Kind == parser.ImportPublic {
				g.addLoc([]int32{10, pubDepIdx}, imp.ModifierSpan)
				pubDepIdx++
			} else if imp.Kind == parser.ImportWeak {
				g.addLoc([]int32{11, weakDepIdx}, imp.ModifierSpan)
				weakDepIdx++
			}
		case parser.DeclOption:
			opt := &f.Options[decl.Index]
			g.generateOption(opt, false, ".google.protobuf.FileOptions", []int32{8}, fileOptCounts, &fileUninterpIdx)
		case parser.DeclMessage:
			m := &f.Messages[decl.Index]
			if m.IsMapEntry || m.IsGroup {
				continue
			}
			g.generateMessage(m, nil, []int32{4, msgIdx})
			msgIdx++
		case parser.DeclEnum:
			en := &f.Enums[decl.Index]
			g.generateEnum(en, []int32{5, enumIdx})
			enumIdx++
		case parser.DeclService:
			svc := &f.Services[decl.Index]
			g.generateService(svc, []int32{6, svcIdx})
			svcIdx++
		case parser.DeclExtend:
			ext := &f.Extends[decl.Index]
			g.generateExtend(ext, &extIdx, []int32{7})
		}
	}

	return &FileSourceInfo{Locations: g.locations}
}

func (g *generator) generateMessage(m *parser.MessageNode, groupFieldPath []int32, msgPath []int32) {
	// Whole message block
	trailingEnd := m.OpenBraceSpan.End
	if trailingEnd == 0 {
		trailingEnd = m.Span.End
	}
	g.addLocWithComments(msgPath, m.Span, m.Span.Start, trailingEnd)

	// Message Name
	g.addLoc(append(slicesClone(msgPath), 1), m.Name)
	if groupFieldPath != nil {
		g.addLoc(append(slicesClone(groupFieldPath), 6), m.Name)
	}

	// Decls in source order
	var extRangeIdx, nestedExtIdx int32
	var resNameIdx, resRangeIdx int32
	msgOptCounts := make(map[int32]int32)
	var msgUninterpIdx int32

	for _, decl := range m.Decls {
		switch decl.Kind {
		case parser.DeclField:
			fld := &g.file.Fields[decl.Index]
			var localFldIdx int32 = -1
			for fi := m.Fields.Start; fi < m.Fields.End(); fi++ {
				if g.file.ChildIndices[fi] == decl.Index {
					localFldIdx = int32(fi - m.Fields.Start)
					break
				}
			}
			if localFldIdx == -1 {
				continue
			}
			fldPath := append(slicesClone(msgPath), 2, localFldIdx)
			if fld.TypeKind == parser.TypeGroup {
				// Find group message
				var groupMsg *parser.MessageNode
				var localGroupIdx int32 = -1
				for nmi := m.NestedMessages.Start; nmi < m.NestedMessages.End(); nmi++ {
					candidate := &g.file.Messages[g.file.ChildIndices[nmi]]
					if candidate.IsGroup && candidate.Name == fld.TypeName {
						groupMsg = candidate
						localGroupIdx = int32(nmi - m.NestedMessages.Start)
						break
					}
				}
				g.generateGroupField(fld, fldPath)
				if groupMsg != nil && localGroupIdx != -1 {
					subMsgPath := append(slicesClone(msgPath), 3, localGroupIdx)
					g.generateMessage(groupMsg, fldPath, subMsgPath)
				}
			} else {
				g.generateField(fld, fldPath)
			}
		case parser.DeclMessage:
			subMsg := &g.file.Messages[decl.Index]
			if subMsg.IsMapEntry || subMsg.IsGroup {
				continue
			}
			var localNestedIdx int32 = -1
			for nmi := m.NestedMessages.Start; nmi < m.NestedMessages.End(); nmi++ {
				if g.file.ChildIndices[nmi] == decl.Index {
					localNestedIdx = int32(nmi - m.NestedMessages.Start)
					break
				}
			}
			if localNestedIdx == -1 {
				continue
			}
			subMsgPath := append(slicesClone(msgPath), 3, localNestedIdx)
			g.generateMessage(subMsg, nil, subMsgPath)
		case parser.DeclEnum:
			subEnum := &g.file.Enums[decl.Index]
			var localEnumIdx int32 = -1
			for ei := m.NestedEnums.Start; ei < m.NestedEnums.End(); ei++ {
				if g.file.ChildIndices[ei] == decl.Index {
					localEnumIdx = int32(ei - m.NestedEnums.Start)
					break
				}
			}
			if localEnumIdx == -1 {
				continue
			}
			subEnumPath := append(slicesClone(msgPath), 4, localEnumIdx)
			g.generateEnum(subEnum, subEnumPath)
		case parser.DeclExtensionRange:
			erDecl := &g.file.ExtensionRangeDecls[decl.Index]
			g.generateExtensionRangeDecl(erDecl, &extRangeIdx, append(slicesClone(msgPath), 5))
		case parser.DeclExtend:
			ext := &g.file.Extends[decl.Index]
			g.generateExtend(ext, &nestedExtIdx, append(slicesClone(msgPath), 6))
		case parser.DeclOneof:
			oo := &g.file.Oneofs[decl.Index]
			if oo.IsSynthetic {
				continue
			}
			var localOneofIdx int32 = -1
			for oi := m.Oneofs.Start; oi < m.Oneofs.End(); oi++ {
				if g.file.ChildIndices[oi] == decl.Index {
					localOneofIdx = int32(oi - m.Oneofs.Start)
					break
				}
			}
			if localOneofIdx == -1 {
				continue
			}
			ooPath := append(slicesClone(msgPath), 8, localOneofIdx)
			g.generateOneof(oo, ooPath)
		case parser.DeclReserved:
			resDecl := &g.file.ReservedDecls[decl.Index]
			g.generateReserved(resDecl, &resNameIdx, &resRangeIdx, msgPath, 10, 9)
		case parser.DeclOption:
			opt := &g.file.Options[decl.Index]
			g.generateOption(opt, false, ".google.protobuf.MessageOptions", append(slicesClone(msgPath), 7), msgOptCounts, &msgUninterpIdx)
		}
	}
}

func (g *generator) generateField(fld *parser.FieldNode, fldPath []int32) {
	// 1. Whole field statement
	g.addLocWithComments(fldPath, fld.Span, fld.Span.Start, fld.Span.End)

	// 2. Extendee
	extSpan := fld.ExtendeeSpan
	if extSpan.IsEmpty() {
		extSpan = fld.Extendee
	}
	if !extSpan.IsEmpty() {
		g.addLoc(append(slicesClone(fldPath), 2), extSpan)
	}

	// 3. Label
	if !fld.LabelSpan.IsEmpty() {
		g.addLoc(append(slicesClone(fldPath), 4), fld.LabelSpan)
	}

	// 4. Type or TypeName
	typeSpan := fld.TypeSpan
	if typeSpan.IsEmpty() {
		typeSpan = fld.TypeName
	}
	if fld.TypeKind != parser.TypeUnresolved && fld.TypeKind != parser.TypeMessage && fld.TypeKind != parser.TypeEnum {
		g.addLoc(append(slicesClone(fldPath), 5), typeSpan)
	} else {
		g.addLoc(append(slicesClone(fldPath), 6), typeSpan)
	}

	// 5. Name
	g.addLoc(append(slicesClone(fldPath), 1), fld.Name)

	// 6. Number
	g.addLoc(append(slicesClone(fldPath), 3), fld.NumberSpan)

	// 7. Field options (options, default value, and json_name in source order)
	g.generateFieldOptions(fld, fldPath)
}

func (g *generator) generateFieldOptions(fld *parser.FieldNode, fldPath []int32) {
	if fld.OptionsSpan.IsEmpty() && fld.DefaultValSpan.IsEmpty() && fld.CustomJSONNameOpt.IsEmpty() {
		return
	}
	optsPath := append(slicesClone(fldPath), 8)
	if !fld.OptionsSpan.IsEmpty() {
		g.addLoc(optsPath, fld.OptionsSpan)
	}

	type optItemKind int
	const (
		optKindRegular optItemKind = iota
		optKindDefault
		optKindJSONName
	)
	type fieldOptItem struct {
		start uint32
		kind  optItemKind
		opt   *parser.OptionNode
	}

	var items []fieldOptItem
	if !fld.OptionsSpan.IsEmpty() {
		for oi := fld.Options.Start; oi < fld.Options.End(); oi++ {
			opt := &g.file.Options[g.file.ChildIndices[oi]]
			items = append(items, fieldOptItem{
				start: opt.Span.Start,
				kind:  optKindRegular,
				opt:   opt,
			})
		}
	}
	if !fld.DefaultValSpan.IsEmpty() {
		items = append(items, fieldOptItem{
			start: fld.DefaultValSpan.Start,
			kind:  optKindDefault,
		})
	}
	if !fld.CustomJSONNameOpt.IsEmpty() {
		items = append(items, fieldOptItem{
			start: fld.CustomJSONNameOpt.Start,
			kind:  optKindJSONName,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].start < items[j].start
	})

	fldOptCounts := make(map[int32]int32)
	var fldUninterpIdx int32

	for _, item := range items {
		switch item.kind {
		case optKindRegular:
			g.generateOption(item.opt, true, ".google.protobuf.FieldOptions", optsPath, fldOptCounts, &fldUninterpIdx)
		case optKindDefault:
			g.addLoc(append(slicesClone(fldPath), 7), fld.DefaultValSpan)
		case optKindJSONName:
			jsonPath := append(slicesClone(fldPath), 10)
			g.addLoc(jsonPath, fld.CustomJSONNameOpt)
			g.addLoc(jsonPath, fld.CustomJSONNameVal)
		}
	}
}

func (g *generator) generateGroupField(fld *parser.FieldNode, fldPath []int32) {
	g.addLoc(fldPath, fld.Span)
	extSpan := fld.ExtendeeSpan
	if extSpan.IsEmpty() {
		extSpan = fld.Extendee
	}
	if !extSpan.IsEmpty() {
		g.addLoc(append(slicesClone(fldPath), 2), extSpan)
	}
	if !fld.LabelSpan.IsEmpty() {
		g.addLoc(append(slicesClone(fldPath), 4), fld.LabelSpan)
	}
	g.addLoc(append(slicesClone(fldPath), 5), fld.GroupKeywordSpan)
	g.addLoc(append(slicesClone(fldPath), 1), fld.TypeName)
	g.addLoc(append(slicesClone(fldPath), 3), fld.NumberSpan)
}

func (g *generator) generateOneof(oo *parser.OneofNode, ooPath []int32) {
	trailingEnd := oo.OpenBraceSpan.End
	if trailingEnd == 0 {
		trailingEnd = oo.Span.End
	}
	g.addLocWithComments(ooPath, oo.Span, oo.Span.Start, trailingEnd)
	g.addLoc(append(slicesClone(ooPath), 1), oo.Name)

	ooOptCounts := make(map[int32]int32)
	var ooUninterpIdx int32
	for oi := oo.Options.Start; oi < oo.Options.End(); oi++ {
		opt := &g.file.Options[g.file.ChildIndices[oi]]
		g.generateOption(opt, false, ".google.protobuf.OneofOptions", append(slicesClone(ooPath), 2), ooOptCounts, &ooUninterpIdx)
	}
}

func (g *generator) generateEnum(en *parser.EnumNode, enumPath []int32) {
	trailingEnd := en.OpenBraceSpan.End
	if trailingEnd == 0 {
		trailingEnd = en.Span.End
	}
	g.addLocWithComments(enumPath, en.Span, en.Span.Start, trailingEnd)
	g.addLoc(append(slicesClone(enumPath), 1), en.Name)

	var valIdx, resNameIdx, resRangeIdx int32
	enumOptCounts := make(map[int32]int32)
	var enumUninterpIdx int32

	for _, decl := range en.Decls {
		switch decl.Kind {
		case parser.DeclEnumValue:
			val := &g.file.EnumValues[decl.Index]
			vPath := append(slicesClone(enumPath), 2, valIdx)
			valIdx++
			g.addLocWithComments(vPath, val.Span, val.Span.Start, val.Span.End)
			g.addLoc(append(slicesClone(vPath), 1), val.Name)
			g.addLoc(append(slicesClone(vPath), 2), val.NumberSpan)
			if !val.OptionsSpan.IsEmpty() {
				optsPath := append(slicesClone(vPath), 3)
				g.addLoc(optsPath, val.OptionsSpan)
				evOptCounts := make(map[int32]int32)
				var evUninterpIdx int32
				for oi := val.Options.Start; oi < val.Options.End(); oi++ {
					opt := &g.file.Options[g.file.ChildIndices[oi]]
					g.generateOption(opt, true, ".google.protobuf.EnumValueOptions", optsPath, evOptCounts, &evUninterpIdx)
				}
			}
		case parser.DeclOption:
			opt := &g.file.Options[decl.Index]
			g.generateOption(opt, false, ".google.protobuf.EnumOptions", append(slicesClone(enumPath), 3), enumOptCounts, &enumUninterpIdx)
		case parser.DeclReserved:
			resDecl := &g.file.ReservedDecls[decl.Index]
			g.generateReserved(resDecl, &resNameIdx, &resRangeIdx, enumPath, 5, 4)
		}
	}
}

func (g *generator) generateService(svc *parser.ServiceNode, svcPath []int32) {
	trailingEnd := svc.OpenBraceSpan.End
	if trailingEnd == 0 {
		trailingEnd = svc.Span.End
	}
	g.addLocWithComments(svcPath, svc.Span, svc.Span.Start, trailingEnd)
	g.addLoc(append(slicesClone(svcPath), 1), svc.Name)

	var rpcIdx int32
	svcOptCounts := make(map[int32]int32)
	var svcUninterpIdx int32

	for _, decl := range svc.Decls {
		switch decl.Kind {
		case parser.DeclMethod:
			m := &g.file.Methods[decl.Index]
			mPath := append(slicesClone(svcPath), 2, rpcIdx)
			rpcIdx++
			trailingEnd := m.Span.End
			if m.HasBody && !m.OpenBraceSpan.IsEmpty() {
				trailingEnd = m.OpenBraceSpan.End
			}
			g.addLocWithComments(mPath, m.Span, m.Span.Start, trailingEnd)
			g.addLoc(append(slicesClone(mPath), 1), m.Name)
			if m.ClientStreaming {
				g.addLoc(append(slicesClone(mPath), 5), m.ClientStreamingSpan)
			}
			inSpan := m.InputTypeSpan
			if inSpan.IsEmpty() {
				inSpan = m.InputType
			}
			g.addLoc(append(slicesClone(mPath), 2), inSpan)
			if m.ServerStreaming {
				g.addLoc(append(slicesClone(mPath), 6), m.ServerStreamingSpan)
			}
			outSpan := m.OutputTypeSpan
			if outSpan.IsEmpty() {
				outSpan = m.OutputType
			}
			g.addLoc(append(slicesClone(mPath), 3), outSpan)
			if m.Options.Count > 0 {
				optsPath := append(slicesClone(mPath), 4)
				mOptCounts := make(map[int32]int32)
				var mUninterpIdx int32
				for oi := m.Options.Start; oi < m.Options.End(); oi++ {
					opt := &g.file.Options[g.file.ChildIndices[oi]]
					g.generateOption(opt, false, ".google.protobuf.MethodOptions", optsPath, mOptCounts, &mUninterpIdx)
				}
			}
		case parser.DeclOption:
			opt := &g.file.Options[decl.Index]
			g.generateOption(opt, false, ".google.protobuf.ServiceOptions", append(slicesClone(svcPath), 3), svcOptCounts, &svcUninterpIdx)
		}
	}
}

func (g *generator) generateExtend(ext *parser.ExtendNode, extIdx *int32, extendPath []int32) {
	trailingEnd := ext.OpenBraceSpan.End
	if trailingEnd == 0 {
		trailingEnd = ext.Span.End
	}
	g.addLocWithComments(extendPath, ext.Span, ext.Span.Start, trailingEnd)

	for i := ext.FieldIndices.Start; i < ext.FieldIndices.End(); i++ {
		fld := &g.file.Extensions[g.file.ChildIndices[i]]
		fldPath := append(slicesClone(extendPath), *extIdx)
		*extIdx++
		g.addLocWithComments(fldPath, fld.Span, fld.Span.Start, fld.Span.End)
		extSpan := ext.ExtendeeSpan
		if extSpan.IsEmpty() {
			extSpan = ext.Extendee
		}
		g.addLoc(append(slicesClone(fldPath), 2), extSpan)
		if !fld.LabelSpan.IsEmpty() {
			g.addLoc(append(slicesClone(fldPath), 4), fld.LabelSpan)
		}
		typeSpan := fld.TypeSpan
		if typeSpan.IsEmpty() {
			typeSpan = fld.TypeName
		}
		if fld.TypeKind != parser.TypeUnresolved && fld.TypeKind != parser.TypeMessage && fld.TypeKind != parser.TypeEnum {
			g.addLoc(append(slicesClone(fldPath), 5), typeSpan)
		} else {
			g.addLoc(append(slicesClone(fldPath), 6), typeSpan)
		}
		g.addLoc(append(slicesClone(fldPath), 1), fld.Name)
		g.addLoc(append(slicesClone(fldPath), 3), fld.NumberSpan)
		g.generateFieldOptions(fld, fldPath)
	}
}

func (g *generator) generateExtensionRangeDecl(erDecl *parser.ExtensionRangeDeclNode, extRangeIdx *int32, path []int32) {
	g.addLocWithComments(path, erDecl.Span, erDecl.Span.Start, erDecl.Span.End)

	for i := erDecl.RangeIndices.Start; i < erDecl.RangeIndices.End(); i++ {
		er := &g.file.ExtensionRanges[g.file.ChildIndices[i]]
		rPath := append(slicesClone(path), *extRangeIdx)
		*extRangeIdx++
		g.addLoc(rPath, er.Span)
		g.addLoc(append(slicesClone(rPath), 1), er.StartSpan)
		endSpan := er.EndSpan
		if endSpan.IsEmpty() {
			endSpan = er.StartSpan
		}
		g.addLoc(append(slicesClone(rPath), 2), endSpan)
	}
}

func (g *generator) generateReserved(resDecl *parser.ReservedDeclNode, resNameIdx, resRangeIdx *int32, parentPath []int32, nameTag, rangeTag int32) {
	if resDecl.IsNames {
		resPath := append(slicesClone(parentPath), nameTag)
		g.addLocWithComments(resPath, resDecl.Span, resDecl.Span.Start, resDecl.Span.End)
		for i := resDecl.ItemIndices.Start; i < resDecl.ItemIndices.End(); i++ {
			nameIdx := g.file.ChildIndices[i]
			rawSpan := g.file.ReservedNameSpans[nameIdx]
			g.addLoc(append(slicesClone(resPath), *resNameIdx), rawSpan)
			*resNameIdx++
		}
	} else {
		resPath := append(slicesClone(parentPath), rangeTag)
		g.addLocWithComments(resPath, resDecl.Span, resDecl.Span.Start, resDecl.Span.End)
		for i := resDecl.ItemIndices.Start; i < resDecl.ItemIndices.End(); i++ {
			rr := &g.file.ReservedRanges[g.file.ChildIndices[i]]
			rPath := append(slicesClone(resPath), *resRangeIdx)
			*resRangeIdx++
			g.addLoc(rPath, rr.Span)
			g.addLoc(append(slicesClone(rPath), 1), rr.StartSpan)
			endSpan := rr.EndSpan
			if endSpan.IsEmpty() {
				endSpan = rr.StartSpan
			}
			g.addLoc(append(slicesClone(rPath), 2), endSpan)
		}
	}
}

func (g *generator) generateOption(opt *parser.OptionNode, compact bool, targetType string, parentPath []int32, optCounts map[int32]int32, uninterpIdx *int32) {
	if !compact {
		g.addLoc(parentPath, opt.Span)
	}

	if opt.NameParts.IsEmpty() {
		return
	}

	firstPart := &g.file.OptionNameParts[opt.NameParts.Start]
	var tag int32
	var isRep bool
	var subTags []int32

	if !firstPart.IsExtension {
		optName := firstPart.Name.String(g.file.Src)
		tag = lookupStandardOptionTag(targetType, optName)
		if targetType == ".google.protobuf.FieldOptions" && (tag == 19 || tag == 20) {
			isRep = true
		}
	} else {
		extName := firstPart.Name.String(g.file.Src)
		pkg := g.file.Package.String(g.file.Src)
		sym, ok := g.symTable.ResolveSymbol(g.fileID, pkg, extName)
		if ok && sym.Kind == linker.SymExtension {
			tag = sym.TagNumber
			isRep = (sym.Label == parser.LabelRepeated)
			curType := sym.TypeName
			for i := opt.NameParts.Start + 1; i < opt.NameParts.End(); i++ {
				subName := g.file.OptionNameParts[i].Name.String(g.file.Src)
				subTag, nextType := g.lookupFieldTag(curType, subName)
				if subTag != 0 {
					subTags = append(subTags, subTag)
					curType = nextType
				} else {
					break
				}
			}
		}
	}

	var optPath []int32
	if tag != 0 {
		count := optCounts[tag]
		optCounts[tag]++
		optPath = append(slicesClone(parentPath), tag)
		if isRep {
			optPath = append(optPath, count)
		}
		for _, st := range subTags {
			optPath = append(optPath, st)
		}
	} else {
		idx := *uninterpIdx
		*uninterpIdx++
		optPath = append(slicesClone(parentPath), 999, idx)
	}

	if len(optPath) > 0 {
		if !compact {
			g.addLocWithComments(optPath, opt.Span, opt.Span.Start, opt.Span.End)
		} else {
			g.addLoc(optPath, opt.Span)
		}
	}
}

func (g *generator) lookupFieldTag(msgTypeName, fieldName string) (int32, string) {
	if g.symTable == nil {
		return 0, ""
	}
	canonMsg := msgTypeName
	if !strings.HasPrefix(canonMsg, ".") {
		canonMsg = "." + canonMsg
	}
	if fields, ok := g.symTable.MessageFields[canonMsg]; ok {
		if fld, ok := fields[fieldName]; ok {
			return fld.TagNumber, fld.TypeName
		}
	}
	symIdx, ok := g.symTable.ByName[strings.TrimPrefix(canonMsg, ".")]
	if !ok {
		return 0, ""
	}
	sym := &g.symTable.Symbols[symIdx]
	if sym.Kind != linker.SymMessage {
		return 0, ""
	}
	f := g.symTable.Files[sym.FileID]
	m := &f.Messages[sym.NodeIndex]
	for i := m.Fields.Start; i < m.Fields.End(); i++ {
		fld := &f.Fields[f.ChildIndices[i]]
		if fld.Name.String(f.Src) == fieldName {
			return fld.Number, fld.TypeName.String(f.Src)
		}
	}
	return 0, ""
}

func lookupStandardOptionTag(targetType, optName string) int32 {
	switch targetType {
	case ".google.protobuf.FileOptions":
		switch optName {
		case "java_package":
			return 1
		case "java_outer_classname":
			return 8
		case "optimize_for":
			return 9
		case "java_multiple_files":
			return 10
		case "go_package":
			return 11
		case "cc_generic_services":
			return 16
		case "java_generic_services":
			return 17
		case "py_generic_services":
			return 18
		case "java_generate_equals_and_hash":
			return 20
		case "deprecated":
			return 23
		case "java_string_check_utf8":
			return 27
		case "cc_enable_arenas":
			return 31
		case "objc_class_prefix":
			return 36
		case "csharp_namespace":
			return 37
		case "swift_prefix":
			return 39
		case "php_class_prefix":
			return 40
		case "php_namespace":
			return 41
		case "php_metadata_namespace":
			return 44
		case "ruby_package":
			return 45
		case "features":
			return 50
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.MessageOptions":
		switch optName {
		case "message_set_wire_format":
			return 1
		case "no_standard_descriptor_accessor":
			return 2
		case "deprecated":
			return 3
		case "map_entry":
			return 7
		case "deprecated_legacy_json_field_conflicts":
			return 11
		case "features":
			return 12
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.FieldOptions":
		switch optName {
		case "ctype":
			return 1
		case "packed":
			return 2
		case "deprecated":
			return 3
		case "lazy":
			return 5
		case "jstype":
			return 6
		case "weak":
			return 10
		case "unverified_lazy":
			return 15
		case "debug_redact":
			return 16
		case "retention":
			return 17
		case "targets":
			return 19
		case "edition_defaults":
			return 20
		case "features":
			return 21
		case "feature_support":
			return 22
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.EnumOptions":
		switch optName {
		case "allow_alias":
			return 2
		case "deprecated":
			return 3
		case "deprecated_legacy_json_field_conflicts":
			return 6
		case "features":
			return 7
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.EnumValueOptions":
		switch optName {
		case "deprecated":
			return 1
		case "features":
			return 2
		case "debug_redact":
			return 3
		case "feature_support":
			return 4
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.ServiceOptions":
		switch optName {
		case "deprecated":
			return 33
		case "features":
			return 34
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.MethodOptions":
		switch optName {
		case "deprecated":
			return 33
		case "idempotency_level":
			return 34
		case "features":
			return 35
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.OneofOptions":
		switch optName {
		case "features":
			return 1
		case "uninterpreted_option":
			return 999
		}
	case ".google.protobuf.ExtensionRangeOptions":
		switch optName {
		case "declaration":
			return 2
		case "verification":
			return 3
		case "features":
			return 50
		case "uninterpreted_option":
			return 999
		}
	}
	return 0
}
