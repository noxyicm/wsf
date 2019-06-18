package response

import (
	"fmt"
	"net/http"
	"strings"
	"wsf/errors"
	"wsf/utils/stack"
)

// HTTP response struct
type HTTP struct {
	Headers                map[string][]string
	Cookies                map[string]*http.Cookie
	Code                   int
	Body                   *stack.Referenced
	Datamap                map[string]interface{}
	Excpts                 []error
	RenderExceptions       bool
	Redirected             bool
	SendRequested          bool
	InsideErrorHandlerLoop bool
	Writer                 http.ResponseWriter
}

// SetHeader sets response header
func (r *HTTP) SetHeader(key string, value string) error {
	key = r.normalizeHeader(key)
	r.Headers[key] = []string{value}
	return nil
}

// AddHeader adds a response header
func (r *HTTP) AddHeader(key string, value string) error {
	key = r.normalizeHeader(key)
	if _, ok := r.Headers[key]; !ok {
		r.Headers[key] = make([]string, 0)
	}

	r.Headers[key] = append(r.Headers[key], value)
	return nil
}

// ClearHeaders removes all headers from stack
func (r *HTTP) ClearHeaders() error {
	if len(r.Headers) == 0 {
		return nil
	}

	r.Headers = make(map[string][]string)
	return nil
}

// RemoveHeader removes a header from stack
func (r *HTTP) RemoveHeader(key string) error {
	if len(r.Headers) == 0 {
		return nil
	}

	for k := range r.Headers {
		if key == k {
			delete(r.Headers, k)
		}
	}

	return nil
}

// SetBody sets the response body
func (r *HTTP) SetBody(b []byte) error {
	r.Body.Clear()
	return r.Body.Append("default", b)
}

// AppendBody adds bytes to body
func (r *HTTP) AppendBody(data []byte, segment string) error {
	if segment == "" {
		segment = "default"
	}

	if v := r.Body.Value(segment); v != nil {
		return r.Body.Set(segment, append(v.([]byte), data...))
	}

	return r.Body.Append(segment, data)
}

// ClearBody body stack
func (r *HTTP) ClearBody() error {
	return r.Body.Clear()
}

// ClearBodySegment clears segment
func (r *HTTP) ClearBodySegment(segment string) error {
	return r.Body.Unset(segment)
}

// ContentLength returns number of bytes of response body
func (r *HTTP) ContentLength() int {
	l := 0
	for _, segment := range r.Body.Stack() {
		l = l + len(segment.([]byte))
	}

	return l
}

// AddCookies adds response cookies
func (r *HTTP) AddCookies(cookies []*http.Cookie) {
	for _, cookie := range cookies {
		r.Cookies[cookie.Name] = cookie
	}
}

// AddCookie adds a cookie to response
func (r *HTTP) AddCookie(cookie *http.Cookie) {
	r.Cookies[cookie.Name] = cookie
}

// AddStringCookies creates a cookie fron string and adds to response
func (r *HTTP) AddStringCookies(cookies map[string]string) {
	for cookieName, cookieValue := range cookies {
		r.Cookies[cookieName] = &http.Cookie{
			Name:  cookieName,
			Value: cookieValue,
		}
	}
}

// AddStringCookie creates a cookie fron string and adds to response
func (r *HTTP) AddStringCookie(key string, value string) {
	r.Cookies[key] = &http.Cookie{
		Name:  key,
		Value: value,
	}
}

// Cookie returns a cookie
func (r *HTTP) Cookie(key string) string {
	if v, ok := r.Cookies[key]; ok {
		return v.Value
	}

	return ""
}

// SetRedirect sets response into redirect state
func (r *HTTP) SetRedirect(url string, code int) error {
	if code == 0 {
		code = 302
	}

	r.SetHeader("Location", url)
	return r.SetResponseCode(code)
}

// IsRedirect returns true if response should redirect
func (r *HTTP) IsRedirect() bool {
	return r.Redirected
}

// SetResponseCode sets a response code
func (r *HTTP) SetResponseCode(code int) error {
	if code < 100 || code > 599 {
		return errors.New("Invalid HTTP response code")
	}

	if code >= 300 && code <= 307 {
		r.Redirected = true
	} else {
		r.Redirected = false
	}

	r.Code = code
	return nil
}

// ResponseCode returns status code of the response
func (r *HTTP) ResponseCode() int {
	return r.Code
}

// SetException adds an error to the response
func (r *HTTP) SetException(err error) {
	r.Excpts = append(r.Excpts, err)
}

// Exceptions returns all errors
func (r *HTTP) Exceptions() []error {
	return r.Excpts
}

// IsException returns true if response has errors
func (r *HTTP) IsException() bool {
	return len(r.Excpts) > 0
}

// SetData adds a value into response data
func (r *HTTP) SetData(key string, value interface{}) {
	r.Datamap[key] = value
}

// GetData returns specified data
func (r *HTTP) GetData(key string) interface{} {
	if v, ok := r.Datamap[key]; ok {
		return v
	}

	return nil
}

// Data returns all registered response data
func (r *HTTP) Data() map[string]interface{} {
	return r.Datamap
}

// IsSendRequested returns true if there was request to send response emmidietly
func (r *HTTP) IsSendRequested() bool {
	return r.SendRequested
}

// RequestSend sets the response in emmidiet send state
func (r *HTTP) RequestSend(flag bool) {
	r.SendRequested = flag
}

// Write writes response headers, status and body into ResponseWriter
func (r *HTTP) Write() error {
	cookies := make([]string, len(r.Cookies))
	i := 0
	for _, v := range r.Cookies {
		cookies[i] = v.String()
		i++
	}
	cookiestr := strings.Join(cookies, "&")

	if _, ok := r.Headers["Set-Cookie"]; ok {
		if r.Headers["Set-Cookie"][0] == "" {
			r.Headers["Set-Cookie"][0] = cookiestr
		} else {
			r.Headers["Set-Cookie"][0] = r.Headers["Set-Cookie"][0] + "&" + cookiestr
		}
	}

	for n, h := range r.Headers {
		for _, v := range h {
			if n == "http2-push" {
				if pusher, ok := r.Writer.(http.Pusher); ok {
					pusher.Push(v, nil)
				}

				continue
			}

			r.Writer.Header().Add(n, v)
		}
	}

	r.Writer.WriteHeader(r.Code)

	if r.IsException() && r.RenderExceptions {
		exceptions := []byte{}
		for _, e := range r.Excpts {
			exceptions = append(exceptions, []byte(e.Error()+"\n")...)
		}

		r.Writer.Write(exceptions)
		return nil
	}

	b := []byte{}
	for _, value := range r.Body.Stack() {
		if data, ok := value.([]byte); ok {
			b = append(b, data...)
		} else {
			fmt.Println("warning")
			// need warning
		}
	}
	r.Writer.Write(b)

	//if rc, ok := r.body.(io.Reader); ok {
	//	if _, err := io.Copy(r.writer, rc); err != nil {
	//		return err
	//	}
	//}

	return nil
}

// Destroy the response
func (r *HTTP) Destroy() {
	r.Code = 500
	r.Headers = make(map[string][]string)
	r.Cookies = make(map[string]*http.Cookie)
	r.Body = stack.NewReferenced()
	r.Datamap = make(map[string]interface{})
	r.Writer = nil
}

func (r *HTTP) normalizeHeader(name string) string {
	filtered := strings.ReplaceAll(strings.ReplaceAll(name, "-", " "), "_", " ")
	filtered = strings.ToLower(filtered)
	filtered = strings.ReplaceAll(filtered, " ", "-")
	return filtered
}

// NewHTTPResponse creates new response based on given roadrunner payload
func NewHTTPResponse(w http.ResponseWriter) (Interface, error) {
	return &HTTP{
		Code:    500,
		Headers: make(map[string][]string),
		Cookies: make(map[string]*http.Cookie),
		Body:    stack.NewReferenced(),
		Datamap: make(map[string]interface{}),
		Writer:  w,
	}, nil
}
