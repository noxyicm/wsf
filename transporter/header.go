package transporter

import (
	"encoding/binary"
	"fmt"
)

const (
	// Empty must be set when no data to be sent
	Empty byte = 2

	// Error must be set when data is error
	Error byte = 4

	// Raw must be set when data is binary data
	Raw byte = 8

	// Control must be set when data is a control header
	Control byte = 16
)

// Header is a header
type Header [9]byte

// String represents header as string
func (h *Header) String() string {
	return fmt.Sprintf("[%08b: %v]", h.Flags(), h.Size())
}

// Flags describe transmission behaviour and data type
func (h *Header) Flags() byte {
	return h[0]
}

// HasFlag returns true if prefix has given flag
func (h *Header) HasFlag(flag byte) bool {
	return h[0]&flag == flag
}

// Valid returns true if prefix is valid
func (h *Header) Valid() bool {
	return binary.LittleEndian.Uint32(h[1:5]) == binary.BigEndian.Uint32(h[5:])
}

// Size returns following data size in bytes
func (h *Header) Size() uint32 {
	if h.HasFlag(Empty) {
		return 0
	}

	return binary.LittleEndian.Uint32(h[1:5])
}

// HasPayload returns true if data is not empty
func (h *Header) HasPayload() bool {
	return h.Size() != 0
}

// SetFlag sets a flag for this header
func (h *Header) SetFlag(flag byte) *Header {
	h[0] = h[0] | flag
	return h
}

// SetFlags overwrites all flags in this header
func (h *Header) SetFlags(flags byte) *Header {
	h[0] = flags
	return h
}

// SetSize sets the size of payload to this header
func (h *Header) SetSize(size uint32) *Header {
	binary.LittleEndian.PutUint32(h[1:5], size)
	binary.BigEndian.PutUint32(h[5:], size)
	return h
}

// NewHeader creates new empty header with no flags and size.
func NewHeader() *Header {
	return new(Header)
}
