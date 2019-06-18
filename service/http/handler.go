package http

import (
	"net/http"
	"strconv"
	"sync"
	"time"
	"wsf/controller"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/registry"
	"wsf/service/http/event"
	"wsf/session"
)

const (
	// EventResponse thrown after the request has been processed
	EventResponse = iota + 500

	// EventError thrown on any non job error provided by server
	EventError
)

// Handler serves http connections
type Handler struct {
	options *Config
	ctrl    controller.Interface
	mul     sync.Mutex
	lsn     func(event int, ctx interface{})
}

// Listen attaches handler event watcher
func (h *Handler) Listen(l func(event int, ctx interface{})) {
	h.mul.Lock()
	defer h.mul.Unlock()

	h.lsn = l
}

// ServeHTTP Serves a HTTP request
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if h.options.MaxRequestSize != 0 {
		if length := r.Header.Get("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				h.handleError(w, r, err, start)
				return
			} else if size > h.options.MaxRequestSize {
				h.handleError(w, r, errors.New("Request body max size is exceeded"), start)
				return
			}
		}
	}

	req, err := request.NewHTTPRequest(r, h.options.Uploads, h.options.Proxy)
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}

	rsp, err := response.NewHTTPResponse(w)
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}

	s, sid, err := session.Start(req, rsp)
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}

	if err := h.ctrl.Dispatch(req, rsp, s); err != nil {
		session.Close(sid)
		h.handleResponse(req, rsp, err, start)
		return
	}

	session.Close(sid)
	h.handleResponse(req, rsp, nil, start)
}

// handleError sends error response to client
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error, start time.Time) {
	for hdr, val := range h.options.Headers {
		w.Header().Add(hdr, val)
	}

	w.WriteHeader(500)

	h.throw(EventError, event.NewError(r, err, start))
	w.Write([]byte(err.Error()))
}

// handleResponse triggers response event
func (h *Handler) handleResponse(req request.Interface, rsp response.Interface, err error, start time.Time) {
	for hdr, val := range h.options.Headers {
		rsp.SetHeader(hdr, val)
	}

	if err != nil {
		switch err.(type) {
		case *errors.HTTPError:
			rsp.SetResponseCode(err.(*errors.HTTPError).Code())

		default:
			rsp.SetResponseCode(500)
		}
	} else {
		rsp.SetResponseCode(200)
	}

	h.throw(EventResponse, event.NewResponse(req, rsp, err, start))
	rsp.Write()
}

// throw invokes event handler if any
func (h *Handler) throw(event int, ctx interface{}) {
	h.mul.Lock()
	defer h.mul.Unlock()

	if h.lsn != nil {
		h.lsn(event, ctx)
	}
}

// NewHandler creates a new handler
func NewHandler(cfg *Config) (h *Handler, err error) {
	rsr := registry.Get("maincontroller")
	if rsr == nil {
		return nil, errors.New("Maincontroller resource must be registered and initialized")
	}

	return &Handler{options: cfg, ctrl: rsr.(controller.Interface)}, nil
}
