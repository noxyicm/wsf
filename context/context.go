package context

import (
	"context"
	"time"
	"wsf/controller/request"
	"wsf/controller/response"
)

// Key is a key
type Key int64

// Puublic context keys
var (
	ContextData      Key = 0
	LayoutKey        Key = 1
	LayoutEnabledKey Key = 2
	SessionKey       Key = 3
	SessionIDKey     Key = 4
	RowConfigKey     Key = 5
	RowsetConfigKey  Key = 6
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
	Cancel()
}

// DefaultContext is a request specific data
type DefaultContext struct {
	context  context.Context
	cancel   context.CancelFunc
	request  request.Interface
	response response.Interface
}

// WithCancel returns a new context with cancel function
func WithCancel(parent Context) (ctx Context, cancelFunc context.CancelFunc) {
	c, cFunc := context.WithCancel(parent)
	return c.(Context), cFunc
}

// WithDeadline returns a new context with deadline and cancel function
func WithDeadline(parent Context, d time.Time) (ctx Context, cancelFunc context.CancelFunc) {
	c, cFunc := context.WithDeadline(parent, d)
	return c.(Context), cFunc
}

// WithTimeout returns a new context with timeout and cancel function
func WithTimeout(parent Context, timeout time.Duration) (ctx Context, cancelFunc context.CancelFunc) {
	c, cFunc := context.WithTimeout(parent, timeout)
	return c.(Context), cFunc
}

// Background returns not-nil, empty context
func Background() context.Context {
	return context.Background()
}

// Deadline is part of context.Context interface
func (c *DefaultContext) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}

// Done is part of context.Context interface
func (c *DefaultContext) Done() <-chan struct{} {
	return c.context.Done()
}

// Err is part of context.Context interface
func (c *DefaultContext) Err() error {
	return c.context.Err()
}

// Value is part of context.Context interface
func (c *DefaultContext) Value(key interface{}) interface{} {
	return c.context.Value(key)
}

// SetValue injects a value into context
func (c *DefaultContext) SetValue(key interface{}, value interface{}) error {
	c.context = context.WithValue(c.context, key, value)
	return nil
}

// SetRequest sets context request
func (c *DefaultContext) SetRequest(req request.Interface) error {
	c.request = req
	c.context = req.Context()
	return nil
}

// Request returns context request
func (c *DefaultContext) Request() request.Interface {
	return c.request
}

// SetResponse sets context response
func (c *DefaultContext) SetResponse(rsp response.Interface) error {
	c.response = rsp
	return nil
}

// Response returns context response
func (c *DefaultContext) Response() response.Interface {
	return c.response
}

// SetDataValue injects a data value into context
func (c *DefaultContext) SetDataValue(key string, value interface{}) error {
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
func (c *DefaultContext) DataValue(key string) interface{} {
	d := c.context.Value(ContextData)
	if v, ok := d.(map[string]interface{}); ok {
		if v, ok := v[key]; ok {
			return v
		}
	}

	return nil
}

// Data return a stored data
func (c *DefaultContext) Data() map[string]interface{} {
	d := c.context.Value(ContextData)
	if v, ok := d.(map[string]interface{}); ok {
		return v
	}

	return nil
}

// Destroy the context
func (c *DefaultContext) Destroy() {
	c.response.Destroy()
	c.request.Destroy()
}

// Cancel context
func (c *DefaultContext) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}

// NewContext creates new request specific context
func NewContext(ctx context.Context) (Context, error) {
	c := &DefaultContext{}

	if ctx == nil {
		ctx = context.Background()
	}

	c.context, c.cancel = context.WithCancel(ctx)
	return c, nil
}
