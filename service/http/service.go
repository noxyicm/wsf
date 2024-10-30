package http

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller/request"
	"github.com/noxyicm/wsf/controller/request/attributes"
	"github.com/noxyicm/wsf/controller/response"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/service"
	"github.com/noxyicm/wsf/service/environment"
	evt "github.com/noxyicm/wsf/service/http/event"
	"github.com/noxyicm/wsf/utils"

	"golang.org/x/net/http2"
)

const (
	// EventDebug thrown if there is something insegnificant to say
	EventDebug = iota + 500

	// EventInfo thrown if there is something to say
	EventInfo

	// EventError thrown on any non job error provided
	EventError

	// EventInitSSL describes TLS initialization
	EventInitSSL

	// EventHTTPResponse thrown after the http request has been processed
	EventHTTPResponse

	// EventHTTPError thrown after the http request has been processed with error
	EventHTTPError

	// ID of service
	ID = "http"
)

// Service manages http servers
type Service struct {
	Name         string
	Options      *Config
	AccessLogger *log.Log
	Logger       *log.Log
	env          environment.Interface
	mdwr         []Middleware
	lsns         []func(event int, ctx service.Event)
	serving      bool
	handler      *Handler
	http         *http.Server
	https        *http.Server
	signalChan   chan os.Signal
	externalChan chan interface{}
	priority     int
	mu           sync.RWMutex
}

// AddMiddleware adds new net/http middleware
func (s *Service) AddMiddleware(m Middleware) {
	s.mdwr = append(s.mdwr, m)

	if s.handler != nil {
		s.handler.AddMiddleware(m)
	}
}

// AddListener attaches event watcher
func (s *Service) AddListener(l func(event int, ctx service.Event)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lsns = append(s.lsns, l)
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx service.Event) {
	s.mu.Lock()
	lsns := s.lsns
	s.mu.Unlock()

	for _, l := range lsns {
		l(event, ctx)
	}
}

// Init HTTP service
func (s *Service) Init(options *Config, env environment.Interface) (bool, error) {
	if !options.Enable {
		return false, nil
	}

	s.Options = options
	s.env = env
	s.signalChan = make(chan os.Signal)

	acclogger, err := log.NewLog(options.AccessLogger)
	if err != nil {
		return false, err
	}
	s.AccessLogger = acclogger
	s.AddListener(s.logAccess)

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return false, errors.Errorf("[%s] Log resource is not configured", s.Name)
	}
	s.Logger = logResource.(*log.Log)

	for _, mc := range s.Options.Middleware {
		if !mc.Enable {
			continue
		}

		mdlw, err := NewMiddleware(mc.Type, mc)
		if err != nil {
			return false, errors.Wrapf(err, "[%s] Unable to add middleware", s.Name)
		}

		s.AddMiddleware(mdlw)
	}

	return true, nil
}

// Priority returns predefined service priority
func (s *Service) Priority() int {
	return s.priority
}

// Serve the service
func (s *Service) Serve(ctx context.Context) (err error) {
	s.mu.Lock()

	s.handler, err = NewHandler(s.Options)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	s.handler.AddListener(s.throw)
	s.handler.SetMiddlewares(s.mdwr)

	s.http = &http.Server{
		Addr:         s.Options.Address(),
		Handler:      s,
		ReadTimeout:  time.Duration(s.Options.MaxRequestTimeout) * time.Second,
		WriteTimeout: time.Duration(s.Options.MaxResponseTimeout) * time.Second,
	}
	if s.Options.EnableTLS() {
		s.https = s.initSSL()
	}
	s.serving = true
	s.mu.Unlock()

	errChan := make(chan error)
	s.throw(EventInfo, service.InfoEvent(fmt.Sprintf("[%s] Starting: Listening on %s...", s.Name, s.Options.Address())))
	go func() { errChan <- s.http.ListenAndServe() }()
	if s.https != nil {
		go func() { errChan <- s.https.ListenAndServeTLS(s.Options.SSL.Cert, s.Options.SSL.Key) }()
	}

	err = <-errChan
	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()

	if err == http.ErrServerClosed {
		s.throw(EventInfo, service.InfoEvent(fmt.Sprintf("[%s] Stoped", s.Name)))
		return nil
	}

	return err
}

// Stop the service
func (s *Service) Stop() {
	s.mu.Lock()
	serviceHTTP := s.http
	serviceHTTPS := s.https
	s.mu.Unlock()

	s.throw(EventInfo, service.InfoEvent(fmt.Sprintf("[%s] Initiating stop...", s.Name)))
	if serviceHTTP == nil {
		return
	}

	if serviceHTTPS != nil {
		go serviceHTTPS.Shutdown(context.Background())
	}

	go serviceHTTP.Shutdown(context.Background())
}

// ServeHTTP handles connection using set of middleware and rr PSR-7 server.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.throw(EventDebug, service.InfoEvent(fmt.Sprintf("Serving HTTP request: %s", r.RequestURI)))
	start := time.Now()
	defer s.recover(w, r, start)

	if s.https != nil && r.TLS == nil && s.Options.SSL.Redirect {
		target := &url.URL{
			Scheme:   "https",
			Host:     s.tlsAddr(r.Host, false),
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}

		http.Redirect(w, r, target.String(), http.StatusTemporaryRedirect)
		return
	}

	r = attributes.Init(r)

	req, err := request.NewHTTPRequest(r, s.Options.Uploads, s.Options.Proxy)
	if err != nil {
		s.handleError(w, r, err, start)
		return
	}

	rsp, err := response.NewHTTPResponse(w)
	if err != nil {
		s.handleError(w, r, err, start)
		return
	}

	// chaining middleware
	for _, m := range s.mdwr {
		if m.Handle(s, req, rsp) {
			return
		}
	}

	s.handler.ServeHTTP(req, rsp)
}

// Init https server.
func (s *Service) initSSL() *http.Server {
	server := &http.Server{Addr: s.tlsAddr(s.Options.Address(), true), Handler: s}
	s.throw(EventInitSSL, service.DebugEvent(fmt.Sprintf("[%s] Initiating SSL", s.Name)))

	// Enable HTTP/2 support by default
	http2.ConfigureServer(server, &http2.Server{})

	return server
}

func (s *Service) logAccess(event int, ctx service.Event) {
	switch event {
	case EventHTTPResponse:
		s.AccessLogger.Info("Logging access", map[string]string{
			"client":     ctx.(*evt.Response).Request.(*request.HTTP).RemoteAddr,
			"user":       "-",
			"request":    ctx.(*evt.Response).Request.(*request.HTTP).Method + " " + ctx.(*evt.Response).Request.(*request.HTTP).RequestURI + " " + ctx.(*evt.Response).Request.(*request.HTTP).Protocol,
			"statusCode": strconv.Itoa(ctx.(*evt.Response).Response.ResponseCode()),
			"bytes":      strconv.Itoa(int(ctx.(*evt.Response).Response.ContentLength())),
			"referer":    ctx.(*evt.Response).Request.(*request.HTTP).Referer,
			"useragent":  ctx.(*evt.Response).Request.(*request.HTTP).UserAgent,
		})

	case EventHTTPError:
		s.AccessLogger.Info("Logging access", map[string]string{
			"client":     ctx.(*evt.Error).Request.RemoteAddr,
			"user":       "-",
			"request":    ctx.(*evt.Error).Request.Method + " " + ctx.(*evt.Error).Request.URL.RequestURI() + " " + ctx.(*evt.Error).Request.Proto,
			"statusCode": "500",
			"bytes":      strconv.Itoa(len([]byte(ctx.Message()))),
			"referer":    ctx.(*evt.Error).Request.Referer(),
			"useragent":  ctx.(*evt.Error).Request.UserAgent(),
		})
	}
}

// tlsAddr replaces listen or host port with port configured by SSL config.
func (s *Service) tlsAddr(host string, forcePort bool) string {
	// remove current forcePort first
	host = strings.Split(host, ":")[0]

	if forcePort || s.Options.SSL.Port != 443 {
		host = fmt.Sprintf("%s:%v", host, s.Options.SSL.Port)
	}

	return host
}

// handleError sends error response to client
func (s *Service) handleError(w http.ResponseWriter, r *http.Request, err error, start time.Time) {
	for hdr, val := range s.Options.Headers {
		w.Header().Add(hdr, val)
	}

	w.WriteHeader(500)

	s.throw(EventHTTPError, evt.NewError(r, err, start))
	w.Write([]byte(err.Error()))
}

func (s *Service) recover(w http.ResponseWriter, r *http.Request, start time.Time) {
	if rec := recover(); rec != nil {
		switch err := rec.(type) {
		case error:
			utils.DebugBacktrace()
			s.handleError(w, r, errors.Wrap(err, "[HTTP Server] Unxpected error equired"), start)
			break

		default:
			utils.DebugBacktrace()
			s.handleError(w, r, errors.Errorf("[HTTP Server] Unxpected error equired: %v", err), start)
		}
	}
}

// NewService creates a new service of type HTTP
func NewService(cfg config.Config) (service.Interface, error) {
	return &Service{
		Name:     "HTTP Server",
		serving:  false,
		priority: 0,
	}, nil
}
