package request

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/noxyicm/wsf/application/file"
	"github.com/noxyicm/wsf/controller/request/attributes"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/utils"
)

const (
	defaultMaxMemory = 1 << 26 // 64 MB
	contentNone      = iota + 900
	contentStream
	contentJSON
	contentMultipart
	contentFormData
	contentApplication
	contentText
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
	Attributes map[string]interface{} `json:"attributes"`
	Form       utils.DataTree         `json:"form"`
	PostForm   utils.DataTree         `json:"postForm"`
	Uploads    file.TransferInterface
}

// ParseBody reads request body and parses it into structures
func (r *HTTP) ParseBody() (err error) {
	if r.Parsed {
		return nil
	}

	if err := r.parseForm(r.original); err != nil {
		return err
	}

	r.Parsed = true
	return nil
}

// Context returns request context
func (r *HTTP) Context() context.Context {
	return r.original.Context()
}

// SetContext sets request context
func (r *HTTP) SetContext(ctx context.Context) {
	r.original = r.original.WithContext(ctx)
}

// GetRequest returns underlying request
func (r *HTTP) GetRequest() *http.Request {
	return r.original
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
	return nil
}

// ParseData parses request into data tree
func (r *HTTP) ParseData(rqs *http.Request) error {
	return nil
}

// ParseJSONData parses JSON request into data tree
func (r *HTTP) ParseJSONData(encoded []byte) error {
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
	r.Form = nil
	r.PostForm = nil
	r.MaxRequestSize = 0
	r.MaxFormSize = 0

	r.Clear()
}

// AddCookie adds a cookie to request
func (r *HTTP) AddCookie(cookie *http.Cookie) {
	if cookie == nil {
		return
	}

	r.original.AddCookie(cookie)
}

// Cookie returns a cookie value by key
func (r *HTTP) Cookie(key string) string {
	if v, err := r.original.Cookie(key); err == nil {
		if c, err := url.QueryUnescape(v.Value); err == nil {
			return c
		}
	}

	return ""
}

// RawCookie returns a Cookie by key
func (r *HTTP) RawCookie(key string) *http.Cookie {
	if v, err := r.original.Cookie(key); err == nil {
		return v
	}

	return nil
}

// Cookies returns all cookies
func (r *HTTP) Cookies() map[string]*http.Cookie {
	m := make(map[string]*http.Cookie)
	for _, c := range r.original.Cookies() {
		m[c.Name] = c
	}

	return m
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
func (r *HTTP) PostParam(name string) interface{} {
	return r.PostForm.Get(name)
}

// PostParamString returns post parameter as string
func (r *HTTP) PostParamString(name string) string {
	return r.PostParamStringDefault(name, "")
}

// PostParamInt returns post parameter as int
func (r *HTTP) PostParamInt(name string) int {
	return r.PostParamIntDefault(name, 0)
}

// PostParamBool returns post parameter as bool
func (r *HTTP) PostParamBool(name string) bool {
	return r.PostParamBoolDefault(name, false)
}

// PostParamStringDefault returns post parameter as string or d
func (r *HTTP) PostParamStringDefault(name string, d string) string {
	vi := r.PostForm.Get(name)
	switch v := vi.(type) {
	case []string:
		if len(v) > 0 {
			return v[0]
		}
		break

	case string:
		return v
	}

	return d
}

// PostParamIntDefault returns post parameter as int or d
func (r *HTTP) PostParamIntDefault(name string, d int) int {
	vi := r.PostForm.Get(name)
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

	return d
}

// PostParamBoolDefault returns post parameter as bool or d
func (r *HTTP) PostParamBoolDefault(name string, d bool) bool {
	vi := r.PostForm.Get(name)
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

	return d
}

// PostParamFloatDefault returns post parameter as int or d
func (r *HTTP) PostParamFloatDefault(name string, d float64) float64 {
	vi := r.PostForm.Get(name)
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

	return d
}

// PostParamMapDefault returns post parameter as int or d
func (r *HTTP) PostParamMapDefault(name string, d map[string]interface{}) map[string]interface{} {
	switch v := r.PostForm.Get(name).(type) {
	case []map[string]interface{}:
		if len(v) > 0 {
			return v[0]
		}

	case map[string]interface{}:
		return v

	case utils.DataTree:
		return utils.MapFromDataTree(v)
	}

	return d
}

// PostParams returns post parameters
func (r *HTTP) PostParams() map[string]interface{} {
	p := make(map[string]interface{})
	for k, v := range r.PostForm {
		p[k] = v
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

func (r *HTTP) SetFileTransfer(t file.TransferInterface) {
	r.Uploads = t
}

func (r HTTP) FileTransfer() file.TransferInterface {
	return r.Uploads
}

// contentType returns the payload content type
func (r *HTTP) contentType() int {
	ct, _, err := mime.ParseMediaType(r.Header("Content-Type"))
	if err != nil {
		return contentText
	}

	if ct == "" || strings.Contains(ct, "application/octet-stream") {
		return contentStream
	} else if strings.Contains(ct, "application/x-www-form-urlencoded") {
		return contentFormData
	} else if strings.Contains(ct, "multipart/form-data") {
		return contentMultipart
	} else if strings.Contains(ct, "application/json") {
		return contentJSON
	}

	return contentText
}

func (r *HTTP) parseForm(rqs *http.Request) error {
	var err error
	if r.PostForm == nil {
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			r.PostForm, err = r.parsePostForm(rqs)
		}

		if r.PostForm == nil {
			r.PostForm = make(utils.DataTree)
		}
	}

	if r.PostForm == nil {
		if len(r.PostForm) > 0 {
			r.Form = make(utils.DataTree)
			copyValues(r.Form, r.PostForm)
		}

		newValues := make(utils.DataTree)
		if rqs.URL != nil {
			er := parseQuery(newValues, rqs.URL.RawQuery)
			if err == nil {
				err = er
			}
		}

		if r.Form == nil {
			r.Form = newValues
		} else {
			copyValues(r.Form, newValues)
		}
	}

	return err
}

func (r *HTTP) parsePostForm(rqs *http.Request) (pf utils.DataTree, err error) {
	pf = make(utils.DataTree)
	if rqs.Body == nil {
		err = errors.New("Missing form body")
		return
	}

	rqs.Body = struct {
		io.Reader
		io.Closer
	}{
		Reader: io.LimitReader(rqs.Body, r.MaxRequestSize),
		Closer: rqs.Body,
	}

	switch r.contentType() {
	case contentFormData:
		b, er := io.ReadAll(rqs.Body)
		if er != nil {
			if err == nil {
				err = er
			}
			break
		}

		if int64(len(b)) > r.MaxRequestSize {
			err = errors.New("Request is too large")
			return
		}

		er = parseQuery(pf, string(b))
		if err == nil {
			err = er
		}

	case contentMultipart:
		reader, er := rqs.MultipartReader()
		if er != nil {
			if err == nil {
				err = errors.Wrap(er, "Unable to parse request body")
				return
			}
			break
		}

		c := int64(0)
	ReadBodyLoop:
		for {
			c++
			if c > r.MaxFormSize {
				log.Instance().Warning(errors.New("Form has too many values"), nil)
				break ReadBodyLoop
			}

			p, er := reader.NextPart()
			if er == io.EOF {
				break ReadBodyLoop
			} else if er == io.ErrUnexpectedEOF {
				err = errors.New("Request is too large")
				return
			} else if er != nil {
				log.Instance().Warning(errors.Wrap(er, "Error parsing multipart form data"), nil)
				continue
			}

			if p.FileName() != "" {
				f := file.NewUpload(p.FileName(), p)
				if er = r.Uploads.UploadFile(f); er != nil {
					log.Instance().Warning(errors.Wrap(er, "Error parsing multipart form data"), nil)
					continue
				}

				pf.Push(p.FormName(), f)

				if er == io.EOF {
					break ReadBodyLoop
				}
			} else {
				value, er := io.ReadAll(p)
				if er == io.EOF {
					break ReadBodyLoop
				} else if er == io.ErrUnexpectedEOF {
					err = errors.New("Request is too large")
					return
				} else if er != nil {
					log.Instance().Warning(errors.Wrap(er, "Error parsing multipart form data"), nil)
					continue
				}

				pf.Push(p.FormName(), string(value))
			}
		}

	case contentStream:
		b, er := io.ReadAll(rqs.Body)
		if er != nil {
			if err == nil {
				err = er
			}
			break
		}

		if int64(len(b)) > r.MaxRequestSize {
			err = errors.New("Request is too large")
			return
		}

		er = parseQuery(pf, string(b))
		if err == nil {
			err = er
		}

	case contentJSON:
		b, er := io.ReadAll(rqs.Body)
		if er != nil {
			if err == nil {
				err = er
			}
			break
		}

		if int64(len(b)) > r.MaxRequestSize {
			err = errors.New("Request is too large")
			return
		}

		m := make(map[string]interface{})
		if er := json.Unmarshal(b, &m); er != nil {
			if err == nil {
				err = errors.Wrap(er, "Unable to parse request")
				return
			}
		}

		for k, v := range m {
			pf.Push(k, v)
		}
		break
	}

	return
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
func NewHTTPRequest(r *http.Request, ft file.TransferInterface, proxyed bool, maxsize int64, maxform int64) (ri Interface, err error) {
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
		Uploads:    ft,
	}

	req.Cks = make(map[string]*http.Cookie)
	req.MdlKey = "module"
	req.CtrlKey = "controller"
	req.ActKey = "action"
	req.Proxyed = proxyed
	req.Prms = make(map[string]interface{})
	req.RemoteAddr = req.clientIP()
	req.Secure = r.URL.Scheme == "https"
	req.MaxRequestSize = maxsize
	req.MaxFormSize = maxform

	if r.URL.RawPath != "" {
		req.Path = r.URL.RawPath
	} else {
		req.Path = r.URL.Path
	}

	for k, v := range r.URL.Query() {
		req.Prms[k] = v[0]
	}

	return req, nil
}

func copyValues(dst, src utils.DataTree) {
	for k, v := range src {
		dst.Push(k, v)
	}
}

func parseQuery(m utils.DataTree, query string) (err error) {
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = errors.Errorf("Invalid semicolon separator in query")
			continue
		}

		if key == "" {
			continue
		}

		key, value, _ := strings.Cut(key, "=")
		key, err1 := url.QueryUnescape(key)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}
		value, err1 = url.QueryUnescape(value)
		if err1 != nil {
			if err == nil {
				err = err1
			}
			continue
		}

		m.Push(key, value)
	}

	return err
}
