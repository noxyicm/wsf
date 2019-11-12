package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"wsf/config"
	"wsf/context"
	"wsf/controller"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/registry"
	"wsf/service/http/event"
	"wsf/session"
)

// Handler serves http connections
type Handler struct {
	options *Config
	ctrl    controller.Interface
	lsns    []func(event int, ctx interface{})
	mu      sync.RWMutex
}

// AddListener attaches handler event watcher
func (h *Handler) AddListener(l func(event int, ctx interface{})) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lsns = append(h.lsns, l)
}

// throw invokes event handler if any
func (h *Handler) throw(event int, ctx interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, l := range h.lsns {
		l(event, ctx)
	}
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
				h.handleError(w, r, errors.New("[REST] Request body max size is exceeded"), start)
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

	ctx, err := context.NewContext(r.Context())
	if err != nil {
		h.handleError(w, r, err, start)
		return
	}
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

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(500)
	data := JSONResponse{
		StatusCode: 500,
		Version:    h.options.Version,
		BasePath:   config.App.GetString("application.Domain"),
		Status:     1,
		Message:    "",
		URL:        r.URL.Path,
		Data:       nil,
	}

	switch err.(type) {
	case *errors.HTTPError:
		if config.AppEnv == "production" {
			data.Message = err.Error()
		} else {
			data.Message = fmt.Sprintf("%+v", err.(*errors.HTTPError))
		}

	case *errors.Exception:
		if config.AppEnv == "production" {
			data.Message = err.Error()
		} else {
			data.Message = fmt.Sprintf("%+v", err.(*errors.Exception).Original)
		}

	default:
		data.Message = err.Error()
	}

	encoded, _ := json.Marshal(data)

	h.throw(EventError, event.NewError(r, err, start))
	w.Write(encoded)
}

// handleResponse triggers response event
func (h *Handler) handleResponse(req request.Interface, rsp response.Interface, err error, start time.Time) {
	for hdr, val := range h.options.Headers {
		rsp.SetHeader(hdr, val)
	}

	rsp.SetHeader("Content-Type", "application/json; charset=utf-8")

	data := JSONResponse{
		StatusCode: 200,
		Version:    h.options.Version,
		BasePath:   config.App.GetString("application.Domain"),
		Status:     0,
		Message:    "Success",
		URL:        req.PathInfo(),
		Data:       rsp.Data(),
	}

	if err != nil {
		switch err.(type) {
		case *errors.HTTPError:
			data.StatusCode = err.(*errors.HTTPError).Code()
			if config.AppEnv == "production" {
				data.Message = err.Error()
			} else {
				data.Message = fmt.Sprintf("%+v", err.(*errors.HTTPError))
			}

		case *errors.Exception:
			data.StatusCode = err.(*errors.Exception).Code()
			if config.AppEnv == "production" {
				data.Message = err.Error()
			} else {
				data.Message = fmt.Sprintf("%+v", err.(*errors.Exception).Original)
			}

		default:
			data.StatusCode = 500
			data.Message = err.Error()
		}

		data.Status = 1
	}

	encoded, _ := json.Marshal(data)
	rsp.SetBody(encoded)
	rsp.SetResponseCode(data.StatusCode)

	h.throw(EventResponse, event.NewResponse(req, rsp, err, start))
	rsp.Write()
}

// NewHandler creates a new handler
func NewHandler(cfg *Config) (h *Handler, err error) {
	rsr := registry.GetResource("maincontroller")
	if rsr == nil {
		return nil, errors.New("[REST] Maincontroller resource must be registered and initialized")
	}

	return &Handler{options: cfg, ctrl: rsr.(controller.Interface)}, nil
}

// JSONResponse represents json response
type JSONResponse struct {
	StatusCode int
	Version    string
	BasePath   string
	Status     int
	Message    string
	URL        string
	Data       interface{}
}
