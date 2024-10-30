package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.throw(EventDebug, service.InfoEvent(fmt.Sprintf("Serving HTTP request: %s", r.RequestURI)))
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

	s, sid, err := session.Start(req, rsp)
	if err != nil {
		h.handleError(w, r, err, start)
		return
	} else if sid == "" {
		if strings.Contains(req.Header("Accept"), "application/json") {
			rsp.SetResponseCode(401)
			rsp.AddHeader("Content-Type", "application-json")
			rsp.AppendBody([]byte("{\"error\":\"invalid_token\",\"error_description\":\"Invalid token\"}"), "")
		} else {
			rsp.SetResponseCode(401)
			rsp.AppendBody([]byte("Invalid token"), "")
		}

		h.handleResponse(req, rsp, nil, start)
		return
	}

	//ctx, err := context.NewContext(context.Background())
	ctx, err := context.NewContext(r.Context())
	if err != nil {
		session.Close(sid)
		h.handleError(w, r, err, start)
		return
	}

	ctx.SetRequest(req)
	ctx.SetResponse(rsp)
	ctx.SetValue(context.SessionIDKey, sid)
	ctx.SetValue(context.SessionKey, s)
	if err := h.ctrl.Dispatch(ctx, req, rsp); err != nil {
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
		switch err := rec.(type) {
		case error:
			utils.DebugBacktrace()
			h.handleError(w, r, errors.Wrap(err, "[HTTP Server] Unxpected error equired"), start)
			break

		default:
			utils.DebugBacktrace()
			h.handleError(w, r, errors.Errorf("[HTTP Server] Unxpected error equired: %v", err), start)
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
