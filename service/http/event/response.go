package event

import (
	"strconv"
	"time"
	"wsf/controller/request"
	"wsf/controller/response"
)

// Response represents singular http response event
type Response struct {
	Request  request.Interface
	Response response.Interface
	err      error
	start    time.Time
	elapsed  time.Duration
}

// Error service.Event interface implementation
func (e *Response) Error() error {
	return e.err
}

// Message service.Event interface implementation
func (e *Response) Message() string {
	return e.Request.(*request.HTTP).RemoteAddr + ` - - [` + e.start.Format(time.RFC3339) + `] "` + e.Request.(*request.HTTP).Method + ` ` + e.Request.(*request.HTTP).RequestURI + ` ` + e.Request.(*request.HTTP).Protocol + `" ` + strconv.Itoa(e.Response.ResponseCode()) + ` ` + strconv.Itoa(int(e.Response.ContentLength())) + ` "` + e.Request.(*request.HTTP).Referer + `" "` + e.Request.(*request.HTTP).UserAgent + `"`
}

// Elapsed returns duration of the invocation
func (e *Response) Elapsed() time.Duration {
	return e.elapsed
}

// NewResponse creates new response event
func NewResponse(rqs request.Interface, rsp response.Interface, err error, start time.Time) *Response {
	return &Response{Request: rqs, Response: rsp, err: err, start: start, elapsed: time.Since(start)}
}
