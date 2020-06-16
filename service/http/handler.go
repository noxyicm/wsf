package http

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"wsf/context"
	"wsf/controller"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/registry"
	"wsf/service"
	"wsf/service/http/event"
)

// Handler serves http connections
type Handler struct {
	options *Config
	ctrl    controller.Interface
	lsns    []func(event int, ctx service.Event)
	mu      sync.RWMutex
}

// AddListener attaches handler event watcher
func (h *Handler) AddListener(l func(event int, ctx service.Event)) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lsns = append(h.lsns, l)
}

// throw invokes event handler if any
func (h *Handler) throw(event int, ctx service.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, l := range h.lsns {
		l(event, ctx)
	}
}

// ServeHTTP Serves a HTTP request
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recover(w, r, start)

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

	//s, sid, err := session.Start(req, rsp)
	//if err != nil {
	//	h.handleError(w, r, err, start)
	//	return
	//}

	ctx, err := context.NewContext(context.Background())
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}
	ctx.SetRequest(req)
	ctx.SetResponse(rsp)
	//ctx.SetValue(context.SessionIDKey, sid)
	//ctx.SetValue(context.SessionKey, s)
	if err := h.ctrl.Dispatch(ctx, req, rsp); err != nil {
		//session.Close(sid)
		h.handleResponse(req, rsp, err, start)
		return
	}

	//session.Close(sid)
	h.handleResponse(req, rsp, nil, start)
}

// handleError sends error response to client
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error, start time.Time) {
	for hdr, val := range h.options.Headers {
		w.Header().Add(hdr, val)
	}

	w.WriteHeader(500)

	fmt.Printf("%+v\n", err)
	h.throw(EventHTTPError, event.NewError(r, err, start))
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
			rsp.SetBody([]byte(err.Error()))
		}
	} else if rsp.ResponseCode() == 0 {
		rsp.SetResponseCode(200)
	}

	h.throw(EventHTTPResponse, event.NewResponse(req, rsp, err, start))
	rsp.Write()
}

func (h *Handler) recover(w http.ResponseWriter, r *http.Request, start time.Time) {
	if rec := recover(); rec != nil {
		h.handleError(w, r, errors.Wrap(rec.(error), "[HTTP Server] Unxpected error equired"), start)
	}
}

// NewHandler creates a new handler
func NewHandler(cfg *Config) (h *Handler, err error) {
	rsr := registry.GetResource("maincontroller")
	if rsr == nil {
		return nil, errors.New("[HTTP Server] Maincontroller resource must be registered and initialized")
	}

	return &Handler{options: cfg, ctrl: rsr.(controller.Interface)}, nil
}
