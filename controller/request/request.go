package request

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"github.com/noxyicm/wsf/utils"
)

// Interface represents a net/http request maped to PSR7 compatible structure
type Interface interface {
	Context() context.Context
	SetContext(ctx context.Context)
	SetParam(name string, value interface{}) error
	Param(name string) interface{}
	ParamString(name string) string
	ParamInt(name string) int
	ParamBool(name string) bool
	ParamStringDefault(name string, d string) string
	ParamIntDefault(name string, d int) int
	ParamBoolDefault(name string, d bool) bool
	Params() map[string]interface{}
	IsDispatched() bool
	SetDispatched(bool) error
	ModuleKey() string
	SetModuleName(string)
	ModuleName() string
	ControllerKey() string
	SetControllerName(string)
	ControllerName() string
	ActionKey() string
	SetActionName(string)
	ActionName() string
	SetPathInfo(path string) error
	PathInfo() string
	Upload()
	Clear() error
	Destroy()
	IsSecure() bool
	AddCookie(cookie *http.Cookie)
	Cookie(key string) string
	RawCookie(key string) *http.Cookie
	Cookies() map[string]*http.Cookie
	AddHeader(name string, value string)
	RemoveHeader(name string)
	Header(name string) string
	SetSessionID(sid string)
	SessionID() string
	RemoteAddress() string
}

// Request is a abstrackt request struct
type Request struct {
	Path       string
	Proxyed    bool
	Dispatched bool
	MdlKey     string
	Mdl        string
	CtrlKey    string
	Ctrl       string
	ActKey     string
	Act        string
	SessID     string
	Secure     bool
	Prms       map[string]interface{}
	Cks        map[string]*http.Cookie
	RemoteAddr string
	Body       interface{}
}

// SetParam sets request parameter
func (r *Request) SetParam(name string, value interface{}) error {
	r.Prms[name] = value
	return nil
}

// Param returns request parameter
func (r *Request) Param(name string) interface{} {
	if v, ok := r.Prms[name]; ok {
		return v
	} else if b, ok := r.Body.(utils.DataTree); ok {
		return b.Get(name)
	}

	return nil
}

// ParamString returns request parameter as string
func (r *Request) ParamString(name string) string {
	return r.ParamStringDefault(name, "")
}

// ParamInt returns request parameter as int
func (r *Request) ParamInt(name string) int {
	return r.ParamIntDefault(name, 0)
}

// ParamBool returns request parameter as bool
func (r *Request) ParamBool(name string) bool {
	return r.ParamBoolDefault(name, false)
}

// ParamStringDefault returns request parameter as string or d
func (r *Request) ParamStringDefault(name string, d string) string {
	if v, ok := r.Prms[name]; ok {
		if v, ok := v.(string); ok {
			return v
		}
	} else if b, ok := r.Body.(utils.DataTree); ok {
		vi := b.Get(name)
		switch v := vi.(type) {
		case []string:
			if len(v) > 0 {
				return v[0]
			}
			break

		case string:
			return v
		}
	}

	return d
}

// ParamIntDefault returns request parameter as int or d
func (r *Request) ParamIntDefault(name string, d int) int {
	if v, ok := r.Prms[name]; ok {
		if i, ok := v.(int); ok {
			return i
		} else if s, err := strconv.Atoi(v.(string)); err == nil {
			return s
		}
	} else if b, ok := r.Body.(utils.DataTree); ok {
		vi := b.Get(name)
		switch v := vi.(type) {
		case []string:
			if len(v) > 0 {
				if ret, err := strconv.Atoi(v[0]); err == nil {
					return ret
				}
			}
			break

		case string:
			if ret, err := strconv.Atoi(v); err == nil {
				return ret
			}
			break

		case []float64:
			if len(v) > 0 {
				return int(v[0])
			}
			break

		case float64:
			return int(v)
		}
	}

	return d
}

// ParamBoolDefault returns request parameter as bool or d
func (r *Request) ParamBoolDefault(name string, d bool) bool {
	if v, ok := r.Prms[name]; ok {
		if b, ok := v.(bool); ok {
			return b
		} else if s, err := strconv.ParseBool(v.(string)); err == nil {
			return s
		}
	} else if b, ok := r.Body.(utils.DataTree); ok {
		vi := b.Get(name)
		switch v := vi.(type) {
		case []string:
			if len(v) > 0 {
				if ret, err := strconv.ParseBool(v[0]); err == nil {
					return ret
				}
			}
			break

		case string:
			if ret, err := strconv.ParseBool(v); err == nil {
				return ret
			}
			break
		}
	}

	return d
}

// ParamFloatDefault returns request parameter as int or d
func (r *Request) ParamFloatDefault(name string, d float64) float64 {
	if v, ok := r.Prms[name]; ok {
		if f, ok := v.(float64); ok {
			return f
		} else if s, err := strconv.ParseFloat(v.(string), 64); err == nil {
			return s
		}
	} else if b, ok := r.Body.(utils.DataTree); ok {
		vi := b.Get(name)
		switch v := vi.(type) {
		case []string:
			if len(v) > 0 {
				if ret, err := strconv.ParseFloat(v[0], 64); err == nil {
					return ret
				}
			}
			break

		case string:
			if ret, err := strconv.ParseFloat(v, 64); err == nil {
				return ret
			}
			break
		}
	}

	return d
}

// ParamMapDefault returns request parameter as int or d
func (r *Request) ParamMapDefault(name string, d map[string]interface{}) map[string]interface{} {
	if v, ok := r.Prms[name]; ok {
		if v, ok := v.(map[string]interface{}); ok {
			return v
		}
	} else if b, ok := r.Body.(utils.DataTree); ok {
		switch v := b.Get(name).(type) {
		case []map[string]interface{}:
			if len(v) > 0 {
				return v[0]
			}

		case map[string]interface{}:
			return v

		case utils.DataTree:
			return utils.MapFromDataTree(v)
		}
	}

	return d
}

// Params returns request parameters
func (r *Request) Params() map[string]interface{} {
	p := make(map[string]interface{})
	if b, ok := r.Body.(utils.DataTree); ok {
		for k, v := range b {
			p[k] = v
		}
	}

	for k, v := range r.Prms {
		p[k] = v
	}

	return p
}

// ModuleKey returns module key
func (r *Request) ModuleKey() string {
	return r.MdlKey
}

// SetModuleName sets module name for request
func (r *Request) SetModuleName(s string) {
	r.Mdl = s
	return
}

// ModuleName returns module name specified for request
func (r *Request) ModuleName() string {
	return r.Mdl
}

// ControllerKey returns controller key
func (r *Request) ControllerKey() string {
	return r.CtrlKey
}

// SetControllerName sets controller name for request
func (r *Request) SetControllerName(s string) {
	r.Ctrl = s
	return
}

// ControllerName returns controller name specified for request
func (r *Request) ControllerName() string {
	return r.Ctrl
}

// ActionKey returns action key
func (r *Request) ActionKey() string {
	return r.ActKey
}

// SetActionName sets action name for request
func (r *Request) SetActionName(s string) {
	r.Act = s
	return
}

// ActionName returns action name specified for request
func (r *Request) ActionName() string {
	return r.Act
}

// SetPathInfo sets request path
func (r *Request) SetPathInfo(path string) error {
	r.Path = path
	return nil
}

// PathInfo returns request path
func (r *Request) PathInfo() string {
	return r.Path
}

// IsSecure returns true if request made throught secure chanel
func (r *Request) IsSecure() bool {
	return r.Secure
}

// AddCookie adds a cookie to request
func (r *Request) AddCookie(cookie *http.Cookie) {
	if cookie == nil {
		return
	}

	r.Cks[cookie.Name] = cookie
}

// Cookie returns a cookie value by key
func (r *Request) Cookie(key string) string {
	if v, ok := r.Cks[key]; ok {
		if c, err := url.QueryUnescape(v.Value); err == nil {
			return c
		}
	}

	return ""
}

// RawCookie returns a Cookie by key
func (r *Request) RawCookie(key string) *http.Cookie {
	if v, ok := r.Cks[key]; ok {
		return v
	}

	return nil
}

// Cookies returns all cookies
func (r *Request) Cookies() map[string]*http.Cookie {
	return r.Cks
}

// SetSessionID sets a session id for this request
func (r *Request) SetSessionID(sid string) {
	r.SessID = sid
}

// SessionID returns a session id asocieted with this request
func (r *Request) SessionID() string {
	return r.SessID
}

// RemoteAddress returns a remote addres of request
func (r *Request) RemoteAddress() string {
	return r.RemoteAddr
}

// uri fetches full uri from request in a form of string (including https scheme if TLS connection is enabled)
func uri(r *http.Request) string {
	if r.TLS != nil {
		return fmt.Sprintf("https://%s%s", r.Host, r.URL.String())
	}

	return fmt.Sprintf("http://%s%s", r.Host, r.URL.String())
}
