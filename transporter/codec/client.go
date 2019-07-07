package codec

import (
	"encoding/json"
	"io"
	"net/rpc"
	"reflect"
	"wsf/errors"
	"wsf/transporter"
)

// Client is a response handler for socket transporter
type Client struct {
	transporter transporter.Interface
	closed      bool
}

// WriteRequest writes request to the connection
func (c *Client) WriteRequest(r *rpc.Request, body interface{}) error {
	if err := c.transporter.Send(transporter.Pack(r.ServiceMethod, r.Seq), transporter.Control|transporter.Raw); err != nil {
		return err
	}

	if bin, ok := body.(*[]byte); ok {
		return c.transporter.Send(*bin, transporter.Raw)
	}

	if bin, ok := body.([]byte); ok {
		return c.transporter.Send(bin, transporter.Raw)
	}

	packed, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return c.transporter.Send(packed, 0)
}

// ReadResponseHeader reads response from the connection.
func (c *Client) ReadResponseHeader(r *rpc.Response) error {
	data, h, err := c.transporter.Receive()
	if err != nil {
		return err
	}

	if !h.HasFlag(transporter.Control) {
		return errors.New("Invalid RPC header, control flag is missing")
	}

	if !h.HasFlag(transporter.Raw) {
		return errors.New("RPC response header must be binary data")
	}

	if !h.HasPayload() {
		return errors.New("RPC response header can't be empty")
	}

	return transporter.Unpack(data, &r.ServiceMethod, &r.Seq)
}

// ReadResponseBody response from the connection
func (c *Client) ReadResponseBody(out interface{}) error {
	data, h, err := c.transporter.Receive()
	if err != nil {
		return err
	}

	if out == nil {
		return nil
	}

	if !h.HasPayload() {
		return nil
	}

	if h.HasFlag(transporter.Error) {
		return errors.New(string(data))
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

// Close closes the client connection
func (c *Client) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	return c.transporter.Close()
}

// NewClient initiates new server rpc codec over socket connection
func NewClient(rwc io.ReadWriteCloser) *Client {
	return &Client{transporter: transporter.NewSocket(rwc)}
}
