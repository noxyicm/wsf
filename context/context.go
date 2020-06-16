package context

import (
	"context"
	"errors"
	"sync"
	"time"
	"wsf/controller/request"
	"wsf/controller/response"
)

// Key is a key
type Key int64

// Puublic context keys
var (
	ContextData Key = 0
	//LayoutKey        Key = 1
	//LayoutEnabledKey Key = 2
	SessionKey      Key = 3
	SessionIDKey    Key = 4
	RowConfigKey    Key = 5
	RowsetConfigKey Key = 6
	NoRenderKey     Key = 7
	AuthIdentityKey Key = 8

	LayoutKey        string = "layout"
	LayoutEnabledKey string = "layoutEnabled"
	IdentityKey      string = "auth.identity"
	CredentialKey    string = "auth.credential"
)

// Context is controller context interface
type Context interface {
	context.Context

	SetValue(key interface{}, value interface{}) error
	SetDataValue(key string, value interface{}) error
	DataValue(key string) interface{}
	Data() map[string]interface{}
	SetParam(key string, value interface{}) error
	Param(key string) interface{}
	ParamBool(key string) bool
	ParamString(key string) string
	ParamInt(key string) int
	Params() map[string]interface{}
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
	params   map[string]interface{}
	data     map[string]interface{}
	mu       sync.Mutex
}

// WithCancel returns a new context with cancel function
func WithCancel(parent Context) (ctx Context, cancelFunc context.CancelFunc) {
	c, cFunc := context.WithCancel(parent)
	return &DefaultContext{context: c, cancel: cFunc, params: parent.Params(), data: parent.Data()}, cFunc
}

// WithDeadline returns a new context with deadline and cancel function
func WithDeadline(parent Context, d time.Time) (ctx Context, cancelFunc context.CancelFunc) {
	c, cFunc := context.WithDeadline(parent, d)
	return &DefaultContext{context: c, cancel: cFunc, params: parent.Params(), data: parent.Data()}, cFunc
}

// WithTimeout returns a new context with timeout and cancel function
func WithTimeout(parent Context, timeout time.Duration) (ctx Context, cancelFunc context.CancelFunc) {
	c, cFunc := context.WithTimeout(parent, timeout)
	return &DefaultContext{context: c, cancel: cFunc, params: parent.Params(), data: parent.Data()}, cFunc
}

// Background returns not-nil, empty context
func Background() context.Context {
	return &DefaultContext{context: context.Background(), cancel: nil, params: make(map[string]interface{}), data: make(map[string]interface{})}
}

// Deadline is part of context.Context interface
func (c *DefaultContext) Deadline() (deadline time.Time, ok bool) {
	return c.context.Deadline()
}

// Done is part of context.Context interface
func (c *DefaultContext) Done() <-chan struct{} {
	select {
	default:
	case <-c.context.Done():
		return c.context.Done()
	}

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

// SetDataValue injects a data value into context
func (c *DefaultContext) SetDataValue(key string, value interface{}) error {
	/*var d map[string]interface{}
	sd := c.context.Value(ContextData)
	if v, ok := sd.(map[string]interface{}); ok {
		d = v
	} else {
		d = make(map[string]interface{})
	}

	d[key] = value
	c.context = context.WithValue(c.context, ContextData, d)*/
	if _, ok := c.data[key]; ok {
		return errors.New("Overloading of existing data keys is not allowed")
	}

	c.data[key] = value
	return nil
}

// DataValue returns a stored data value
func (c *DefaultContext) DataValue(key string) interface{} {
	/*d := c.context.Value(ContextData)
	if v, ok := d.(map[string]interface{}); ok {
		if v, ok := v[key]; ok {
			return v
		}
	}

	return nil*/
	if v, ok := c.data[key]; ok {
		return v
	}

	return nil
}

// Data return a stored data
func (c *DefaultContext) Data() map[string]interface{} {
	/*d := c.context.Value(ContextData)
	if v, ok := d.(map[string]interface{}); ok {
		return v
	}

	return nil*/
	return c.data
}

// SetRequest sets context request
func (c *DefaultContext) SetRequest(req request.Interface) error {
	c.request = req
	c.context = req.Context()
	c.cancel = nil
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

// SetParam injects a data value into context
func (c *DefaultContext) SetParam(key string, value interface{}) error {
	c.params[key] = value
	return nil
}

// Param returns a stored data value
func (c *DefaultContext) Param(key string) interface{} {
	if v, ok := c.params[key]; ok {
		return v
	}

	return nil
}

// ParamBool returns a stored data value as boolean
func (c *DefaultContext) ParamBool(key string) bool {
	if v, ok := c.params[key]; ok {
		if v, ok := v.(bool); ok {
			return v
		}
	}

	return false
}

// ParamString returns a stored data value as string
func (c *DefaultContext) ParamString(key string) string {
	if v, ok := c.params[key]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	}

	return ""
}

// ParamInt returns a stored data value as string
func (c *DefaultContext) ParamInt(key string) int {
	if v, ok := c.params[key]; ok {
		if v, ok := v.(int); ok {
			return v
		}
	}

	return 0
}

// Params return a stored data
func (c *DefaultContext) Params() map[string]interface{} {
	return c.params
}

// Destroy the context
func (c *DefaultContext) Destroy() {
	c.response.Destroy()
	c.request.Destroy()

	c.context = nil
	c.cancel = nil
	c.request = nil
	c.response = nil
	c.params = make(map[string]interface{})
	c.data = make(map[string]interface{})
}

// Cancel context
func (c *DefaultContext) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}

// NewContext creates new request specific context
func NewContext(ctx context.Context) (Context, error) {
	c := &DefaultContext{
		params: make(map[string]interface{}),
		data:   make(map[string]interface{}),
	}

	if ctx == nil {
		ctx = context.Background()
	}

	c.context, c.cancel = context.WithCancel(ctx)
	return c, nil
}
