package transporter

import (
	"io"
	"sync"

	"github.com/pkg/errors"
)

// Socket transporter type
type Socket struct {
	BufferSize uint32
	muw        sync.Mutex
	mur        sync.Mutex
	rwc        io.ReadWriteCloser
}

// Send signed (prefixed) data to PHP process.
func (s *Socket) Send(data []byte, flags byte) error {
	s.muw.Lock()
	defer s.muw.Unlock()

	header := NewHeader().SetFlags(flags).SetSize(uint32(len(data)))
	if _, err := s.rwc.Write(header[:]); err != nil {
		return err
	}

	if _, err := s.rwc.Write(data); err != nil {
		return err
	}

	return nil
}

// Receive data from the underlying process and returns associated prefix or error.
func (s *Socket) Receive() (data []byte, h Header, err error) {
	s.mur.Lock()
	defer s.mur.Unlock()

	defer func() {
		if recoverable, ok := recover().(error); ok {
			err = recoverable
		}
	}()

	if _, err := s.rwc.Read(h[:]); err != nil {
		return nil, h, err
	}

	if !h.Valid() {
		return nil, h, errors.New("Payload is not protocol-wise")
	}

	if !h.HasPayload() {
		return nil, h, nil
	}

	bytesToRead := h.Size()
	data = make([]byte, 0, bytesToRead)
	buffer := make([]byte, minu32(uint32(cap(data)), s.BufferSize))

	for {
		if n, err := s.rwc.Read(buffer); err == nil {
			data = append(data, buffer[:n]...)
			bytesToRead -= uint32(n)
		} else {
			return nil, h, err
		}

		if bytesToRead == 0 {
			break
		}
	}

	return data, h, nil
}

// Close the connection
func (s *Socket) Close() error {
	return s.rwc.Close()
}

// NewSocket creates new socket based data transporter
func NewSocket(rwc io.ReadWriteCloser) *Socket {
	return &Socket{BufferSize: BufferSize, rwc: rwc}
}
