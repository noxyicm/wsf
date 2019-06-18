package transporter

import (
	"bytes"
	"encoding/binary"
)

// min provides int comparison
func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// minu32 provides uint32 comparison
func minu32(a, b uint32) uint32 {
	if a < b {
		return a
	}

	return b
}

// minu64 provides uint64 comparison
func minu64(a, b uint64) uint64 {
	if a < b {
		return a
	}

	return b
}

// Pack converts the string into byte slice
func Pack(m string, s uint64) []byte {
	b := bytes.Buffer{}
	b.WriteString(m)

	b.Write([]byte{
		byte(s),
		byte(s >> 8),
		byte(s >> 16),
		byte(s >> 24),
		byte(s >> 32),
		byte(s >> 40),
		byte(s >> 48),
		byte(s >> 56),
	})

	return b.Bytes()
}

// Unpack converts byte slice into string
func Unpack(in []byte, m *string, s *uint64) error {
	*m = string(in[:len(in)-8])
	*s = binary.LittleEndian.Uint64(in[len(in)-8:])

	return nil
}
