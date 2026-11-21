package sourceinfo

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

func writeStringField(buf []byte, pos int, fieldNum uint32, s string) int {
	if len(s) == 0 {
		return pos
	}
	pos = writeTag(buf, pos, fieldNum, 2)
	pos = writeVarint(buf, pos, uint64(len(s)))
	copy(buf[pos:], s)
	return pos + len(s)
}

func computeLocationSize(loc *Location) (pathPayload uint32, spanPayload uint32, locSize uint32) {
	for _, p := range loc.Path {
		pathPayload += uint32(varintLen(uint64(p)))
	}
	for i := 0; i < int(loc.SpanLen); i++ {
		spanPayload += uint32(varintLen(uint64(loc.Span[i])))
	}

	if len(loc.Path) > 0 {
		locSize += uint32(tagLen(1)+varintLen(uint64(pathPayload))) + pathPayload
	}
	locSize += uint32(tagLen(2)+varintLen(uint64(spanPayload))) + spanPayload
	if len(loc.LeadingComments) > 0 {
		locSize += uint32(stringFieldLen(3, loc.LeadingComments))
	}
	if len(loc.TrailingComments) > 0 {
		locSize += uint32(stringFieldLen(4, loc.TrailingComments))
	}
	for _, d := range loc.LeadingDetachedComments {
		locSize += uint32(stringFieldLen(6, d))
	}
	return pathPayload, spanPayload, locSize
}

// Serialize encodes a FileSourceInfo into wire bytes matching descriptorpb.SourceCodeInfo.
func Serialize(sci *FileSourceInfo) []byte {
	if sci == nil || len(sci.Locations) == 0 {
		return nil
	}

	// Pass 1: compute size
	var totalSize uint32
	for i := range sci.Locations {
		_, _, locSize := computeLocationSize(&sci.Locations[i])
		totalSize += uint32(tagLen(1)+varintLen(uint64(locSize))) + locSize
	}

	// Pass 2: allocate single buffer and write
	buf := make([]byte, totalSize)
	pos := 0

	for i := range sci.Locations {
		loc := &sci.Locations[i]
		pathPayload, spanPayload, locSize := computeLocationSize(loc)
		pos = writeTag(buf, pos, 1, 2)
		pos = writeVarint(buf, pos, uint64(locSize))

		// Tag 1: path
		if len(loc.Path) > 0 {
			pos = writeTag(buf, pos, 1, 2)
			pos = writeVarint(buf, pos, uint64(pathPayload))
			for _, p := range loc.Path {
				pos = writeVarint(buf, pos, uint64(p))
			}
		}

		// Tag 2: span
		pos = writeTag(buf, pos, 2, 2)
		pos = writeVarint(buf, pos, uint64(spanPayload))
		for j := 0; j < int(loc.SpanLen); j++ {
			pos = writeVarint(buf, pos, uint64(loc.Span[j]))
		}

		// Tag 3: leading_comments
		if len(loc.LeadingComments) > 0 {
			pos = writeStringField(buf, pos, 3, loc.LeadingComments)
		}

		// Tag 4: trailing_comments
		if len(loc.TrailingComments) > 0 {
			pos = writeStringField(buf, pos, 4, loc.TrailingComments)
		}

		// Tag 6: leading_detached_comments
		for _, d := range loc.LeadingDetachedComments {
			pos = writeStringField(buf, pos, 6, d)
		}
	}

	return buf[:pos]
}
