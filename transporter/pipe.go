package transporter

import (
	"io"
	"sync"

	"github.com/pkg/errors"
)

// Pipe based transporter
type Pipe struct {
	BufferSize uint32
	in         io.ReadCloser
	out        io.WriteCloser
	mur        sync.Mutex
	muw        sync.Mutex
}

// Send data
func (p *Pipe) Send(data []byte, flags byte) (err error) {
	p.muw.Lock()
	defer p.muw.Unlock()

	header := NewHeader().SetFlags(flags).SetSize(uint32(len(data)))
	if _, err := p.out.Write(append(header[:], data...)); err != nil {
		return err
	}

	return nil
}

// Receive data
func (p *Pipe) Receive() (data []byte, h Header, err error) {
	p.mur.Lock()
	defer p.mur.Unlock()

	defer func() {
		if recoverable, ok := recover().(error); ok {
			err = recoverable
		}
	}()

	if _, err := p.in.Read(h[:]); err != nil {
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
	buffer := make([]byte, minu32(uint32(cap(data)), p.BufferSize))

	for {
		if n, err := p.in.Read(buffer); err == nil {
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
func (p *Pipe) Close() error {
	return nil
}

// NewPipe creates new pipe based data transporter.
func NewPipe(in io.ReadCloser, out io.WriteCloser) *Pipe {
	return &Pipe{BufferSize: BufferSize, in: in, out: out}
}
