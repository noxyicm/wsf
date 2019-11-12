package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"wsf/application/modules"
	"wsf/config"
	"wsf/context"
	"wsf/controller"
	"wsf/controller/request"
	"wsf/controller/request/attributes"
	"wsf/errors"
	"wsf/filter/word"
	"wsf/log"
	"wsf/registry"
	"wsf/service"
	"wsf/service/environment"
	evt "wsf/service/http/event"

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

	// EventResponse thrown after the request has been processed
	EventResponse

	// ID of service
	ID = "rest"
)

// http middleware
type middleware func(f http.HandlerFunc) http.HandlerFunc

// Service manages http servers
type Service struct {
	options      *Config
	accessLogger *log.Log
	logger       *log.Log
	env          environment.Interface
	prefix       string
	mdwr         []middleware
	serviceMap   sync.Map
	lsns         []func(event int, ctx interface{})
	mu           sync.RWMutex
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

// AddListener attaches event watcher
func (s *Service) AddListener(l func(event int, ctx interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lsns = append(s.lsns, l)
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx interface{}) {
	for _, l := range s.lsns {
		l(event, ctx)
	}
}

// Init HTTP service
func (s *Service) Init(options *Config, env environment.Interface) (bool, error) {
	if !options.Enable {
		return false, nil
	}

	s.options = options
	s.env = env
	s.signalChan = make(chan os.Signal)
	s.prefix = options.Prefix

	acclogger, err := log.NewLog(options.AccessLogger)
	if err != nil {
		return false, err
	}
	s.accessLogger = acclogger
	s.AddListener(s.logAccess)

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return false, errors.New("[REST Server] Log resource is not configured")
	}
	s.logger = logResource.(*log.Log)

	mdls := registry.GetResource("modules")
	if mdls == nil {
		return false, errors.New("[REST Server] Resource 'modules' is required but not initialized")
	}

	var mdl *modules.Module
	var mdlPart string
	var ctrlPart string

	for _, svc := range options.Services {
		parts := strings.Split(svc, ".")
		if len(parts) == 0 {
			return false, errors.Errorf("[REST Server] Service '%s' is not a valid REST service", svc)
		} else if len(parts) == 1 {
			return false, errors.Errorf("[REST Server] Service '%s' is not a valid REST service", svc)
		} else {
			mdlPart = parts[0]
			ctrlPart = parts[1]
			mdl = mdls.(modules.Handler).Module(mdlPart)
		}

		if mdl == nil {
			return false, errors.Errorf("[REST Server] Module '%s' is not registered", mdlPart)
		}

		ctrlType, err := mdl.ControllerType(ctrlPart)
		if err != nil {
			return false, errors.Wrap(err, "[REST Server] Error")
		}

		ctrl := reflect.New(ctrlType).Interface()
		if _, ok := ctrl.(Controller); !ok {
			return false, errors.Errorf("[REST Server] Controller '%s' does not implements rest.Controller interface", ctrlPart)
		}

		if _, dup := s.serviceMap.LoadOrStore(svc, ctrlType); dup {
			return false, errors.New("[REST Server] Rest receiver already defined: " + svc)
		}
	}

	rtr, err := NewRestRoute("", map[string]string{
		"module":     "index",
		"controller": "front",
		"action":     "index",
	}, nil)
	if err != nil {
		return false, errors.Wrap(err, "[REST Server] Init error")
	}
	rtr.(*Route).SetService(s)

	uts, err := word.NewSeparatorToSeparator("/", "-")
	if err != nil {
		return false, errors.Wrap(err, "[REST Server] Init error")
	}

	routeName, err := uts.Filter(s.prefix)
	if err != nil {
		return false, errors.Wrap(err, "[REST Server] Init error")
	}

	if err := controller.Instance().Router().AddRoute(rtr, strings.Trim(routeName.(string), "-")); err != nil {
		return false, errors.Wrap(err, "[REST Server] Init error")
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

	s.handler, err = NewHandler(s.options)
	if err != nil {
		return err
	}
	s.handler.AddListener(s.throw)

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

	errChan := make(chan error, 1)
	s.throw(EventInfo, fmt.Sprintf("[REST Server] Starting: Listening on %s...", s.options.Address()))
	go func() { errChan <- s.http.ListenAndServe() }()
	if s.https != nil {
		go func() { errChan <- s.https.ListenAndServeTLS(s.options.SSL.Cert, s.options.SSL.Key) }()
	}

	err = <-errChan
	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()

	if err == http.ErrServerClosed {
		s.throw(EventInfo, "[REST Server] Stoped")
		return nil
	}
	return err
}

// Stop the service
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.throw(EventInfo, "[REST Server] Initiating stop...")
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

// Register publishes in the server the set of methods of the
// receiver value that satisfy the rest.Controller interface
func (s *Service) Register(name string, ctrl Controller) error {
	rs := new(restService)
	rs.typ = reflect.TypeOf(ctrl)
	rs.ctrl = reflect.ValueOf(ctrl)
	rs.name = name

	if _, dup := s.serviceMap.LoadOrStore(name, rs); dup {
		return errors.New("[REST Server] Service already defined: " + name)
	}

	return nil
}

// IsRestfull returns true if module is registered
func (s *Service) IsRestfull(name string) bool {
	if _, ok := s.serviceMap.Load(name); ok {
		return true
	}

	return false
}

// RoutePrefix returns route prefix
func (s *Service) RoutePrefix() string {
	return s.prefix
}

// Init https server
func (s *Service) initSSL() *http.Server {
	server := &http.Server{Addr: s.tlsAddr(s.options.Address(), true), Handler: s}
	s.throw(EventInitSSL, server)

	// Enable HTTP/2 support by default
	http2.ConfigureServer(server, &http2.Server{})

	return server
}

func (s *Service) logAccess(event int, ctx interface{}) {
	switch event {
	case EventResponse:
		s.accessLogger.Info("[REST Server] Logging access", map[string]string{
			"client":     ctx.(*evt.Response).Request.(*request.HTTP).RemoteAddr,
			"user":       "-",
			"request":    ctx.(*evt.Response).Request.(*request.HTTP).Method + " " + ctx.(*evt.Response).Request.(*request.HTTP).RequestURI + " " + ctx.(*evt.Response).Request.(*request.HTTP).Protocol,
			"statusCode": strconv.Itoa(ctx.(*evt.Response).Response.ResponseCode()),
			"bytes":      strconv.Itoa(int(ctx.(*evt.Response).Response.ContentLength())),
			"referer":    ctx.(*evt.Response).Request.(*request.HTTP).Referer,
			"useragent":  ctx.(*evt.Response).Request.(*request.HTTP).UserAgent,
		})

	case EventError:
		s.accessLogger.Info("[REST Server] Logging access", map[string]string{
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
	return &Service{serving: false, priority: 6}, nil
}

type restService struct {
	name string
	ctrl reflect.Value
	typ  reflect.Type
}

/*func (server *Server) register(rcvr interface{}, name string, useName bool) error {
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		s := "rpc.Register: no service name for type " + s.typ.String()
		log.Print(s)
		return errors.New(s)
	}
	if !isExported(sname) && !useName {
		s := "rpc.Register: type " + sname + " is not exported"
		log.Print(s)
		return errors.New(s)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		log.Print(str)
		return errors.New(str)
	}

	if _, dup := server.serviceMap.LoadOrStore(sname, s); dup {
		return errors.New("rpc: service already defined: " + sname)
	}
	return nil
}

func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name

		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			if reportErr {
				log.Printf("[REST] Register: method %q has %d input parameters; needs exactly three\n", mname, mtype.NumIn())
			}
			continue
		}

		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				log.Printf("[REST] Register: argument type of method %q is not exported: %q\n", mname, argType)
			}
			continue
		}

		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				log.Printf("[REST] Register: reply type of method %q is not a pointer: %q\n", mname, replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				log.Printf("rpc.Register: reply type of method %q is not exported: %q\n", mname, replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				log.Printf("rpc.Register: method %q has %d output parameters; needs exactly one\n", mname, mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				log.Printf("rpc.Register: return type of method %q is %q, must be error\n", mname, returnType)
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}
*/
