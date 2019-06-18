package context

import (
	"context"
	"wsf/controller/request"
	"wsf/controller/response"
)

// Context is a request specific data
type Context struct {
	context  context.Context
	cancel   context.CancelFunc
	request  request.Interface
	response response.Interface
}

// SetRequest sets context request
func (c *Context) SetRequest(req request.Interface) error {
	c.request = req
	c.context = req.Context()
	return nil
}

// Request returns context request
func (c *Context) Request() request.Interface {
	return c.request
}

// SetResponse sets context response
func (c *Context) SetResponse(rsp response.Interface) error {
	c.response = rsp
	return nil
}

// Response returns context response
func (c *Context) Response() response.Interface {
	return c.response
}

// SetParam sets context parameter
func (c *Context) SetParam(name string, value interface{}) error {
	//c.context = context.WithValue(c.context, name, value)
	return nil
}

// Destroy the context
func (c *Context) Destroy() {
	c.response.Destroy()
	c.request.Destroy()
	//if c.socket != nil {
	//	c.socket.Destroy()
	//	c.socket = nil
	//}
}

// NewContext creates new request specific context
func NewContext() (*Context, error) {
	return &Context{}, nil
}
