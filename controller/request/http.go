package request

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"github.com/noxyicm/wsf/application/file"
	"github.com/noxyicm/wsf/controller/request/attributes"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/utils"
)

const (
	defaultMaxMemory = 1 << 26 // 64 MB
	contentNone      = iota + 900
	contentStream
	contentJSON
	contentMultipart
	contentFormData
)

var braketsPattern = regexp.MustCompile(`\[(.)+\]`)

// HTTP maps net/http requests to PSR7 compatible structure and managed state of temporary uploaded files
type HTTP struct {
	*Request

	original   *http.Request
	Protocol   string                 `json:"protocol"`
	Method     string                 `json:"method"`
	URI        string                 `json:"uri"`
	RequestURI string                 `json:"requestUri"`
	Headers    http.Header            `json:"headers"`
	RawQuery   string                 `json:"rawQuery"`
	Referer    string                 `json:"referer"`
	UserAgent  string                 `json:"userAgent"`
	Parsed     bool                   `json:"parsed"`
	Uploads    *file.Transfer         `json:"uploads"`
	Attributes map[string]interface{} `json:"attributes"`
	FormData   utils.DataTree
	fileCfg    *file.Config
}

// Context returns request context
func (r *HTTP) Context() context.Context {
	return r.original.Context()
}

// SetContext sets request context
func (r *HTTP) SetContext(ctx context.Context) {
	r.original = r.original.WithContext(ctx)
}

// IsDispatched returns true if request was dispatched
func (r *HTTP) IsDispatched() bool {
	return r.Dispatched
}

// SetDispatched sets the state of dispatching
func (r *HTTP) SetDispatched(is bool) error {
	r.Dispatched = is
	return nil
}

// ParseMultipart parses request uploads into file transfer structure
func (r *HTTP) ParseMultipart(rqs *http.Request) error {
	if rqs.MultipartForm != nil {
		for k, v := range rqs.MultipartForm.Value {
			r.FormData.Push(k, v)
		}

		if len(rqs.MultipartForm.File) > 0 {
			t, err := file.NewTransfer(r.fileCfg)
			if err != nil {
				return err
			}

			var key string
			for k, v := range rqs.MultipartForm.File {
				for i, f := range v {
					m := braketsPattern.FindStringSubmatch(k)
					if len(m) > 0 {
						key = k
					} else {
						key = k + "[" + strconv.Itoa(i) + "]"
					}

					t.Push(key, file.NewUpload(f))
				}
			}

			if err := t.Upload(); err != nil {
				return err
			}

			for _, k := range t.Uploaded() {
				f := make([]map[string]interface{}, 0)
				if fd := t.Get(k); fd != nil {
					f = append(f, map[string]interface{}{
						"name":    fd.Name,
						"size":    fd.Size,
						"mime":    fd.Mime,
						"tmpName": fd.TempFilename,
					})

					r.FormData.Push(k, f)
				}
			}

			/*for k, v := range rqs.MultipartForm.File {
				f := make([]map[string]interface{}, len(v))
				for i := range v {
					if fd := t.Get(key); fd != nil {
						f[i] = map[string]interface{}{
							"name":    fd.Name,
							"size":    fd.Size,
							"mime":    fd.Mime,
							"tmpName": fd.TempFilename,
						}
					}
				}

				r.FormData.Push(k, f)
			}*/
		}
	}

	r.Body = r.FormData
	return nil
}

// ParseData parses request into data tree
func (r *HTTP) ParseData(rqs *http.Request) error {
	if rqs.PostForm != nil {
		for k, v := range rqs.PostForm {
			r.FormData.Push(k, v)
		}
	}

	r.Body = r.FormData
	return nil
}

// ParseJSONData parses JSON request into data tree
func (r *HTTP) ParseJSONData(encoded []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(encoded, &m); err != nil {
		return errors.Wrap(err, "Unable to parse request")
	}

	for k, v := range m {
		r.FormData.Push(k, v)
	}

	r.Body = r.FormData
	return nil
}

// Upload moves all uploaded files to temporary directory
func (r *HTTP) Upload() {
	if r.Uploads == nil {
		return
	}

	r.Uploads.Upload()
}

// Clear clears all temp file uploads
func (r *HTTP) Clear() error {
	if r.Uploads == nil {
		return nil
	}

	return r.Uploads.Clear()
}

// Destroy the request
func (r *HTTP) Destroy() {
	r.Dispatched = false
	r.Mdl = ""
	r.Ctrl = ""
	r.Act = ""
	r.Prms = make(map[string]interface{})
	r.Body = nil
	r.Path = ""
	r.Secure = false
	r.Cks = make(map[string]*http.Cookie)
	r.RemoteAddr = ""

	r.original = nil
	r.Protocol = ""
	r.Method = ""
	r.URI = ""
	r.RequestURI = ""
	r.Headers = nil
	r.RawQuery = ""
	r.Referer = ""
	r.UserAgent = ""
	r.Parsed = false
	r.Uploads = nil
	r.Attributes = make(map[string]interface{})
	r.fileCfg = nil
	r.FormData = make(utils.DataTree)

	r.Clear()
}

// AddCookie adds a cookie to request
func (r *HTTP) AddCookie(cookie *http.Cookie) {
	if cookie == nil {
		return
	}

	r.Cks[cookie.Name] = cookie
	r.original.AddCookie(cookie)
}

// IsPost Was the request made by POST?
func (r *HTTP) IsPost() bool {
	if r.Method == "POST" {
		return true
	}

	return false
}

// IsGet Was the request made by GET?
func (r *HTTP) IsGet() bool {
	if r.Method == "GET" {
		return true
	}

	return false
}

// IsPut Was the request made by PUT?
func (r *HTTP) IsPut() bool {
	if r.Method == "PUT" {
		return true
	}

	return false
}

// IsDelete Was the request made by DELETE?
func (r *HTTP) IsDelete() bool {
	if r.Method == "DELETE" {
		return true
	}

	return false
}

// IsHead Was the request made by HEAD?
func (r *HTTP) IsHead() bool {
	if r.Method == "HEAD" {
		return true
	}

	return false
}

// IsOptions Was the request made by OPTIONS?
func (r *HTTP) IsOptions() bool {
	if r.Method == "OPTIONS" {
		return true
	}

	return false
}

// IsPatch Was the request made by PATCH?
func (r *HTTP) IsPatch() bool {
	if r.Method == "PATCH" {
		return true
	}

	return false
}

// PostParam returns post parameter
func (r *Request) PostParam(name string) interface{} {
	if b, ok := r.Body.(utils.DataTree); ok {
		return b.Get(name)
	}

	return nil
}

// PostParamString returns post parameter as string
func (r *Request) PostParamString(name string) string {
	return r.PostParamStringDefault(name, "")
}

// PostParamInt returns post parameter as int
func (r *Request) PostParamInt(name string) int {
	return r.PostParamIntDefault(name, 0)
}

// PostParamBool returns post parameter as bool
func (r *Request) PostParamBool(name string) bool {
	return r.PostParamBoolDefault(name, false)
}

// PostParamStringDefault returns post parameter as string or d
func (r *Request) PostParamStringDefault(name string, d string) string {
	if b, ok := r.Body.(utils.DataTree); ok {
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

// PostParamIntDefault returns post parameter as int or d
func (r *Request) PostParamIntDefault(name string, d int) int {
	if b, ok := r.Body.(utils.DataTree); ok {
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

// PostParamBoolDefault returns post parameter as bool or d
func (r *Request) PostParamBoolDefault(name string, d bool) bool {
	if b, ok := r.Body.(utils.DataTree); ok {
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

// PostParamFloatDefault returns post parameter as int or d
func (r *Request) PostParamFloatDefault(name string, d float64) float64 {
	if b, ok := r.Body.(utils.DataTree); ok {
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

// PostParamMapDefault returns post parameter as int or d
func (r *Request) PostParamMapDefault(name string, d map[string]interface{}) map[string]interface{} {
	if b, ok := r.Body.(utils.DataTree); ok {
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

// PostParams returns post parameters
func (r *Request) PostParams() map[string]interface{} {
	p := make(map[string]interface{})
	if b, ok := r.Body.(utils.DataTree); ok {
		for k, v := range b {
			p[k] = v
		}
	}

	return p
}

// AddHeader adds a header to request
func (r *HTTP) AddHeader(name string, value string) {
	r.Headers.Add(name, value)
}

// RemoveHeader removes a header
func (r *HTTP) RemoveHeader(name string) {
	r.Headers.Del(name)
}

// Header returns specific request header
func (r *HTTP) Header(name string) string {
	return r.Headers.Get(name)
}

// GetHeaders returns request headers
func (r *HTTP) GetHeaders() http.Header {
	return r.Headers
}

// contentType returns the payload content type
func (r *HTTP) contentType() int {
	if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
		return contentNone
	}

	ct := r.Headers.Get("content-type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") {
		return contentFormData
	}

	if strings.Contains(ct, "multipart/form-data") {
		return contentMultipart
	}

	if strings.Contains(ct, "application/json") {
		return contentJSON
	}

	return contentStream
}

// clientIP method returns client IP address from HTTP request
//
// Note: Set property "app.behind.proxy" to true only if server is running
// behind proxy like nginx, haproxy, apache, etc. Otherwise
// you may get inaccurate Client IP address. Parses the
// IP address in the order of X-Forwarded-For, X-Real-IP.
//
// By default server will get http.Request's RemoteAddr
func (r *HTTP) clientIP() string {
	//if r.Proxyed {
	// Header X-Forwarded-For
	if fwdFor := strings.TrimSpace(r.original.Header.Get(http.CanonicalHeaderKey("X-Forwarded-For"))); fwdFor != "" {
		index := strings.Index(fwdFor, ",")
		if index == -1 {
			return fwdFor
		}
		return fwdFor[:index]
	}

	// Header X-Real-Ip
	if realIP := strings.TrimSpace(r.original.Header.Get(http.CanonicalHeaderKey("X-Real-Ip"))); realIP != "" {
		return realIP
	}
	//}

	if remoteAddr, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return remoteAddr
	}

	return ""
}

// NewHTTPRequest creates new PSR7 compatible request using net/http request
func NewHTTPRequest(r *http.Request, cfg *file.Config, proxyed bool) (ri Interface, err error) {
	req := &HTTP{
		Request:    &Request{},
		original:   r,
		Protocol:   r.Proto,
		Method:     r.Method,
		URI:        uri(r),
		RequestURI: r.URL.RequestURI(),
		Headers:    r.Header,
		RawQuery:   r.URL.RawQuery,
		Referer:    r.Referer(),
		UserAgent:  r.UserAgent(),
		Attributes: attributes.All(r),
		fileCfg:    cfg,
		FormData:   make(utils.DataTree),
	}

	req.Cks = make(map[string]*http.Cookie)
	req.MdlKey = "module"
	req.CtrlKey = "controller"
	req.ActKey = "action"
	req.Proxyed = proxyed
	req.Prms = make(map[string]interface{})
	req.RemoteAddr = req.clientIP()
	req.Secure = r.URL.Scheme == "https"

	if r.URL.RawPath != "" {
		req.Path = r.URL.RawPath
	} else {
		req.Path = r.URL.Path
	}

	for k, v := range r.URL.Query() {
		req.Prms[k] = v[0]
	}

	for _, c := range r.Cookies() {
		req.Cks[c.Name] = c
	}

	switch req.contentType() {
	case contentNone:
		return req, nil

	case contentStream:
		req.Body, err = ioutil.ReadAll(r.Body)
		return req, err

	case contentJSON:
		bd, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return req, errors.Wrap(err, "Unable to parse request")
		}

		if err = req.ParseJSONData(bd); err != nil {
			return req, errors.Wrap(err, "Unable to parse request")
		}
		break

	case contentMultipart:
		if err = r.ParseMultipartForm(defaultMaxMemory); err != nil {
			return nil, err
		}

		req.ParseMultipart(r)
		//fallthrough
	case contentFormData:
		if err = r.ParseForm(); err != nil {
			return nil, err
		}

		req.ParseData(r)
		break
	}

	req.Parsed = true
	return req, nil
}
