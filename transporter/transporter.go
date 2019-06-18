package transporter

const (
	// BufferSize defines max amount of bytes to read from connection at once.
	BufferSize = 655336

	// MemoryType represents direct struct to struct transporter
	MemoryType = "memory"

	// PipeType represents transporter over system pipes
	PipeType = "pipe"

	// SocketType represents transporter over system sockets
	SocketType = "socket"
)

// Interface provides data payload transfer
type Interface interface {
	Send(data []byte, flags byte) (err error)
	Receive() (data []byte, h Header, err error)
	Close() error
}
