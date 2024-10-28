package codec

import (
	"encoding/json"
	"io"
	"net/rpc"
	"reflect"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/transporter"
)

// Server is a request handler for socket transporter
type Server struct {
	transporter transporter.Interface
	closed      bool
}

// ReadRequestHeader receives RPC request
func (s *Server) ReadRequestHeader(r *rpc.Request) error {
	data, h, err := s.transporter.Receive()
	if err != nil {
		return err
	}

	if !h.HasFlag(transporter.Control) {
		return errors.New("RPC response header must be control header")
	}

	if !h.HasFlag(transporter.Raw) {
		return errors.New("RPC response header must be binary data")
	}

	if !h.HasPayload() {
		return errors.New("RPC request header can't be empty")
	}

	return transporter.Unpack(data, &r.ServiceMethod, &r.Seq)
}

// ReadRequestBody unmarshals request body into json
func (s *Server) ReadRequestBody(out interface{}) error {
	data, h, err := s.transporter.Receive()
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	if !h.HasPayload() {
		return nil
	}

	if h.HasFlag(transporter.Raw) {
		if bin, ok := out.(*[]byte); ok {
			*bin = append(*bin, data...)
			return nil
		}

		return errors.New("Binary data request for " + reflect.ValueOf(out).String())
	}

	return json.Unmarshal(data, out)
}

// WriteResponse marshals response, byte slice or error and send it
func (s *Server) WriteResponse(r *rpc.Response, body interface{}) error {
	if err := s.transporter.Send(transporter.Pack(r.ServiceMethod, r.Seq), transporter.Control|transporter.Raw); err != nil {
		return err
	}

	if r.Error != "" {
		return s.transporter.Send([]byte(r.Error), transporter.Error|transporter.Raw)
	}

	if bin, ok := body.(*[]byte); ok {
		return s.transporter.Send(*bin, transporter.Raw)
	}

	if bin, ok := body.([]byte); ok {
		return s.transporter.Send(bin, transporter.Raw)
	}

	packed, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return s.transporter.Send(packed, 0)
}

// Close underlying socket
func (s *Server) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true
	return s.transporter.Close()
}

// NewServer initiates new server rpc codec over socket connection
func NewServer(rwc io.ReadWriteCloser) *Server {
	return &Server{transporter: transporter.NewSocket(rwc)}
}
