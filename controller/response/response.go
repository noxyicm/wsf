package response

import "net/http"

// Interface is an interface for responses
type Interface interface {
	SetHeader(key string, value string) error
	AddHeader(key string, value string) error
	ClearHeaders() error
	RemoveHeader(key string) error
	SetResponseCode(code int) error
	ResponseCode() int
	SetBody([]byte) error
	AppendBody(data []byte, segment string) error
	ContentLength() int
	AddCookies(cookies []*http.Cookie)
	AddCookie(value *http.Cookie)
	AddStringCookies(cookies map[string]string)
	AddStringCookie(key string, value string)
	Cookie(key string) string
	SetRedirect(url string, code int) error
	IsRedirect() bool
	Write() error
	Destroy()
	SetException(err error)
	Exceptions() []error
	IsException() bool
	RequestSend(flag bool)
	IsSendRequested() bool
	SetData(key string, value interface{})
	GetData(key string) interface{}
	Data() map[string]interface{}
}
