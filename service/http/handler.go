package http

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/service"
	"github.com/noxyicm/wsf/service/http/event"
	"github.com/noxyicm/wsf/session"
	"github.com/noxyicm/wsf/utils"
)

// Handler serves http connections
type Handler struct {
	options *Config
	mdwr    []Middleware
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

// AddMiddleware adds new net/http middleware
func (h *Handler) AddMiddleware(m Middleware) {
	h.mdwr = append(h.mdwr, m)
}

// SetMiddlewares adds new net/http middlewares
func (h *Handler) SetMiddlewares(mdwrs []Middleware) {
	h.mdwr = mdwrs
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
func (h *Handler) ServeHTTP(r request.Interface, w response.Interface) {
	h.throw(EventDebug, service.InfoEvent(fmt.Sprintf("Serving HTTP request: %s", r.PathInfo())))
	start := time.Now()
	defer h.recover(r, w, start)

	if err := r.ParseBody(); err != nil {
		h.handleError(r, w, err, start)
		return
	}

	if h.options.MaxRequestSize != 0 {
		if length := r.Header("content-length"); length != "" {
			if size, err := strconv.ParseInt(length, 10, 64); err != nil {
				h.handleError(r, w, err, start)
				return
			} else if size > h.options.MaxRequestSize {
				h.handleError(r, w, errors.New("Request body max size is exceeded"), start)
				return
			}
		}
	}

	var s session.Interface
	var sid string
	var err error
	if session.Created() {
		s, sid, err = session.Start(r, w)
		if err != nil {
			h.handleError(r, w, err, start)
			return
		}
	}

	//ctx, err := context.NewContext(context.Background())
	ctx, err := context.NewContext(r.Context())
	if err != nil {
		if session.Created() {
			session.Close(sid)
		}

		h.handleError(r, w, err, start)
		return
	}

	ctx.SetRequest(r)
	ctx.SetResponse(w)
	if session.Created() {
		ctx.SetValue(context.SessionIDKey, sid)
		ctx.SetValue(context.SessionKey, s)
	}

	if err := h.ctrl.Dispatch(ctx, r, w); err != nil {
		if session.Created() {
			session.Close(sid)
		}

		h.handleResponse(r, w, err, start)
		return
	}

	if session.Created() {
		session.Close(sid)
	}

	h.handleResponse(r, w, nil, start)
}

// handleError sends error response to client
func (h *Handler) handleError(r request.Interface, w response.Interface, err error, start time.Time) {
	for hdr, val := range h.options.Headers {
		w.SetHeader(hdr, val)
	}

	w.SetResponseCode(500)

	h.throw(EventHTTPError, event.NewError(r.GetRequest(), err, start))
	w.SetBody([]byte(err.Error()))
	w.Write()
}

// handleResponse triggers response event
func (h *Handler) handleResponse(r request.Interface, w response.Interface, err error, start time.Time) {
	for hdr, val := range h.options.Headers {
		w.SetHeader(hdr, val)
	}

	if err != nil {
		switch err.(type) {
		case *errors.HTTPError:
			w.SetResponseCode(err.(*errors.HTTPError).Code())

		default:
			w.SetResponseCode(500)
			w.SetBody([]byte(err.Error()))
		}
	} else if w.ResponseCode() == 0 {
		w.SetResponseCode(200)
	}

	h.throw(EventHTTPResponse, event.NewResponse(r, w, err, start))
	w.Write()
}

func (h *Handler) recover(r request.Interface, w response.Interface, start time.Time) {
	if rec := recover(); rec != nil {
		switch err := rec.(type) {
		case error:
			utils.DebugBacktrace()
			h.handleError(r, w, errors.Wrap(err, "[HTTP Server] Unxpected error equired"), start)
			break

		default:
			utils.DebugBacktrace()
			h.handleError(r, w, errors.Errorf("[HTTP Server] Unxpected error equired: %v", err), start)
		}
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
