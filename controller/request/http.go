package request

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"wsf/application/file"
	"wsf/controller/request/attributes"
	"wsf/utils"
)

const (
	defaultMaxMemory = 1 << 26 // 64 MB
	contentNone      = iota + 900
	contentStream
	contentMultipart
	contentFormData
)

// HTTP maps net/http requests to PSR7 compatible structure and managed state of temporary uploaded files
type HTTP struct {
	Request
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

// ParseUploads parses request uploads into file transfer structure
func (r *HTTP) ParseUploads(rqs *http.Request) error {
	t, err := file.NewTransfer(r.fileCfg)
	if err != nil {
		return err
	}

	for k, v := range rqs.MultipartForm.File {
		files := make([]*file.File, 0, len(v))
		for _, f := range v {
			files = append(files, file.NewUpload(f))
		}

		t.Append(files)
		t.Push(k, files)
	}

	r.Uploads = t
	return nil
}

// ParseData parses request into data tree
func (r *HTTP) ParseData(rqs *http.Request) error {
	data := make(utils.DataTree)
	if rqs.PostForm != nil {
		for k, v := range rqs.PostForm {
			data.Push(k, v)
		}
	}

	if rqs.MultipartForm != nil {
		for k, v := range rqs.MultipartForm.Value {
			data.Push(k, v)
		}
	}

	r.Body = data
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

	case contentMultipart:
		if err = r.ParseMultipartForm(defaultMaxMemory); err != nil {
			return nil, err
		}

		req.ParseUploads(r)
		fallthrough
	case contentFormData:
		if err = r.ParseForm(); err != nil {
			return nil, err
		}

		req.ParseData(r)
	}

	req.Parsed = true
	return req, nil
}
