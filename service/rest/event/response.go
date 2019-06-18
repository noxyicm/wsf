package event

import (
	"time"
	"wsf/controller/request"
	"wsf/controller/response"
)

// Response represents singular http response event
type Response struct {
	Request  request.Interface
	Response response.Interface
	Error    error
	start    time.Time
	elapsed  time.Duration
}

// Elapsed returns duration of the invocation
func (e *Response) Elapsed() time.Duration {
	return e.elapsed
}

// NewResponse creates new response event
func NewResponse(rqs request.Interface, rsp response.Interface, err error, start time.Time) *Response {
	return &Response{Request: rqs, Response: rsp, Error: err, start: start, elapsed: time.Since(start)}
}
