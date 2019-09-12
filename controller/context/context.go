package context

import (
	"context"
	"time"
	"wsf/controller/request"
	"wsf/controller/response"
)

// Key is a key
type Key int64

// Puublic vars
var (
	ContextData   Key
	Session       Key = 1
	SessionID     Key = 2
	Layout        Key = 3
	LayoutEnabled Key = 4
)

// Context is controller context interface
type Context interface {
	context.Context

	SetValue(key interface{}, value interface{}) error
	SetDataValue(key string, value interface{}) error
	DataValue(key string) interface{}
	Data() map[string]interface{}
	SetRequest(req request.Interface) error
	Request() request.Interface
	SetResponse(rsp response.Interface) error
	Response() response.Interface
	Destroy()
}

// RequestContext is a request specific data
type RequestContext struct {
	context  context.Context
	cancel   context.CancelFunc
	request  request.Interface
	response response.Interface
}

// Deadline is part of context.Context interface
func (c *RequestContext) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}

// Done is part of context.Context interface
func (c *RequestContext) Done() <-chan struct{} {
	return c.context.Done()
}

// Err is part of context.Context interface
func (c *RequestContext) Err() error {
	return c.context.Err()
}

// Value is part of context.Context interface
func (c *RequestContext) Value(key interface{}) interface{} {
	return c.context.Value(key)
}

// SetValue injects a value into context
func (c *RequestContext) SetValue(key interface{}, value interface{}) error {
	c.context = context.WithValue(c.context, key, value)
	return nil
}

// SetRequest sets context request
func (c *RequestContext) SetRequest(req request.Interface) error {
	c.request = req
	c.context = req.Context()
	return nil
}

// Request returns context request
func (c *RequestContext) Request() request.Interface {
	return c.request
}

// SetResponse sets context response
func (c *RequestContext) SetResponse(rsp response.Interface) error {
	c.response = rsp
	return nil
}

// Response returns context response
func (c *RequestContext) Response() response.Interface {
	return c.response
}

// SetDataValue injects a data value into context
func (c *RequestContext) SetDataValue(key string, value interface{}) error {
	var d map[string]interface{}
	sd := c.context.Value(ContextData)
	if v, ok := sd.(map[string]interface{}); ok {
		d = v
	} else {
		d = make(map[string]interface{})
	}

	d[key] = value
	c.context = context.WithValue(c.context, ContextData, d)
	return nil
}

// DataValue returns a stored data value
func (c *RequestContext) DataValue(key string) interface{} {
	d := c.context.Value(ContextData)
	if v, ok := d.(map[string]interface{}); ok {
		if v, ok := v[key]; ok {
			return v
		}
	}

	return nil
}

// Data return a stored data
func (c *RequestContext) Data() map[string]interface{} {
	d := c.context.Value(ContextData)
	if v, ok := d.(map[string]interface{}); ok {
		return v
	}

	return nil
}

// Destroy the context
func (c *RequestContext) Destroy() {
	c.response.Destroy()
	c.request.Destroy()
}

// NewContext creates new request specific context
func NewContext(ctx context.Context) (Context, error) {
	c := &RequestContext{
		context: ctx,
	}

	if c.context == nil {
		c.context = context.Background()
	}

	return c, nil
}
