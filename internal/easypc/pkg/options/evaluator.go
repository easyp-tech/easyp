package options

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/easyp-tech/easyp/internal/easypc/pkg/linker"
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

type assignment struct {
	nameParts []parser.OptionNamePartNode
	val       *parser.OptionNode
	sym       *linker.Symbol
}

type tagGroup struct {
	tag         int32
	sym         *linker.Symbol
	assignments []assignment
}

// Evaluate evaluates all custom options across all files and attaches the resulting wire bytes
// to the AST nodes' CustomOptionsWire slices.
func Evaluate(table *linker.SymbolTable) {
	for i, f := range table.Files {
		EvaluateFile(table, uint32(i), f)
	}
}

// EvaluateFile evaluates all custom options for a single file.
func EvaluateFile(table *linker.SymbolTable, fileID uint32, f *parser.ASTFile) {
	pkg := f.Package.String(f.Src)

	// 1. File options
	f.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, pkg, ".google.protobuf.FileOptions", f.FileOptions)

	// 2. Top-level extensions
	for ei := range f.Extensions {
		ext := &f.Extensions[ei]
		ext.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, pkg, ".google.protobuf.FieldOptions", ext.Options)
	}

	// 3. Top-level enums
	for ei := f.TopLevelEnums.Start; ei < f.TopLevelEnums.End(); ei++ {
		enumIdx := f.ChildIndices[ei]
		evalEnumOptions(table, fileID, f, pkg, enumIdx)
	}

	// 4. Top-level messages
	for mi := f.TopLevelMessages.Start; mi < f.TopLevelMessages.End(); mi++ {
		msgIdx := f.ChildIndices[mi]
		evalMessageOptions(table, fileID, f, pkg, msgIdx)
	}

	// 5. Services and methods
	for si := range f.Services {
		svc := &f.Services[si]
		svcScope := joinScope(pkg, svc.Name.String(f.Src))
		svc.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, svcScope, ".google.protobuf.ServiceOptions", svc.Options)

		for mi := svc.Methods.Start; mi < svc.Methods.End(); mi++ {
			methodIdx := f.ChildIndices[mi]
			method := &f.Methods[methodIdx]
			method.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, svcScope, ".google.protobuf.MethodOptions", method.Options)
		}
	}
}

func evalEnumOptions(table *linker.SymbolTable, fileID uint32, f *parser.ASTFile, parentScope string, enumIdx uint32) {
	en := &f.Enums[enumIdx]
	enScope := joinScope(parentScope, en.Name.String(f.Src))
	en.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, enScope, ".google.protobuf.EnumOptions", en.Options)

	for vi := en.Values.Start; vi < en.Values.End(); vi++ {
		valIdx := f.ChildIndices[vi]
		ev := &f.EnumValues[valIdx]
		ev.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, enScope, ".google.protobuf.EnumValueOptions", ev.Options)
	}
}

func evalMessageOptions(table *linker.SymbolTable, fileID uint32, f *parser.ASTFile, parentScope string, msgIdx uint32) {
	m := &f.Messages[msgIdx]
	mScope := joinScope(parentScope, m.Name.String(f.Src))
	m.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, mScope, ".google.protobuf.MessageOptions", m.Options)

	// Fields
	for fi := m.Fields.Start; fi < m.Fields.End(); fi++ {
		fldIdx := f.ChildIndices[fi]
		fld := &f.Fields[fldIdx]
		fld.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, mScope, ".google.protobuf.FieldOptions", fld.Options)
	}

	// Oneofs
	for oi := m.Oneofs.Start; oi < m.Oneofs.End(); oi++ {
		oneofIdx := f.ChildIndices[oi]
		oo := &f.Oneofs[oneofIdx]
		oo.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, mScope, ".google.protobuf.OneofOptions", oo.Options)
	}

	// Extension ranges
	for eri := m.ExtensionRanges.Start; eri < m.ExtensionRanges.End(); eri++ {
		erIdx := f.ChildIndices[eri]
		er := &f.ExtensionRanges[erIdx]
		er.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, mScope, ".google.protobuf.ExtensionRangeOptions", er.Options)
	}

	// Nested extensions
	for exi := m.Extensions.Start; exi < m.Extensions.End(); exi++ {
		extIdx := f.ChildIndices[exi]
		ext := &f.Extensions[extIdx]
		ext.CustomOptionsWire = evaluateNodeOptions(table, fileID, f, mScope, ".google.protobuf.FieldOptions", ext.Options)
	}

	// Nested enums
	for nei := m.NestedEnums.Start; nei < m.NestedEnums.End(); nei++ {
		enumIdx := f.ChildIndices[nei]
		evalEnumOptions(table, fileID, f, mScope, enumIdx)
	}

	// Nested messages
	for nmi := m.NestedMessages.Start; nmi < m.NestedMessages.End(); nmi++ {
		subIdx := f.ChildIndices[nmi]
		evalMessageOptions(table, fileID, f, mScope, subIdx)
	}
}

func evaluateNodeOptions(table *linker.SymbolTable, fileID uint32, f *parser.ASTFile, scope string, targetExtendee string, opts parser.Span32) []byte {
	if opts.IsEmpty() {
		return nil
	}

	// Group root assignments by extension tag
	groups := make(map[int32]*tagGroup)
	tags := make([]int32, 0, 8)

	for i := opts.Start; i < opts.End(); i++ {
		optIdx := f.ChildIndices[i]
		opt := &f.Options[optIdx]
		if opt.NameParts.IsEmpty() {
			continue
		}

		firstPart := &f.OptionNameParts[opt.NameParts.Start]
		if !firstPart.IsExtension {
			checkStandardOption(f, targetExtendee, opt, firstPart.Name.String(f.Src))
			continue // Standard option, handled in wire serializer
		}

		extName := firstPart.Name.String(f.Src)
		sym, ok := table.ResolveSymbol(fileID, scope, extName)
		if !ok {
			optionError(f, optionNamePos(opt), fmt.Sprintf("Option \"(%s)\" unknown. Ensure that your proto definition file imports the proto which defines the option.", extName))
			continue
		}
		if sym.Kind != linker.SymExtension {
			continue
		}
		if opt.NameParts.Count == 1 {
			checkOptionValue(f, opt, sym.TypeKind, sym.FullName)
		}

		if sym.Extendee != targetExtendee {
			continue
		}

		parts := f.OptionNameParts[opt.NameParts.Start:opt.NameParts.End()]
		as := assignment{
			nameParts: parts,
			val:       opt,
			sym:       sym,
		}

		tag := sym.TagNumber
		grp, exists := groups[tag]
		if !exists {
			grp = &tagGroup{
				tag: tag,
				sym: sym,
			}
			groups[tag] = grp
			tags = append(tags, tag)
		}
		grp.assignments = append(grp.assignments, as)
	}

	if len(tags) == 0 {
		return nil
	}

	// Sort tags ascending
	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })

	var out []byte
	for _, tag := range tags {
		grp := groups[tag]
		if grp.sym.TypeKind == parser.TypeMessage {
			if grp.sym.Label == parser.LabelRepeated {
				for _, as := range grp.assignments {
					if as.val.ValKind == parser.OptionValList {
						for ei := as.val.SubOptions.Start; ei < as.val.SubOptions.End(); ei++ {
							elem := &f.Options[f.ChildIndices[ei]]
							if elem.ValKind == parser.OptionValAggregate {
								var elemAssignments []assignment
								for si := elem.SubOptions.Start; si < elem.SubOptions.End(); si++ {
									sub := &f.Options[f.ChildIndices[si]]
									subParts := f.OptionNameParts[sub.NameParts.Start:sub.NameParts.End()]
									elemAssignments = append(elemAssignments, assignment{nameParts: subParts, val: sub})
								}
								msgBytes := encodeMessage(table, fileID, f, scope, grp.sym.TypeName, elemAssignments)
								out = appendTag(out, uint32(tag), 2)
								out = appendVarint(out, uint64(len(msgBytes)))
								out = append(out, msgBytes...)
							}
						}
					} else if as.val.ValKind == parser.OptionValAggregate {
						var elemAssignments []assignment
						for si := as.val.SubOptions.Start; si < as.val.SubOptions.End(); si++ {
							sub := &f.Options[f.ChildIndices[si]]
							subParts := f.OptionNameParts[sub.NameParts.Start:sub.NameParts.End()]
							elemAssignments = append(elemAssignments, assignment{nameParts: subParts, val: sub})
						}
						msgBytes := encodeMessage(table, fileID, f, scope, grp.sym.TypeName, elemAssignments)
						out = appendTag(out, uint32(tag), 2)
						out = appendVarint(out, uint64(len(msgBytes)))
						out = append(out, msgBytes...)
					}
				}
			} else {
				// Message extension: collect all sub-assignments and encode message
				var subAssignments []assignment
				for _, as := range grp.assignments {
					if len(as.nameParts) > 1 {
						subAssignments = append(subAssignments, assignment{nameParts: as.nameParts[1:], val: as.val})
					} else if len(as.nameParts) == 1 {
						if as.val.ValKind == parser.OptionValAggregate {
							for si := as.val.SubOptions.Start; si < as.val.SubOptions.End(); si++ {
								sub := &f.Options[f.ChildIndices[si]]
								subParts := f.OptionNameParts[sub.NameParts.Start:sub.NameParts.End()]
								subAssignments = append(subAssignments, assignment{nameParts: subParts, val: sub})
							}
						}
					}
				}
				msgBytes := encodeMessage(table, fileID, f, scope, grp.sym.TypeName, subAssignments)
				out = appendTag(out, uint32(tag), 2)
				out = appendVarint(out, uint64(len(msgBytes)))
				out = append(out, msgBytes...)
			}
		} else {
			// Scalar / Enum extension
			for _, as := range grp.assignments {
				if as.val.ValKind == parser.OptionValList {
					for ei := as.val.SubOptions.Start; ei < as.val.SubOptions.End(); ei++ {
						elem := &f.Options[f.ChildIndices[ei]]
						encodeScalarOrEnum(&out, tag, grp.sym.TypeKind, grp.sym.TypeName, table, fileID, f, scope, elem)
					}
				} else {
					encodeScalarOrEnum(&out, tag, grp.sym.TypeKind, grp.sym.TypeName, table, fileID, f, scope, as.val)
				}
			}
		}
	}

	return out
}

func encodeMessage(table *linker.SymbolTable, fileID uint32, f *parser.ASTFile, scope string, msgTypeName string, assignments []assignment) []byte {
	if len(assignments) == 0 {
		return nil
	}

	// Group assignments by field tag within the message
	type fieldGroup struct {
		info        linker.FieldInfo
		assignments []assignment
	}

	groups := make(map[int32]*fieldGroup)
	tags := make([]int32, 0, 8)

	for _, as := range assignments {
		if len(as.nameParts) == 0 {
			continue
		}
		first := as.nameParts[0]
		var fldInfo linker.FieldInfo
		var found bool

		if !first.IsExtension {
			fldName := first.Name.String(f.Src)
			fldInfo, found = table.MessageFields[msgTypeName][fldName]
		} else {
			extName := first.Name.String(f.Src)
			extSym, ok := table.ResolveSymbol(fileID, scope, extName)
			if !ok {
				extSym, ok = table.ResolveSymbol(fileID, msgTypeName, extName)
			}
			if ok && extSym.Kind == linker.SymExtension {
				fldInfo = linker.FieldInfo{
					Name:      extName,
					TagNumber: extSym.TagNumber,
					TypeKind:  extSym.TypeKind,
					TypeName:  extSym.TypeName,
					Label:     extSym.Label,
				}
				found = true
			}
		}

		if !found {
			continue
		}

		tag := fldInfo.TagNumber
		grp, exists := groups[tag]
		if !exists {
			grp = &fieldGroup{info: fldInfo}
			groups[tag] = grp
			tags = append(tags, tag)
		}
		grp.assignments = append(grp.assignments, as)
	}

	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })

	var msgBuf []byte
	for _, tag := range tags {
		grp := groups[tag]
		if grp.info.TypeKind == parser.TypeMessage {
			if grp.info.Label == parser.LabelRepeated {
				for _, as := range grp.assignments {
					if as.val.ValKind == parser.OptionValList {
						// Repeated submessages: [ { ... }, { ... } ]
						for ei := as.val.SubOptions.Start; ei < as.val.SubOptions.End(); ei++ {
							elem := &f.Options[f.ChildIndices[ei]]
							if elem.ValKind == parser.OptionValAggregate {
								var elemAssignments []assignment
								for si := elem.SubOptions.Start; si < elem.SubOptions.End(); si++ {
									sub := &f.Options[f.ChildIndices[si]]
									subParts := f.OptionNameParts[sub.NameParts.Start:sub.NameParts.End()]
									elemAssignments = append(elemAssignments, assignment{nameParts: subParts, val: sub})
								}
								subBytes := encodeMessage(table, fileID, f, scope, grp.info.TypeName, elemAssignments)
								msgBuf = appendTag(msgBuf, uint32(tag), 2)
								msgBuf = appendVarint(msgBuf, uint64(len(subBytes)))
								msgBuf = append(msgBuf, subBytes...)
							}
						}
					} else if as.val.ValKind == parser.OptionValAggregate {
						var elemAssignments []assignment
						for si := as.val.SubOptions.Start; si < as.val.SubOptions.End(); si++ {
							sub := &f.Options[f.ChildIndices[si]]
							subParts := f.OptionNameParts[sub.NameParts.Start:sub.NameParts.End()]
							elemAssignments = append(elemAssignments, assignment{nameParts: subParts, val: sub})
						}
						subBytes := encodeMessage(table, fileID, f, scope, grp.info.TypeName, elemAssignments)
						msgBuf = appendTag(msgBuf, uint32(tag), 2)
						msgBuf = appendVarint(msgBuf, uint64(len(subBytes)))
						msgBuf = append(msgBuf, subBytes...)
					}
				}
			} else {
				// Singular submessage field: merge all sub-assignments and encode ONE message!
				var subAssignments []assignment
				for _, as := range grp.assignments {
					if len(as.nameParts) > 1 {
						subAssignments = append(subAssignments, assignment{nameParts: as.nameParts[1:], val: as.val})
					} else if as.val.ValKind == parser.OptionValAggregate {
						for si := as.val.SubOptions.Start; si < as.val.SubOptions.End(); si++ {
							sub := &f.Options[f.ChildIndices[si]]
							subParts := f.OptionNameParts[sub.NameParts.Start:sub.NameParts.End()]
							subAssignments = append(subAssignments, assignment{nameParts: subParts, val: sub})
						}
					}
				}
				subBytes := encodeMessage(table, fileID, f, scope, grp.info.TypeName, subAssignments)
				msgBuf = appendTag(msgBuf, uint32(tag), 2)
				msgBuf = appendVarint(msgBuf, uint64(len(subBytes)))
				msgBuf = append(msgBuf, subBytes...)
			}
		} else {
			// Scalar / Enum field
			for _, as := range grp.assignments {
				if as.val.ValKind == parser.OptionValList {
					for ei := as.val.SubOptions.Start; ei < as.val.SubOptions.End(); ei++ {
						elem := &f.Options[f.ChildIndices[ei]]
						encodeScalarOrEnum(&msgBuf, tag, grp.info.TypeKind, grp.info.TypeName, table, fileID, f, scope, elem)
					}
				} else {
					encodeScalarOrEnum(&msgBuf, tag, grp.info.TypeKind, grp.info.TypeName, table, fileID, f, scope, as.val)
				}
			}
		}
	}

	return msgBuf
}

func encodeScalarOrEnum(buf *[]byte, tag int32, typeKind parser.FieldTypeKind, typeName string, table *linker.SymbolTable, fileID uint32, f *parser.ASTFile, scope string, opt *parser.OptionNode) {
	switch typeKind {
	case parser.TypeBool:
		var b uint64
		if opt.ValKind == parser.OptionValIdent {
			s := opt.ValSpan.String(f.Src)
			if s == "true" || s == "TRUE" {
				b = 1
			}
		} else if opt.ValKind == parser.OptionValInt {
			if opt.IntVal != 0 {
				b = 1
			}
		}
		*buf = appendTag(*buf, uint32(tag), 0)
		*buf = appendVarint(*buf, b)

	case parser.TypeInt32, parser.TypeInt64, parser.TypeUint32, parser.TypeUint64:
		var v uint64
		if opt.ValKind == parser.OptionValInt {
			if opt.IsNegative {
				v = uint64(-int64(opt.IntVal))
			} else {
				v = opt.IntVal
			}
		}
		*buf = appendTag(*buf, uint32(tag), 0)
		*buf = appendVarint(*buf, v)

	case parser.TypeSint32:
		var v int32
		if opt.ValKind == parser.OptionValInt {
			v = int32(opt.IntVal)
			if opt.IsNegative {
				v = -v
			}
		}
		zz := uint32((v << 1) ^ (v >> 31))
		*buf = appendTag(*buf, uint32(tag), 0)
		*buf = appendVarint(*buf, uint64(zz))

	case parser.TypeSint64:
		var v int64
		if opt.ValKind == parser.OptionValInt {
			v = int64(opt.IntVal)
			if opt.IsNegative {
				v = -v
			}
		}
		zz := uint64((v << 1) ^ (v >> 63))
		*buf = appendTag(*buf, uint32(tag), 0)
		*buf = appendVarint(*buf, zz)

	case parser.TypeFixed32, parser.TypeSfixed32:
		var v uint32
		if opt.ValKind == parser.OptionValInt {
			if opt.IsNegative {
				v = uint32(-int64(opt.IntVal))
			} else {
				v = uint32(opt.IntVal)
			}
		}
		*buf = appendTag(*buf, uint32(tag), 5)
		*buf = appendFixed32(*buf, v)

	case parser.TypeFixed64, parser.TypeSfixed64:
		var v uint64
		if opt.ValKind == parser.OptionValInt {
			if opt.IsNegative {
				v = uint64(-int64(opt.IntVal))
			} else {
				v = opt.IntVal
			}
		}
		*buf = appendTag(*buf, uint32(tag), 1)
		*buf = appendFixed64(*buf, v)

	case parser.TypeFloat:
		var fval float32
		if opt.ValKind == parser.OptionValFloat {
			fval = float32(opt.FloatVal)
		} else if opt.ValKind == parser.OptionValInt {
			fval = float32(opt.IntVal)
			if opt.IsNegative {
				fval = -fval
			}
		} else if opt.ValKind == parser.OptionValIdent {
			s := opt.ValSpan.String(f.Src)
			switch strings.ToLower(s) {
			case "inf", "infinity":
				fval = float32(math.Inf(1))
			case "-inf", "-infinity":
				fval = float32(math.Inf(-1))
			case "nan":
				fval = float32(math.NaN())
			}
		}
		*buf = appendTag(*buf, uint32(tag), 5)
		*buf = appendFloat32(*buf, fval)

	case parser.TypeDouble:
		var dval float64
		if opt.ValKind == parser.OptionValFloat {
			dval = opt.FloatVal
		} else if opt.ValKind == parser.OptionValInt {
			dval = float64(opt.IntVal)
			if opt.IsNegative {
				dval = -dval
			}
		} else if opt.ValKind == parser.OptionValIdent {
			s := opt.ValSpan.String(f.Src)
			switch strings.ToLower(s) {
			case "inf", "infinity":
				dval = math.Inf(1)
			case "-inf", "-infinity":
				dval = math.Inf(-1)
			case "nan":
				dval = math.NaN()
			}
		}
		*buf = appendTag(*buf, uint32(tag), 1)
		*buf = appendFloat64(*buf, dval)

	case parser.TypeString, parser.TypeBytes:
		bytesVal := opt.ValSpan.Bytes(f.Src)
		*buf = appendBytesField(*buf, uint32(tag), bytesVal)

	case parser.TypeEnum:
		var enumNum int32
		if opt.ValKind == parser.OptionValInt {
			enumNum = int32(opt.IntVal)
			if opt.IsNegative {
				enumNum = -enumNum
			}
		} else if opt.ValKind == parser.OptionValIdent {
			ident := opt.ValSpan.String(f.Src)
			if num, ok := table.EnumValues[typeName][ident]; ok {
				enumNum = num
			} else if sym, ok := table.ResolveSymbol(fileID, scope, ident); ok && sym.Kind == linker.SymEnumValue {
				enumNum = sym.TagNumber
			}
		}
		*buf = appendTag(*buf, uint32(tag), 0)
		*buf = appendVarint(*buf, uint64(int64(enumNum)))
	}
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

// checkStandardOption reports a standard option name that the options message
// does not have, and a bool option set to anything but true or false.
func checkStandardOption(f *parser.ASTFile, target string, opt *parser.OptionNode, name string) {
	known, ok := standardOptions[target]
	if !ok {
		return
	}
	isBool, exists := known[name]
	if !exists {
		optionError(f, optionNamePos(opt), fmt.Sprintf("Option \"%s\" unknown.", name))
		return
	}
	statement := opt.Span.Start != opt.NameSpan.Start
	if !isBool || opt.NameParts.Count != 1 || (statement && name == "deprecated") {
		// The parser already checks "option deprecated = ...;" statements.
		return
	}
	checkOptionValue(f, opt, parser.TypeBool, strings.TrimPrefix(target, ".")+"."+name)
}

// checkOptionValue reports a scalar option value of the wrong kind, as protoc does.
func checkOptionValue(f *parser.ASTFile, opt *parser.OptionNode, kind parser.FieldTypeKind, fullName string) {
	value := opt.ValSpan.String(f.Src)
	var want, typeName string
	switch kind {
	case parser.TypeBool:
		if opt.ValKind == parser.OptionValIdent && (value == "true" || value == "false") {
			return
		}
		want, typeName = "identifier", "boolean"
	case parser.TypeInt32, parser.TypeInt64, parser.TypeUint32, parser.TypeUint64, parser.TypeSint32, parser.TypeSint64,
		parser.TypeFixed32, parser.TypeFixed64, parser.TypeSfixed32, parser.TypeSfixed64:
		if opt.ValKind == parser.OptionValInt {
			return
		}
		want, typeName = "integer", scalarTypeNames[kind]
	case parser.TypeFloat, parser.TypeDouble:
		if opt.ValKind == parser.OptionValInt || opt.ValKind == parser.OptionValFloat || opt.ValKind == parser.OptionValIdent {
			return
		}
		want, typeName = "number", scalarTypeNames[kind]
	case parser.TypeString, parser.TypeBytes:
		if opt.ValKind == parser.OptionValString {
			return
		}
		want, typeName = "quoted string", scalarTypeNames[kind]
	case parser.TypeEnum:
		if opt.ValKind == parser.OptionValIdent {
			return
		}
		want, typeName = "identifier", "enum-valued"
	default:
		return
	}
	pos := opt.ValRawSpan.Start
	if opt.ValRawSpan.IsEmpty() {
		pos = opt.Span.Start
	}
	optionError(f, pos, fmt.Sprintf("Value must be %s for %s option \"%s\".", want, typeName, fullName))
}

var scalarTypeNames = map[parser.FieldTypeKind]string{
	parser.TypeInt32: "int32", parser.TypeInt64: "int64", parser.TypeUint32: "uint32", parser.TypeUint64: "uint64",
	parser.TypeSint32: "sint32", parser.TypeSint64: "sint64", parser.TypeFixed32: "fixed32", parser.TypeFixed64: "fixed64",
	parser.TypeSfixed32: "sfixed32", parser.TypeSfixed64: "sfixed64", parser.TypeFloat: "float", parser.TypeDouble: "double",
	parser.TypeString: "string", parser.TypeBytes: "bytes",
}

func optionNamePos(opt *parser.OptionNode) uint32 {
	if opt.NameSpan.IsEmpty() {
		return opt.Span.Start
	}
	return opt.NameSpan.Start
}

func optionError(f *parser.ASTFile, offset uint32, msg string) {
	line, col := f.LineCol(offset)
	path := f.DisplayPath
	if path == "" {
		path = f.Path
	}
	f.Errors = append(f.Errors, parser.Diagnostic{Path: path, Line: line, Col: col, Message: msg, Semantic: true})
}
