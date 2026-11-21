package options

import (
	"encoding/binary"
	"math"
)

func appendVarint(buf []byte, v uint64) []byte {
	for v >= 0x80 {
		buf = append(buf, byte(v)|0x80)
		v >>= 7
	}
	return append(buf, byte(v))
}

func appendTag(buf []byte, fieldNum uint32, wireType uint8) []byte {
	return appendVarint(buf, uint64(fieldNum<<3)|uint64(wireType))
}

func appendFixed32(buf []byte, v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return append(buf, b[:]...)
}

func appendFixed64(buf []byte, v uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return append(buf, b[:]...)
}

func appendFloat32(buf []byte, f float32) []byte {
	return appendFixed32(buf, math.Float32bits(f))
}

func appendFloat64(buf []byte, f float64) []byte {
	return appendFixed64(buf, math.Float64bits(f))
}

func appendBytesField(buf []byte, tag uint32, b []byte) []byte {
	buf = appendTag(buf, tag, 2)
	buf = appendVarint(buf, uint64(len(b)))
	return append(buf, b...)
}
