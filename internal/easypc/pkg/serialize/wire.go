package serialize

import (
	"github.com/easyp-tech/easyp/internal/easypc/pkg/parser"
)

type WireSerializer struct {
	fileSizes []uint32
	msgSizes  [][]uint32 // [fileIdx][msgIdx]
	enumSizes [][]uint32 // [fileIdx][enumIdx]
}

func NewWireSerializer() *WireSerializer {
	return &WireSerializer{
		fileSizes: make([]uint32, 0, 16),
		msgSizes:  make([][]uint32, 0, 16),
		enumSizes: make([][]uint32, 0, 16),
	}
}

func (s *WireSerializer) reset(totalFiles int) {
	s.fileSizes = s.fileSizes[:0]
	if cap(s.msgSizes) < totalFiles {
		s.msgSizes = make([][]uint32, totalFiles)
		s.enumSizes = make([][]uint32, totalFiles)
	} else {
		s.msgSizes = s.msgSizes[:totalFiles]
		s.enumSizes = s.enumSizes[:totalFiles]
	}
	for i := 0; i < totalFiles; i++ {
		s.msgSizes[i] = s.msgSizes[i][:0]
		s.enumSizes[i] = s.enumSizes[i][:0]
	}
}

// Low-level varint primitives
func varintLen(v uint64) int {
	switch {
	case v < (1 << 7):
		return 1
	case v < (1 << 14):
		return 2
	case v < (1 << 21):
		return 3
	case v < (1 << 28):
		return 4
	case v < (1 << 35):
		return 5
	case v < (1 << 42):
		return 6
	case v < (1 << 49):
		return 7
	case v < (1 << 56):
		return 8
	case v < (1 << 63):
		return 9
	default:
		return 10
	}
}

func writeVarint(buf []byte, pos int, v uint64) int {
	for v >= 0x80 {
		buf[pos] = byte(v) | 0x80
		v >>= 7
		pos++
	}
	buf[pos] = byte(v)
	return pos + 1
}

func tagLen(fieldNum uint32) int {
	return varintLen(uint64(fieldNum << 3))
}

func writeTag(buf []byte, pos int, fieldNum uint32, wireType uint8) int {
	return writeVarint(buf, pos, uint64(fieldNum<<3)|uint64(wireType))
}

func stringFieldLen(fieldNum uint32, s string) int {
	if len(s) == 0 {
		return 0
	}
	return tagLen(fieldNum) + varintLen(uint64(len(s))) + len(s)
}

func bytesFieldLen(fieldNum uint32, b []byte) int {
	if len(b) == 0 {
		return 0
	}
	return tagLen(fieldNum) + varintLen(uint64(len(b))) + len(b)
}

func int32FieldLen(fieldNum uint32, v int32) int {
	return tagLen(fieldNum) + varintLen(uint64(v))
}

func boolFieldLen(fieldNum uint32, v bool) int {
	if !v {
		return 0
	}
	return tagLen(fieldNum) + 1
}

func explicitBoolFieldLen(fieldNum uint32) int {
	return tagLen(fieldNum) + 1
}

func writeStringField(buf []byte, pos int, fieldNum uint32, s string) int {
	if len(s) == 0 {
		return pos
	}
	pos = writeTag(buf, pos, fieldNum, 2)
	pos = writeVarint(buf, pos, uint64(len(s)))
	copy(buf[pos:], s)
	return pos + len(s)
}

func writeBytesField(buf []byte, pos int, fieldNum uint32, b []byte) int {
	if len(b) == 0 {
		return pos
	}
	pos = writeTag(buf, pos, fieldNum, 2)
	pos = writeVarint(buf, pos, uint64(len(b)))
	copy(buf[pos:], b)
	return pos + len(b)
}

func writeInt32Field(buf []byte, pos int, fieldNum uint32, v int32) int {
	pos = writeTag(buf, pos, fieldNum, 0)
	return writeVarint(buf, pos, uint64(v))
}

func writeBoolField(buf []byte, pos int, fieldNum uint32, v bool) int {
	if !v {
		return pos
	}
	pos = writeTag(buf, pos, fieldNum, 0)
	buf[pos] = 1
	return pos + 1
}

func writeExplicitBoolField(buf []byte, pos int, fieldNum uint32, v bool) int {
	pos = writeTag(buf, pos, fieldNum, 0)
	if v {
		buf[pos] = 1
	} else {
		buf[pos] = 0
	}
	return pos + 1
}

// SerializeFileDescriptorSet encodes files into a single buffer matching FileDescriptorSet.
func (s *WireSerializer) SerializeFileDescriptorSet(files []*parser.ASTFile, includeSourceInfo bool) []byte {
	s.reset(len(files))

	// Pass 1: Compute exact sizes
	totalSetSize := 0
	for fileIdx, f := range files {
		fSize := s.computeFileSize(fileIdx, f, includeSourceInfo)
		s.fileSizes = append(s.fileSizes, fSize)
		// Field 1 of FileDescriptorSet: repeated FileDescriptorProto file = 1; (wire type 2)
		totalSetSize += tagLen(1) + varintLen(uint64(fSize)) + int(fSize)
	}

	// Pass 2: Allocate exactly one buffer and serialize
	buf := make([]byte, totalSetSize)
	pos := 0

	for fileIdx, f := range files {
		pos = writeTag(buf, pos, 1, 2)
		pos = writeVarint(buf, pos, uint64(s.fileSizes[fileIdx]))
		pos = s.writeFileDescriptorProto(buf, pos, fileIdx, f, includeSourceInfo)
	}

	return buf
}

// Pass 1: Size computations
func (s *WireSerializer) computeFileSize(fileIdx int, f *parser.ASTFile, includeSourceInfo bool) uint32 {
	var total int

	// Tag 1: name
	if f.Path != "" {
		total += stringFieldLen(1, f.Path)
	}
	// Tag 2: package
	if !f.Package.IsEmpty() {
		total += bytesFieldLen(2, f.Package.Bytes(f.Src))
	}
	// Tag 3: dependency
	for _, imp := range f.Imports {
		total += bytesFieldLen(3, imp.Path.Bytes(f.Src))
	}
	// Tag 4: message_type
	for i := f.TopLevelMessages.Start; i < f.TopLevelMessages.End(); i++ {
		mi := f.ChildIndices[i]
		mSize := s.computeMessageSize(fileIdx, f, mi)
		total += tagLen(4) + varintLen(uint64(mSize)) + int(mSize)
	}
	// Tag 5: enum_type
	for i := f.TopLevelEnums.Start; i < f.TopLevelEnums.End(); i++ {
		ei := f.ChildIndices[i]
		eSize := s.computeEnumSize(fileIdx, f, ei)
		total += tagLen(5) + varintLen(uint64(eSize)) + int(eSize)
	}
	// Tag 6: service
	for si := range f.Services {
		sSize := s.computeServiceSize(f, uint32(si))
		total += tagLen(6) + varintLen(uint64(sSize)) + int(sSize)
	}
	// Tag 7: extension
	for xi := range f.Extensions {
		fldSize := s.computeFieldSize(f, &f.Extensions[xi])
		total += tagLen(7) + varintLen(uint64(fldSize)) + int(fldSize)
	}
	// Tag 8: options (FileOptions)
	optSize := s.computeFileOptionsSize(f)
	if optSize > 0 {
		total += tagLen(8) + varintLen(uint64(optSize)) + int(optSize)
	}
	// Tag 9: source_code_info
	if includeSourceInfo && len(f.SourceInfoWire) > 0 {
		total += bytesFieldLen(9, f.SourceInfoWire)
	}
	// Tag 10: public_dependency
	for i, imp := range f.Imports {
		if imp.Kind == parser.ImportPublic {
			total += int32FieldLen(10, int32(i))
		}
	}
	// Tag 11: weak_dependency
	for i, imp := range f.Imports {
		if imp.Kind == parser.ImportWeak {
			total += int32FieldLen(11, int32(i))
		}
	}
	// Tag 12: syntax
	if f.Syntax == parser.SyntaxProto3 {
		total += stringFieldLen(12, "proto3")
	} else if f.Syntax == parser.SyntaxEditions {
		total += stringFieldLen(12, "editions")
	}
	// Tag 14: edition
	if f.Syntax == parser.SyntaxEditions && f.Edition != 0 {
		total += tagLen(14) + varintLen(uint64(f.Edition))
	}

	return uint32(total)
}

func (s *WireSerializer) computeMessageSize(fileIdx int, f *parser.ASTFile, msgIdx uint32) uint32 {
	m := &f.Messages[msgIdx]
	var total int

	// Tag 1: name
	total += bytesFieldLen(1, m.Name.Bytes(f.Src))

	// Tag 2: field
	for i := m.Fields.Start; i < m.Fields.End(); i++ {
		fi := f.ChildIndices[i]
		fldSize := s.computeFieldSize(f, &f.Fields[fi])
		total += tagLen(2) + varintLen(uint64(fldSize)) + int(fldSize)
	}

	// Tag 3: nested_type
	for i := m.NestedMessages.Start; i < m.NestedMessages.End(); i++ {
		mi := f.ChildIndices[i]
		subSize := s.computeMessageSize(fileIdx, f, mi)
		total += tagLen(3) + varintLen(uint64(subSize)) + int(subSize)
	}

	// Tag 4: enum_type
	for i := m.NestedEnums.Start; i < m.NestedEnums.End(); i++ {
		ei := f.ChildIndices[i]
		subSize := s.computeEnumSize(fileIdx, f, ei)
		total += tagLen(4) + varintLen(uint64(subSize)) + int(subSize)
	}

	// Tag 5: extension_range
	for i := m.ExtensionRanges.Start; i < m.ExtensionRanges.End(); i++ {
		eri := f.ChildIndices[i]
		erSize := s.computeExtensionRangeSize(f, &f.ExtensionRanges[eri])
		total += tagLen(5) + varintLen(uint64(erSize)) + int(erSize)
	}

	// Tag 6: extension
	for i := m.Extensions.Start; i < m.Extensions.End(); i++ {
		xi := f.ChildIndices[i]
		fldSize := s.computeFieldSize(f, &f.Extensions[xi])
		total += tagLen(6) + varintLen(uint64(fldSize)) + int(fldSize)
	}

	// Tag 7: options (MessageOptions)
	mOptSize := s.computeMessageOptionsSize(f, m)
	if mOptSize > 0 {
		total += tagLen(7) + varintLen(uint64(mOptSize)) + int(mOptSize)
	}

	// Tag 8: oneof_decl
	for i := m.Oneofs.Start; i < m.Oneofs.End(); i++ {
		oi := f.ChildIndices[i]
		ooSize := s.computeOneofSize(f, &f.Oneofs[oi])
		total += tagLen(8) + varintLen(uint64(ooSize)) + int(ooSize)
	}

	// Tag 9: reserved_range
	for i := m.ReservedRanges.Start; i < m.ReservedRanges.End(); i++ {
		rri := f.ChildIndices[i]
		rrSize := s.computeReservedRangeSize(&f.ReservedRanges[rri])
		total += tagLen(9) + varintLen(uint64(rrSize)) + int(rrSize)
	}

	// Tag 10: reserved_name
	for i := m.ReservedNames.Start; i < m.ReservedNames.End(); i++ {
		rni := f.ChildIndices[i]
		total += bytesFieldLen(10, f.ReservedNames[rni].Bytes(f.Src))
	}

	for int(msgIdx) >= len(s.msgSizes[fileIdx]) {
		s.msgSizes[fileIdx] = append(s.msgSizes[fileIdx], 0)
	}
	s.msgSizes[fileIdx][msgIdx] = uint32(total)
	return uint32(total)
}

func (s *WireSerializer) computeFieldSize(f *parser.ASTFile, fld *parser.FieldNode) uint32 {
	var total int

	// Tag 1: name
	total += bytesFieldLen(1, fld.Name.Bytes(f.Src))

	// Tag 2: extendee
	if !fld.Extendee.IsEmpty() {
		total += bytesFieldLen(2, fld.Extendee.Bytes(f.Src))
	}

	// Tag 3: number
	total += int32FieldLen(3, fld.Number)

	// Tag 4: label
	if fld.Label != parser.LabelUnspecified {
		total += int32FieldLen(4, int32(fld.Label))
	}

	// Tag 5: type
	if fld.TypeKind != parser.TypeUnresolved {
		total += int32FieldLen(5, int32(fld.TypeKind))
	}

	// Tag 6: type_name
	if !fld.TypeName.IsEmpty() && (fld.TypeKind == parser.TypeMessage || fld.TypeKind == parser.TypeEnum || fld.TypeKind == parser.TypeGroup || fld.TypeKind == parser.TypeUnresolved) {
		total += bytesFieldLen(6, fld.TypeName.Bytes(f.Src))
	}

	// Tag 7: default_value
	if !fld.DefaultVal.IsEmpty() {
		total += bytesFieldLen(7, fld.DefaultVal.Bytes(f.Src))
	}

	// Tag 8: options (FieldOptions)
	fOptSize := s.computeFieldOptionsSize(f, fld)
	if fOptSize > 0 {
		total += tagLen(8) + varintLen(uint64(fOptSize)) + int(fOptSize)
	}

	// Tag 9: oneof_index
	if fld.OneofIndex >= 0 {
		total += int32FieldLen(9, fld.OneofIndex)
	}

	// Tag 10: json_name
	if !fld.CustomJSONName.IsEmpty() {
		total += bytesFieldLen(10, fld.CustomJSONName.Bytes(f.Src))
	} else {
		defJSON := parser.JSONName(fld.Name.String(f.Src))
		total += stringFieldLen(10, defJSON)
	}

	// Tag 17: proto3_optional
	if fld.Proto3Optional {
		total += boolFieldLen(17, true)
	}

	return uint32(total)
}

func (s *WireSerializer) computeEnumSize(fileIdx int, f *parser.ASTFile, enumIdx uint32) uint32 {
	en := &f.Enums[enumIdx]
	var total int

	// Tag 1: name
	total += bytesFieldLen(1, en.Name.Bytes(f.Src))

	// Tag 2: value
	for i := en.Values.Start; i < en.Values.End(); i++ {
		vi := f.ChildIndices[i]
		vSize := s.computeEnumValueSize(f, &f.EnumValues[vi])
		total += tagLen(2) + varintLen(uint64(vSize)) + int(vSize)
	}

	// Tag 3: options
	enumOptSize := s.computeEnumOptionsSize(f, en)
	if enumOptSize > 0 {
		total += tagLen(3) + varintLen(uint64(enumOptSize)) + int(enumOptSize)
	}

	// Tag 4: reserved_range
	for i := en.ReservedRanges.Start; i < en.ReservedRanges.End(); i++ {
		rri := f.ChildIndices[i]
		rrSize := s.computeReservedRangeSize(&f.ReservedRanges[rri])
		total += tagLen(4) + varintLen(uint64(rrSize)) + int(rrSize)
	}

	// Tag 5: reserved_name
	for i := en.ReservedNames.Start; i < en.ReservedNames.End(); i++ {
		rni := f.ChildIndices[i]
		total += bytesFieldLen(5, f.ReservedNames[rni].Bytes(f.Src))
	}

	for int(enumIdx) >= len(s.enumSizes[fileIdx]) {
		s.enumSizes[fileIdx] = append(s.enumSizes[fileIdx], 0)
	}
	s.enumSizes[fileIdx][enumIdx] = uint32(total)
	return uint32(total)
}

func (s *WireSerializer) computeEnumValueSize(f *parser.ASTFile, ev *parser.EnumValueNode) uint32 {
	var total int
	total += bytesFieldLen(1, ev.Name.Bytes(f.Src))
	total += int32FieldLen(2, ev.Number)

	evOptSize := s.computeEnumValueOptionsSize(f, ev)
	if evOptSize > 0 {
		total += tagLen(3) + varintLen(uint64(evOptSize)) + int(evOptSize)
	}
	return uint32(total)
}

func (s *WireSerializer) computeServiceSize(f *parser.ASTFile, svcIdx uint32) uint32 {
	svc := &f.Services[svcIdx]
	var total int
	total += bytesFieldLen(1, svc.Name.Bytes(f.Src))

	for i := svc.Methods.Start; i < svc.Methods.End(); i++ {
		mi := f.ChildIndices[i]
		mSize := s.computeMethodSize(f, &f.Methods[mi])
		total += tagLen(2) + varintLen(uint64(mSize)) + int(mSize)
	}

	sOptSize := s.computeServiceOptionsSize(f, svc)
	if sOptSize > 0 {
		total += tagLen(3) + varintLen(uint64(sOptSize)) + int(sOptSize)
	}
	return uint32(total)
}

func (s *WireSerializer) computeMethodSize(f *parser.ASTFile, m *parser.MethodNode) uint32 {
	var total int
	total += bytesFieldLen(1, m.Name.Bytes(f.Src))
	total += bytesFieldLen(2, m.InputType.Bytes(f.Src))
	total += bytesFieldLen(3, m.OutputType.Bytes(f.Src))

	mOptSize := s.computeMethodOptionsSize(f, m)
	if mOptSize > 0 || m.HasBody {
		total += tagLen(4) + varintLen(uint64(mOptSize)) + int(mOptSize)
	}
	if m.ClientStreaming {
		total += boolFieldLen(5, true)
	}
	if m.ServerStreaming {
		total += boolFieldLen(6, true)
	}
	return uint32(total)
}

func (s *WireSerializer) computeExtensionRangeSize(f *parser.ASTFile, er *parser.ExtensionRangeNode) uint32 {
	var total int
	total += int32FieldLen(1, er.Start)
	total += int32FieldLen(2, er.End)
	if len(er.CustomOptionsWire) > 0 {
		optLen := len(er.CustomOptionsWire)
		total += tagLen(3) + varintLen(uint64(optLen)) + optLen
	}
	return uint32(total)
}

func (s *WireSerializer) computeReservedRangeSize(rr *parser.ReservedRangeNode) uint32 {
	var total int
	total += int32FieldLen(1, rr.Start)
	total += int32FieldLen(2, rr.End)
	return uint32(total)
}

func (s *WireSerializer) computeOneofSize(f *parser.ASTFile, oo *parser.OneofNode) uint32 {
	var total int
	total += bytesFieldLen(1, oo.Name.Bytes(f.Src))
	if len(oo.CustomOptionsWire) > 0 {
		optLen := len(oo.CustomOptionsWire)
		total += tagLen(2) + varintLen(uint64(optLen)) + optLen
	}
	return uint32(total)
}

// Option sizes
func (s *WireSerializer) computeFileOptionsSize(f *parser.ASTFile) uint32 {
	total := len(f.CustomOptionsWire)
	for i := f.FileOptions.Start; i < f.FileOptions.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "java_package":
			total += bytesFieldLen(1, opt.ValSpan.Bytes(f.Src))
		case "java_outer_classname":
			total += bytesFieldLen(8, opt.ValSpan.Bytes(f.Src))
		case "optimize_for":
			val := resolveOptimizeMode(opt.ValSpan.String(f.Src))
			total += tagLen(9) + varintLen(val)
		case "java_multiple_files":
			total += explicitBoolFieldLen(10)
		case "go_package":
			total += bytesFieldLen(11, opt.ValSpan.Bytes(f.Src))
		case "cc_generic_services":
			total += explicitBoolFieldLen(16)
		case "java_generic_services":
			total += explicitBoolFieldLen(17)
		case "py_generic_services":
			total += explicitBoolFieldLen(18)
		case "java_generate_equals_and_hash":
			total += explicitBoolFieldLen(20)
		case "deprecated":
			total += explicitBoolFieldLen(23)
		case "java_string_check_utf8":
			total += explicitBoolFieldLen(27)
		case "cc_enable_arenas":
			total += explicitBoolFieldLen(31)
		case "objc_class_prefix":
			total += bytesFieldLen(36, opt.ValSpan.Bytes(f.Src))
		case "csharp_namespace":
			total += bytesFieldLen(37, opt.ValSpan.Bytes(f.Src))
		case "swift_prefix":
			total += bytesFieldLen(39, opt.ValSpan.Bytes(f.Src))
		case "php_class_prefix":
			total += bytesFieldLen(40, opt.ValSpan.Bytes(f.Src))
		case "php_namespace":
			total += bytesFieldLen(41, opt.ValSpan.Bytes(f.Src))
		case "php_metadata_namespace":
			total += bytesFieldLen(44, opt.ValSpan.Bytes(f.Src))
		case "ruby_package":
			total += bytesFieldLen(45, opt.ValSpan.Bytes(f.Src))
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeMessageOptionsSize(f *parser.ASTFile, m *parser.MessageNode) uint32 {
	total := len(m.CustomOptionsWire)
	if m.IsMapEntry {
		total += boolFieldLen(7, true)
	}
	for i := m.Options.Start; i < m.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "message_set_wire_format":
			total += explicitBoolFieldLen(1)
		case "no_standard_descriptor_accessor":
			total += explicitBoolFieldLen(2)
		case "deprecated":
			total += explicitBoolFieldLen(3)
		case "map_entry":
			if !m.IsMapEntry {
				total += explicitBoolFieldLen(7)
			}
		case "deprecated_legacy_json_field_conflicts":
			total += explicitBoolFieldLen(11)
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeFieldOptionsSize(f *parser.ASTFile, fld *parser.FieldNode) uint32 {
	total := len(fld.CustomOptionsWire)
	for i := fld.Options.Start; i < fld.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "ctype":
			val := resolveCType(opt.ValSpan.String(f.Src))
			total += tagLen(1) + varintLen(val)
		case "packed":
			total += explicitBoolFieldLen(2)
		case "deprecated":
			total += explicitBoolFieldLen(3)
		case "lazy":
			total += explicitBoolFieldLen(5)
		case "jstype":
			val := resolveJSType(opt.ValSpan.String(f.Src))
			total += tagLen(6) + varintLen(val)
		case "weak":
			total += explicitBoolFieldLen(10)
		case "unverified_lazy":
			total += explicitBoolFieldLen(15)
		case "debug_redact":
			total += explicitBoolFieldLen(16)
		case "retention":
			val := resolveOptionRetention(opt.ValSpan.String(f.Src))
			total += tagLen(17) + varintLen(val)
		case "targets":
			val := resolveOptionTargetType(opt.ValSpan.String(f.Src))
			total += tagLen(19) + varintLen(val)
		case "edition_defaults":
			edSize := s.computeEditionDefaultSize(f, opt)
			total += tagLen(20) + varintLen(uint64(edSize)) + int(edSize)
		case "feature_support":
			fsSize := s.computeFeatureSupportSize(f, opt)
			total += tagLen(22) + varintLen(uint64(fsSize)) + int(fsSize)
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeEditionDefaultSize(f *parser.ASTFile, opt *parser.OptionNode) uint32 {
	var total int
	for i := opt.SubOptions.Start; i < opt.SubOptions.End(); i++ {
		sub := &f.Options[f.ChildIndices[i]]
		if sub.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[sub.NameParts.Start]
		subName := part.Name.String(f.Src)
		switch subName {
		case "value":
			total += bytesFieldLen(2, sub.ValSpan.Bytes(f.Src))
		case "edition":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			total += tagLen(3) + varintLen(ed)
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeFeatureSupportSize(f *parser.ASTFile, opt *parser.OptionNode) uint32 {
	var total int
	for i := opt.SubOptions.Start; i < opt.SubOptions.End(); i++ {
		sub := &f.Options[f.ChildIndices[i]]
		if sub.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[sub.NameParts.Start]
		subName := part.Name.String(f.Src)
		switch subName {
		case "edition_introduced":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			total += tagLen(1) + varintLen(ed)
		case "edition_deprecated":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			total += tagLen(2) + varintLen(ed)
		case "deprecation_warning":
			total += bytesFieldLen(3, sub.ValSpan.Bytes(f.Src))
		case "edition_removed":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			total += tagLen(4) + varintLen(ed)
		case "removal_error":
			total += bytesFieldLen(5, sub.ValSpan.Bytes(f.Src))
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeEnumOptionsSize(f *parser.ASTFile, en *parser.EnumNode) uint32 {
	total := len(en.CustomOptionsWire)
	for i := en.Options.Start; i < en.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "allow_alias":
			total += explicitBoolFieldLen(2)
		case "deprecated":
			total += explicitBoolFieldLen(3)
		case "deprecated_legacy_json_field_conflicts":
			total += explicitBoolFieldLen(6)
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeEnumValueOptionsSize(f *parser.ASTFile, ev *parser.EnumValueNode) uint32 {
	total := len(ev.CustomOptionsWire)
	for i := ev.Options.Start; i < ev.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "deprecated":
			total += explicitBoolFieldLen(1)
		case "debug_redact":
			total += explicitBoolFieldLen(3)
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeServiceOptionsSize(f *parser.ASTFile, svc *parser.ServiceNode) uint32 {
	total := len(svc.CustomOptionsWire)
	for i := svc.Options.Start; i < svc.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "deprecated":
			total += explicitBoolFieldLen(33)
		}
	}
	return uint32(total)
}

func (s *WireSerializer) computeMethodOptionsSize(f *parser.ASTFile, m *parser.MethodNode) uint32 {
	total := len(m.CustomOptionsWire)
	for i := m.Options.Start; i < m.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "deprecated":
			total += explicitBoolFieldLen(33)
		case "idempotency_level":
			val := resolveIdempotencyLevel(opt.ValSpan.String(f.Src))
			total += tagLen(34) + varintLen(val)
		}
	}
	return uint32(total)
}

// Pass 2: Direct encoding into pre-allocated buffer
func (s *WireSerializer) writeFileDescriptorProto(buf []byte, pos int, fileIdx int, f *parser.ASTFile, includeSourceInfo bool) int {
	// Tag 1: name (string)
	if f.Path != "" {
		pos = writeStringField(buf, pos, 1, f.Path)
	}
	// Tag 2: package (string)
	if !f.Package.IsEmpty() {
		pos = writeBytesField(buf, pos, 2, f.Package.Bytes(f.Src))
	}
	// Tag 3: dependency (repeated string)
	for _, imp := range f.Imports {
		pos = writeBytesField(buf, pos, 3, imp.Path.Bytes(f.Src))
	}
	// Tag 4: message_type (repeated DescriptorProto)
	for i := f.TopLevelMessages.Start; i < f.TopLevelMessages.End(); i++ {
		mi := f.ChildIndices[i]
		mSize := s.msgSizes[fileIdx][mi]
		pos = writeTag(buf, pos, 4, 2)
		pos = writeVarint(buf, pos, uint64(mSize))
		pos = s.writeDescriptorProto(buf, pos, fileIdx, f, mi)
	}
	// Tag 5: enum_type (repeated EnumDescriptorProto)
	for i := f.TopLevelEnums.Start; i < f.TopLevelEnums.End(); i++ {
		ei := f.ChildIndices[i]
		eSize := s.enumSizes[fileIdx][ei]
		pos = writeTag(buf, pos, 5, 2)
		pos = writeVarint(buf, pos, uint64(eSize))
		pos = s.writeEnumDescriptorProto(buf, pos, f, ei)
	}
	// Tag 6: service (repeated ServiceDescriptorProto)
	for si := range f.Services {
		sSize := s.computeServiceSize(f, uint32(si))
		pos = writeTag(buf, pos, 6, 2)
		pos = writeVarint(buf, pos, uint64(sSize))
		pos = s.writeServiceDescriptorProto(buf, pos, f, uint32(si))
	}
	// Tag 7: extension (repeated FieldDescriptorProto)
	for xi := range f.Extensions {
		fldSize := s.computeFieldSize(f, &f.Extensions[xi])
		pos = writeTag(buf, pos, 7, 2)
		pos = writeVarint(buf, pos, uint64(fldSize))
		pos = s.writeFieldDescriptorProto(buf, pos, f, &f.Extensions[xi])
	}
	// Tag 8: options (FileOptions)
	optSize := s.computeFileOptionsSize(f)
	if optSize > 0 {
		pos = writeTag(buf, pos, 8, 2)
		pos = writeVarint(buf, pos, uint64(optSize))
		pos = s.writeFileOptions(buf, pos, f)
	}
	// Tag 9: source_code_info
	if includeSourceInfo && len(f.SourceInfoWire) > 0 {
		pos = writeBytesField(buf, pos, 9, f.SourceInfoWire)
	}
	// Tag 10: public_dependency (repeated int32)
	for i, imp := range f.Imports {
		if imp.Kind == parser.ImportPublic {
			pos = writeInt32Field(buf, pos, 10, int32(i))
		}
	}
	// Tag 11: weak_dependency (repeated int32)
	for i, imp := range f.Imports {
		if imp.Kind == parser.ImportWeak {
			pos = writeInt32Field(buf, pos, 11, int32(i))
		}
	}
	// Tag 12: syntax (string)
	if f.Syntax == parser.SyntaxProto3 {
		pos = writeStringField(buf, pos, 12, "proto3")
	} else if f.Syntax == parser.SyntaxEditions {
		pos = writeStringField(buf, pos, 12, "editions")
	}
	// Tag 14: edition (Edition)
	if f.Syntax == parser.SyntaxEditions && f.Edition != 0 {
		pos = writeTag(buf, pos, 14, 0)
		pos = writeVarint(buf, pos, uint64(f.Edition))
	}

	return pos
}

func (s *WireSerializer) writeDescriptorProto(buf []byte, pos int, fileIdx int, f *parser.ASTFile, msgIdx uint32) int {
	m := &f.Messages[msgIdx]

	// Tag 1: name
	pos = writeBytesField(buf, pos, 1, m.Name.Bytes(f.Src))

	// Tag 2: field (repeated FieldDescriptorProto)
	for i := m.Fields.Start; i < m.Fields.End(); i++ {
		fi := f.ChildIndices[i]
		fldSize := s.computeFieldSize(f, &f.Fields[fi])
		pos = writeTag(buf, pos, 2, 2)
		pos = writeVarint(buf, pos, uint64(fldSize))
		pos = s.writeFieldDescriptorProto(buf, pos, f, &f.Fields[fi])
	}

	// Tag 3: nested_type (repeated DescriptorProto)
	for i := m.NestedMessages.Start; i < m.NestedMessages.End(); i++ {
		mi := f.ChildIndices[i]
		subSize := s.msgSizes[fileIdx][mi]
		pos = writeTag(buf, pos, 3, 2)
		pos = writeVarint(buf, pos, uint64(subSize))
		pos = s.writeDescriptorProto(buf, pos, fileIdx, f, mi)
	}

	// Tag 4: enum_type (repeated EnumDescriptorProto)
	for i := m.NestedEnums.Start; i < m.NestedEnums.End(); i++ {
		ei := f.ChildIndices[i]
		subSize := s.enumSizes[fileIdx][ei]
		pos = writeTag(buf, pos, 4, 2)
		pos = writeVarint(buf, pos, uint64(subSize))
		pos = s.writeEnumDescriptorProto(buf, pos, f, ei)
	}

	// Tag 5: extension_range (repeated ExtensionRange)
	for i := m.ExtensionRanges.Start; i < m.ExtensionRanges.End(); i++ {
		eri := f.ChildIndices[i]
		erSize := s.computeExtensionRangeSize(f, &f.ExtensionRanges[eri])
		pos = writeTag(buf, pos, 5, 2)
		pos = writeVarint(buf, pos, uint64(erSize))
		pos = s.writeExtensionRange(buf, pos, &f.ExtensionRanges[eri])
	}

	// Tag 6: extension (repeated FieldDescriptorProto)
	for i := m.Extensions.Start; i < m.Extensions.End(); i++ {
		xi := f.ChildIndices[i]
		fldSize := s.computeFieldSize(f, &f.Extensions[xi])
		pos = writeTag(buf, pos, 6, 2)
		pos = writeVarint(buf, pos, uint64(fldSize))
		pos = s.writeFieldDescriptorProto(buf, pos, f, &f.Extensions[xi])
	}

	// Tag 7: options (MessageOptions)
	mOptSize := s.computeMessageOptionsSize(f, m)
	if mOptSize > 0 {
		pos = writeTag(buf, pos, 7, 2)
		pos = writeVarint(buf, pos, uint64(mOptSize))
		pos = s.writeMessageOptions(buf, pos, f, m)
	}

	// Tag 8: oneof_decl (repeated OneofDescriptorProto)
	for i := m.Oneofs.Start; i < m.Oneofs.End(); i++ {
		oi := f.ChildIndices[i]
		ooSize := s.computeOneofSize(f, &f.Oneofs[oi])
		pos = writeTag(buf, pos, 8, 2)
		pos = writeVarint(buf, pos, uint64(ooSize))
		pos = s.writeOneofDescriptorProto(buf, pos, f, &f.Oneofs[oi])
	}

	// Tag 9: reserved_range (repeated ReservedRange)
	for i := m.ReservedRanges.Start; i < m.ReservedRanges.End(); i++ {
		rri := f.ChildIndices[i]
		rrSize := s.computeReservedRangeSize(&f.ReservedRanges[rri])
		pos = writeTag(buf, pos, 9, 2)
		pos = writeVarint(buf, pos, uint64(rrSize))
		pos = s.writeReservedRange(buf, pos, &f.ReservedRanges[rri])
	}

	// Tag 10: reserved_name (repeated string)
	for i := m.ReservedNames.Start; i < m.ReservedNames.End(); i++ {
		rni := f.ChildIndices[i]
		pos = writeBytesField(buf, pos, 10, f.ReservedNames[rni].Bytes(f.Src))
	}

	return pos
}

func (s *WireSerializer) writeFieldDescriptorProto(buf []byte, pos int, f *parser.ASTFile, fld *parser.FieldNode) int {
	// Tag 1: name
	pos = writeBytesField(buf, pos, 1, fld.Name.Bytes(f.Src))

	// Tag 2: extendee
	if !fld.Extendee.IsEmpty() {
		pos = writeBytesField(buf, pos, 2, fld.Extendee.Bytes(f.Src))
	}

	// Tag 3: number
	pos = writeInt32Field(buf, pos, 3, fld.Number)

	// Tag 4: label
	if fld.Label != parser.LabelUnspecified {
		pos = writeInt32Field(buf, pos, 4, int32(fld.Label))
	}

	// Tag 5: type
	if fld.TypeKind != parser.TypeUnresolved {
		pos = writeInt32Field(buf, pos, 5, int32(fld.TypeKind))
	}

	// Tag 6: type_name
	if !fld.TypeName.IsEmpty() && (fld.TypeKind == parser.TypeMessage || fld.TypeKind == parser.TypeEnum || fld.TypeKind == parser.TypeGroup || fld.TypeKind == parser.TypeUnresolved) {
		pos = writeBytesField(buf, pos, 6, fld.TypeName.Bytes(f.Src))
	}

	// Tag 7: default_value
	if !fld.DefaultVal.IsEmpty() {
		pos = writeBytesField(buf, pos, 7, fld.DefaultVal.Bytes(f.Src))
	}

	// Tag 8: options
	fOptSize := s.computeFieldOptionsSize(f, fld)
	if fOptSize > 0 {
		pos = writeTag(buf, pos, 8, 2)
		pos = writeVarint(buf, pos, uint64(fOptSize))
		pos = s.writeFieldOptions(buf, pos, f, fld)
	}

	// Tag 9: oneof_index
	if fld.OneofIndex >= 0 {
		pos = writeInt32Field(buf, pos, 9, fld.OneofIndex)
	}

	// Tag 10: json_name
	if !fld.CustomJSONName.IsEmpty() {
		pos = writeBytesField(buf, pos, 10, fld.CustomJSONName.Bytes(f.Src))
	} else {
		defJSON := parser.JSONName(fld.Name.String(f.Src))
		pos = writeStringField(buf, pos, 10, defJSON)
	}

	// Tag 17: proto3_optional
	if fld.Proto3Optional {
		pos = writeBoolField(buf, pos, 17, true)
	}

	return pos
}

func (s *WireSerializer) writeEnumDescriptorProto(buf []byte, pos int, f *parser.ASTFile, enumIdx uint32) int {
	en := &f.Enums[enumIdx]

	// Tag 1: name
	pos = writeBytesField(buf, pos, 1, en.Name.Bytes(f.Src))

	// Tag 2: value (repeated EnumValueDescriptorProto)
	for i := en.Values.Start; i < en.Values.End(); i++ {
		vi := f.ChildIndices[i]
		vSize := s.computeEnumValueSize(f, &f.EnumValues[vi])
		pos = writeTag(buf, pos, 2, 2)
		pos = writeVarint(buf, pos, uint64(vSize))
		pos = s.writeEnumValueDescriptorProto(buf, pos, f, &f.EnumValues[vi])
	}

	// Tag 3: options
	enumOptSize := s.computeEnumOptionsSize(f, en)
	if enumOptSize > 0 {
		pos = writeTag(buf, pos, 3, 2)
		pos = writeVarint(buf, pos, uint64(enumOptSize))
		pos = s.writeEnumOptions(buf, pos, f, en)
	}

	// Tag 4: reserved_range
	for i := en.ReservedRanges.Start; i < en.ReservedRanges.End(); i++ {
		rri := f.ChildIndices[i]
		rrSize := s.computeReservedRangeSize(&f.ReservedRanges[rri])
		pos = writeTag(buf, pos, 4, 2)
		pos = writeVarint(buf, pos, uint64(rrSize))
		pos = s.writeReservedRange(buf, pos, &f.ReservedRanges[rri])
	}

	// Tag 5: reserved_name
	for i := en.ReservedNames.Start; i < en.ReservedNames.End(); i++ {
		rni := f.ChildIndices[i]
		pos = writeBytesField(buf, pos, 5, f.ReservedNames[rni].Bytes(f.Src))
	}

	return pos
}

func (s *WireSerializer) writeEnumValueDescriptorProto(buf []byte, pos int, f *parser.ASTFile, ev *parser.EnumValueNode) int {
	pos = writeBytesField(buf, pos, 1, ev.Name.Bytes(f.Src))
	pos = writeInt32Field(buf, pos, 2, ev.Number)

	evOptSize := s.computeEnumValueOptionsSize(f, ev)
	if evOptSize > 0 {
		pos = writeTag(buf, pos, 3, 2)
		pos = writeVarint(buf, pos, uint64(evOptSize))
		pos = s.writeEnumValueOptions(buf, pos, f, ev)
	}
	return pos
}

func (s *WireSerializer) writeServiceDescriptorProto(buf []byte, pos int, f *parser.ASTFile, svcIdx uint32) int {
	svc := &f.Services[svcIdx]
	pos = writeBytesField(buf, pos, 1, svc.Name.Bytes(f.Src))

	for i := svc.Methods.Start; i < svc.Methods.End(); i++ {
		mi := f.ChildIndices[i]
		mSize := s.computeMethodSize(f, &f.Methods[mi])
		pos = writeTag(buf, pos, 2, 2)
		pos = writeVarint(buf, pos, uint64(mSize))
		pos = s.writeMethodDescriptorProto(buf, pos, f, &f.Methods[mi])
	}

	sOptSize := s.computeServiceOptionsSize(f, svc)
	if sOptSize > 0 {
		pos = writeTag(buf, pos, 3, 2)
		pos = writeVarint(buf, pos, uint64(sOptSize))
		pos = s.writeServiceOptions(buf, pos, f, svc)
	}
	return pos
}

func (s *WireSerializer) writeMethodDescriptorProto(buf []byte, pos int, f *parser.ASTFile, m *parser.MethodNode) int {
	pos = writeBytesField(buf, pos, 1, m.Name.Bytes(f.Src))
	pos = writeBytesField(buf, pos, 2, m.InputType.Bytes(f.Src))
	pos = writeBytesField(buf, pos, 3, m.OutputType.Bytes(f.Src))

	mOptSize := s.computeMethodOptionsSize(f, m)
	if mOptSize > 0 || m.HasBody {
		pos = writeTag(buf, pos, 4, 2)
		pos = writeVarint(buf, pos, uint64(mOptSize))
		pos = s.writeMethodOptions(buf, pos, f, m)
	}
	if m.ClientStreaming {
		pos = writeBoolField(buf, pos, 5, true)
	}
	if m.ServerStreaming {
		pos = writeBoolField(buf, pos, 6, true)
	}
	return pos
}

func (s *WireSerializer) writeExtensionRange(buf []byte, pos int, er *parser.ExtensionRangeNode) int {
	pos = writeInt32Field(buf, pos, 1, er.Start)
	pos = writeInt32Field(buf, pos, 2, er.End)
	if len(er.CustomOptionsWire) > 0 {
		pos = writeTag(buf, pos, 3, 2)
		pos = writeVarint(buf, pos, uint64(len(er.CustomOptionsWire)))
		copy(buf[pos:], er.CustomOptionsWire)
		pos += len(er.CustomOptionsWire)
	}
	return pos
}

func (s *WireSerializer) writeReservedRange(buf []byte, pos int, rr *parser.ReservedRangeNode) int {
	pos = writeInt32Field(buf, pos, 1, rr.Start)
	pos = writeInt32Field(buf, pos, 2, rr.End)
	return pos
}

func (s *WireSerializer) writeOneofDescriptorProto(buf []byte, pos int, f *parser.ASTFile, oo *parser.OneofNode) int {
	pos = writeBytesField(buf, pos, 1, oo.Name.Bytes(f.Src))
	if len(oo.CustomOptionsWire) > 0 {
		pos = writeTag(buf, pos, 2, 2)
		pos = writeVarint(buf, pos, uint64(len(oo.CustomOptionsWire)))
		copy(buf[pos:], oo.CustomOptionsWire)
		pos += len(oo.CustomOptionsWire)
	}
	return pos
}

// Option writers
func (s *WireSerializer) writeFileOptions(buf []byte, pos int, f *parser.ASTFile) int {
	for i := f.FileOptions.Start; i < f.FileOptions.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "java_package":
			pos = writeBytesField(buf, pos, 1, opt.ValSpan.Bytes(f.Src))
		case "java_outer_classname":
			pos = writeBytesField(buf, pos, 8, opt.ValSpan.Bytes(f.Src))
		case "optimize_for":
			val := resolveOptimizeMode(opt.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 9, 0)
			pos = writeVarint(buf, pos, val)
		case "java_multiple_files":
			pos = writeExplicitBoolField(buf, pos, 10, isBoolTrue(opt, f.Src))
		case "go_package":
			pos = writeBytesField(buf, pos, 11, opt.ValSpan.Bytes(f.Src))
		case "cc_generic_services":
			pos = writeExplicitBoolField(buf, pos, 16, isBoolTrue(opt, f.Src))
		case "java_generic_services":
			pos = writeExplicitBoolField(buf, pos, 17, isBoolTrue(opt, f.Src))
		case "py_generic_services":
			pos = writeExplicitBoolField(buf, pos, 18, isBoolTrue(opt, f.Src))
		case "java_generate_equals_and_hash":
			pos = writeExplicitBoolField(buf, pos, 20, isBoolTrue(opt, f.Src))
		case "deprecated":
			pos = writeExplicitBoolField(buf, pos, 23, isBoolTrue(opt, f.Src))
		case "java_string_check_utf8":
			pos = writeExplicitBoolField(buf, pos, 27, isBoolTrue(opt, f.Src))
		case "cc_enable_arenas":
			pos = writeExplicitBoolField(buf, pos, 31, isBoolTrue(opt, f.Src))
		case "objc_class_prefix":
			pos = writeBytesField(buf, pos, 36, opt.ValSpan.Bytes(f.Src))
		case "csharp_namespace":
			pos = writeBytesField(buf, pos, 37, opt.ValSpan.Bytes(f.Src))
		case "swift_prefix":
			pos = writeBytesField(buf, pos, 39, opt.ValSpan.Bytes(f.Src))
		case "php_class_prefix":
			pos = writeBytesField(buf, pos, 40, opt.ValSpan.Bytes(f.Src))
		case "php_namespace":
			pos = writeBytesField(buf, pos, 41, opt.ValSpan.Bytes(f.Src))
		case "php_metadata_namespace":
			pos = writeBytesField(buf, pos, 44, opt.ValSpan.Bytes(f.Src))
		case "ruby_package":
			pos = writeBytesField(buf, pos, 45, opt.ValSpan.Bytes(f.Src))
		}
	}
	if len(f.CustomOptionsWire) > 0 {
		copy(buf[pos:], f.CustomOptionsWire)
		pos += len(f.CustomOptionsWire)
	}
	return pos
}

func (s *WireSerializer) writeMessageOptions(buf []byte, pos int, f *parser.ASTFile, m *parser.MessageNode) int {
	for i := m.Options.Start; i < m.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "message_set_wire_format":
			pos = writeExplicitBoolField(buf, pos, 1, isBoolTrue(opt, f.Src))
		case "no_standard_descriptor_accessor":
			pos = writeExplicitBoolField(buf, pos, 2, isBoolTrue(opt, f.Src))
		case "deprecated":
			pos = writeExplicitBoolField(buf, pos, 3, isBoolTrue(opt, f.Src))
		}
	}
	if m.IsMapEntry {
		pos = writeBoolField(buf, pos, 7, true)
	}
	for i := m.Options.Start; i < m.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "map_entry":
			if !m.IsMapEntry {
				pos = writeExplicitBoolField(buf, pos, 7, isBoolTrue(opt, f.Src))
			}
		case "deprecated_legacy_json_field_conflicts":
			pos = writeExplicitBoolField(buf, pos, 11, isBoolTrue(opt, f.Src))
		}
	}
	if len(m.CustomOptionsWire) > 0 {
		copy(buf[pos:], m.CustomOptionsWire)
		pos += len(m.CustomOptionsWire)
	}
	return pos
}

func (s *WireSerializer) writeFieldOptions(buf []byte, pos int, f *parser.ASTFile, fld *parser.FieldNode) int {
	for i := fld.Options.Start; i < fld.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "ctype":
			val := resolveCType(opt.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 1, 0)
			pos = writeVarint(buf, pos, val)
		case "packed":
			pos = writeExplicitBoolField(buf, pos, 2, isBoolTrue(opt, f.Src))
		case "deprecated":
			pos = writeExplicitBoolField(buf, pos, 3, isBoolTrue(opt, f.Src))
		case "lazy":
			pos = writeExplicitBoolField(buf, pos, 5, isBoolTrue(opt, f.Src))
		case "jstype":
			val := resolveJSType(opt.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 6, 0)
			pos = writeVarint(buf, pos, val)
		case "weak":
			pos = writeExplicitBoolField(buf, pos, 10, isBoolTrue(opt, f.Src))
		case "unverified_lazy":
			pos = writeExplicitBoolField(buf, pos, 15, isBoolTrue(opt, f.Src))
		case "debug_redact":
			pos = writeExplicitBoolField(buf, pos, 16, isBoolTrue(opt, f.Src))
		case "retention":
			val := resolveOptionRetention(opt.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 17, 0)
			pos = writeVarint(buf, pos, val)
		case "targets":
			val := resolveOptionTargetType(opt.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 19, 0)
			pos = writeVarint(buf, pos, val)
		case "edition_defaults":
			edSize := s.computeEditionDefaultSize(f, opt)
			pos = writeTag(buf, pos, 20, 2)
			pos = writeVarint(buf, pos, uint64(edSize))
			pos = s.writeEditionDefault(buf, pos, f, opt)
		case "feature_support":
			fsSize := s.computeFeatureSupportSize(f, opt)
			pos = writeTag(buf, pos, 22, 2)
			pos = writeVarint(buf, pos, uint64(fsSize))
			pos = s.writeFeatureSupport(buf, pos, f, opt)
		}
	}
	if len(fld.CustomOptionsWire) > 0 {
		copy(buf[pos:], fld.CustomOptionsWire)
		pos += len(fld.CustomOptionsWire)
	}
	return pos
}

func (s *WireSerializer) writeEditionDefault(buf []byte, pos int, f *parser.ASTFile, opt *parser.OptionNode) int {
	for i := opt.SubOptions.Start; i < opt.SubOptions.End(); i++ {
		sub := &f.Options[f.ChildIndices[i]]
		if sub.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[sub.NameParts.Start]
		subName := part.Name.String(f.Src)
		switch subName {
		case "value":
			pos = writeBytesField(buf, pos, 2, sub.ValSpan.Bytes(f.Src))
		case "edition":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 3, 0)
			pos = writeVarint(buf, pos, ed)
		}
	}
	return pos
}

func (s *WireSerializer) writeFeatureSupport(buf []byte, pos int, f *parser.ASTFile, opt *parser.OptionNode) int {
	for i := opt.SubOptions.Start; i < opt.SubOptions.End(); i++ {
		sub := &f.Options[f.ChildIndices[i]]
		if sub.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[sub.NameParts.Start]
		subName := part.Name.String(f.Src)
		switch subName {
		case "edition_introduced":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 1, 0)
			pos = writeVarint(buf, pos, ed)
		case "edition_deprecated":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 2, 0)
			pos = writeVarint(buf, pos, ed)
		case "deprecation_warning":
			pos = writeBytesField(buf, pos, 3, sub.ValSpan.Bytes(f.Src))
		case "edition_removed":
			ed := resolveEdition(sub.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 4, 0)
			pos = writeVarint(buf, pos, ed)
		case "removal_error":
			pos = writeBytesField(buf, pos, 5, sub.ValSpan.Bytes(f.Src))
		}
	}
	return pos
}

func (s *WireSerializer) writeEnumOptions(buf []byte, pos int, f *parser.ASTFile, en *parser.EnumNode) int {
	for i := en.Options.Start; i < en.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "allow_alias":
			pos = writeExplicitBoolField(buf, pos, 2, isBoolTrue(opt, f.Src))
		case "deprecated":
			pos = writeExplicitBoolField(buf, pos, 3, isBoolTrue(opt, f.Src))
		case "deprecated_legacy_json_field_conflicts":
			pos = writeExplicitBoolField(buf, pos, 6, isBoolTrue(opt, f.Src))
		}
	}
	if len(en.CustomOptionsWire) > 0 {
		copy(buf[pos:], en.CustomOptionsWire)
		pos += len(en.CustomOptionsWire)
	}
	return pos
}

func (s *WireSerializer) writeEnumValueOptions(buf []byte, pos int, f *parser.ASTFile, ev *parser.EnumValueNode) int {
	for i := ev.Options.Start; i < ev.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "deprecated":
			pos = writeExplicitBoolField(buf, pos, 1, isBoolTrue(opt, f.Src))
		case "debug_redact":
			pos = writeExplicitBoolField(buf, pos, 3, isBoolTrue(opt, f.Src))
		}
	}
	if len(ev.CustomOptionsWire) > 0 {
		copy(buf[pos:], ev.CustomOptionsWire)
		pos += len(ev.CustomOptionsWire)
	}
	return pos
}

func (s *WireSerializer) writeServiceOptions(buf []byte, pos int, f *parser.ASTFile, svc *parser.ServiceNode) int {
	for i := svc.Options.Start; i < svc.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "deprecated":
			pos = writeExplicitBoolField(buf, pos, 33, isBoolTrue(opt, f.Src))
		}
	}
	if len(svc.CustomOptionsWire) > 0 {
		copy(buf[pos:], svc.CustomOptionsWire)
		pos += len(svc.CustomOptionsWire)
	}
	return pos
}

func (s *WireSerializer) writeMethodOptions(buf []byte, pos int, f *parser.ASTFile, m *parser.MethodNode) int {
	for i := m.Options.Start; i < m.Options.End(); i++ {
		opt := &f.Options[f.ChildIndices[i]]
		if opt.NameParts.IsEmpty() {
			continue
		}
		part := &f.OptionNameParts[opt.NameParts.Start]
		if part.IsExtension {
			continue
		}
		name := part.Name.String(f.Src)
		switch name {
		case "deprecated":
			pos = writeExplicitBoolField(buf, pos, 33, isBoolTrue(opt, f.Src))
		case "idempotency_level":
			val := resolveIdempotencyLevel(opt.ValSpan.String(f.Src))
			pos = writeTag(buf, pos, 34, 0)
			pos = writeVarint(buf, pos, val)
		}
	}
	if len(m.CustomOptionsWire) > 0 {
		copy(buf[pos:], m.CustomOptionsWire)
		pos += len(m.CustomOptionsWire)
	}
	return pos
}

// Option enum resolvers
func resolveEdition(name string) uint64 {
	switch name {
	case "EDITION_LEGACY":
		return 900
	case "EDITION_PROTO2":
		return 998
	case "EDITION_PROTO3":
		return 999
	case "EDITION_2023":
		return 1000
	case "EDITION_2024":
		return 1001
	case "EDITION_2026":
		return 1002
	default:
		return 0
	}
}

func resolveOptionTargetType(name string) uint64 {
	switch name {
	case "TARGET_TYPE_FILE":
		return 1
	case "TARGET_TYPE_EXTENSION_RANGE":
		return 2
	case "TARGET_TYPE_MESSAGE":
		return 3
	case "TARGET_TYPE_FIELD":
		return 4
	case "TARGET_TYPE_ONEOF":
		return 5
	case "TARGET_TYPE_ENUM":
		return 6
	case "TARGET_TYPE_ENUM_ENTRY":
		return 7
	case "TARGET_TYPE_SERVICE":
		return 8
	case "TARGET_TYPE_METHOD":
		return 9
	default:
		return 0
	}
}

func resolveOptionRetention(name string) uint64 {
	switch name {
	case "RETENTION_RUNTIME":
		return 1
	case "RETENTION_SOURCE":
		return 2
	default:
		return 0
	}
}

func resolveOptimizeMode(name string) uint64 {
	switch name {
	case "SPEED":
		return 1
	case "CODE_SIZE":
		return 2
	case "LITE_RUNTIME":
		return 3
	default:
		return 1
	}
}

func resolveCType(name string) uint64 {
	switch name {
	case "STRING":
		return 0
	case "CORD":
		return 1
	case "STRING_PIECE":
		return 2
	default:
		return 0
	}
}

func resolveJSType(name string) uint64 {
	switch name {
	case "JS_NORMAL":
		return 0
	case "JS_STRING":
		return 1
	case "JS_NUMBER":
		return 2
	default:
		return 0
	}
}

func resolveIdempotencyLevel(name string) uint64 {
	switch name {
	case "IDEMPOTENCY_UNKNOWN":
		return 0
	case "NO_SIDE_EFFECTS":
		return 1
	case "IDEMPOTENT":
		return 2
	default:
		return 0
	}
}

func isBoolTrue(opt *parser.OptionNode, src []byte) bool {
	if opt.ValKind == parser.OptionValIdent {
		return opt.ValSpan.String(src) == "true"
	}
	if opt.ValKind == parser.OptionValInt {
		return opt.IntVal != 0
	}
	return false
}
