package http

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"wsf/config"
	"wsf/controller/request"
	"wsf/controller/request/attributes"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
	"wsf/service"
	"wsf/service/environment"
	evt "wsf/service/http/event"

	"golang.org/x/net/http2"
)

const (
	// ID of service
	ID = "http"

	// EventInitSSL describes TLS initialization
	EventInitSSL int = iota
)

// http middleware
type middleware func(f http.HandlerFunc) http.HandlerFunc

// Service manages http servers
type Service struct {
	options      *Config
	accessLogger *log.Log
	logger       *log.Log
	env          environment.Interface
	lsns         []func(event int, ctx interface{})
	mdwr         []middleware
	mu           sync.Mutex
	serving      bool
	handler      *Handler
	http         *http.Server
	https        *http.Server
	signalChan   chan os.Signal
	priority     int
}

// AddMiddleware adds new net/http middleware
func (s *Service) AddMiddleware(m middleware) {
	s.mdwr = append(s.mdwr, m)
}

// AddListener attaches server event watcher
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.lsns = append(s.lsns, l)
}

// Init HTTP service
func (s *Service) Init(options *Config, env environment.Interface) (bool, error) {
	if !options.Enable {
		return false, nil
	}

	s.options = options
	s.env = env
	s.signalChan = make(chan os.Signal)
	s.AddListener(s.logAccess)

	acclogger, err := log.NewLog(options.AccessLogger)
	if err != nil {
		return false, err
	}
	s.accessLogger = acclogger

	logResource := registry.Get("log")
	if logResource == nil {
		return false, errors.New("Log resource is not configured")
	}
	s.logger = logResource.(*log.Log)

	return true, nil
}

// Priority returns predefined service priority
func (s *Service) Priority() int {
	return s.priority
}

// Serve the service
func (s *Service) Serve() (err error) {
	s.mu.Lock()

	s.handler, err = NewHandler(s.options)
	if err != nil {
		return err
	}
	s.handler.Listen(s.throw)

	s.http = &http.Server{
		Addr:         s.options.Address(),
		Handler:      s,
		ReadTimeout:  time.Duration(s.options.MaxRequestTimeout) * time.Second,
		WriteTimeout: time.Duration(s.options.MaxResponseTimeout) * time.Second,
	}
	if s.options.EnableTLS() {
		s.https = s.initSSL()
	}
	s.serving = true
	s.mu.Unlock()

	errChan := make(chan error, 2)
	s.logger.Infof("[HTTP Server] Starting: Listening on %s...", nil, s.options.Address())
	go func() { errChan <- s.http.ListenAndServe() }()
	if s.https != nil {
		go func() { errChan <- s.https.ListenAndServeTLS(s.options.SSL.Cert, s.options.SSL.Key) }()
	}

	err = <-errChan
	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()

	if err == http.ErrServerClosed {
		s.logger.Info("[HTTP Server] Stoped", nil)
		return nil
	}
	return err
}

// Stop the service
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("[HTTP Server] Stopping", nil)
	if s.http == nil {
		return
	}

	if s.https != nil {
		go s.https.Shutdown(context.Background())
	}

	go s.http.Shutdown(context.Background())
}

// ServeHTTP handles connection using set of middleware and rr PSR-7 server.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.https != nil && r.TLS == nil && s.options.SSL.Redirect {
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

	// chaining middleware
	f := s.handler.ServeHTTP
	for _, m := range s.mdwr {
		f = m(f)
	}
	f(w, r)
}

// Init https server.
func (s *Service) initSSL() *http.Server {
	server := &http.Server{Addr: s.tlsAddr(s.options.Address(), true), Handler: s}
	s.throw(EventInitSSL, server)

	// Enable HTTP/2 support by default
	http2.ConfigureServer(server, &http2.Server{})

	return server
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.lsns {
		l(event, ctx)
	}
}

func (s *Service) logAccess(event int, ctx interface{}) {
	switch event {
	case EventResponse:
		s.accessLogger.Info("Logging access", map[string]string{
			"client":     ctx.(*evt.Response).Request.(*request.HTTP).RemoteAddr,
			"user":       "-",
			"request":    ctx.(*evt.Response).Request.(*request.HTTP).Method + " " + ctx.(*evt.Response).Request.(*request.HTTP).RequestURI + " " + ctx.(*evt.Response).Request.(*request.HTTP).Protocol,
			"statusCode": strconv.Itoa(ctx.(*evt.Response).Response.ResponseCode()),
			"bytes":      strconv.Itoa(int(ctx.(*evt.Response).Response.ContentLength())),
			"referer":    ctx.(*evt.Response).Request.(*request.HTTP).Referer,
			"useragent":  ctx.(*evt.Response).Request.(*request.HTTP).UserAgent,
		})

	case EventError:
		s.accessLogger.Info("Logging access", map[string]string{
			"client":     ctx.(*evt.Error).Request.RemoteAddr,
			"user":       "-",
			"request":    ctx.(*evt.Error).Request.Method + " " + ctx.(*evt.Error).Request.URL.RequestURI() + " " + ctx.(*evt.Error).Request.Proto,
			"statusCode": "500",
			"bytes":      strconv.Itoa(len([]byte(ctx.(*evt.Error).Error.Error()))),
			"referer":    ctx.(*evt.Error).Request.Referer(),
			"useragent":  ctx.(*evt.Error).Request.UserAgent(),
		})
	}
}

// tlsAddr replaces listen or host port with port configured by SSL config.
func (s *Service) tlsAddr(host string, forcePort bool) string {
	// remove current forcePort first
	host = strings.Split(host, ":")[0]

	if forcePort || s.options.SSL.Port != 443 {
		host = fmt.Sprintf("%s:%v", host, s.options.SSL.Port)
	}

	return host
}

// NewService creates a new service of type HTTP
func NewService(cfg config.Config) (service.Interface, error) {
	return &Service{serving: false, priority: 0}, nil
}
