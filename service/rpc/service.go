package rpc

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"syscall"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/service"
	"github.com/noxyicm/wsf/transporter/codec"
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
	ID = "rpc"
)

// Service is RPC service
type Service struct {
	Name     string
	options  *Config
	stop     chan interface{}
	rpc      *rpc.Server
	lsns     []func(event int, ctx service.Event)
	mu       sync.Mutex
	serving  bool
	priority int
}

// Init RPC service
func (s *Service) Init(options *Config) (bool, error) {
	if !options.Enable {
		return false, nil
	}

	s.options = options
	s.rpc = rpc.NewServer()
	return true, nil
}

// AddListener attaches server event watcher
func (s *Service) AddListener(l func(event int, ctx service.Event)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lsns = append(s.lsns, l)
}

// throw handles service, server and pool events.
func (s *Service) throw(event int, ctx service.Event) {
	for _, l := range s.lsns {
		l(event, ctx)
	}
}

// Priority returns predefined service priority
func (s *Service) Priority() int {
	return s.priority
}

// Serve serves the service
func (s *Service) Serve(ctx context.Context) error {
	if s.rpc == nil {
		return errors.New("RPC service is not configured")
	}

	s.mu.Lock()
	s.serving = true
	s.stop = make(chan interface{})
	s.mu.Unlock()

	rpc.HandleHTTP()
	ln, err := s.Listener()
	if err != nil {
		return err
	}
	defer ln.Close()

	go http.Serve(ln, nil)

	/*go func() {
		for {
			select {
			case <-s.stop:
				return

			default:
				conn, err := ln.Accept()
				if err != nil {
					continue
				}

				go s.rpc.ServeCodec(codec.NewServer(conn))
			}
		}
	}()*/

	<-s.stop

	s.mu.Lock()
	s.serving = false
	s.mu.Unlock()

	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.serving {
		close(s.stop)
	}
}

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//   - exported method of exported type
//   - two arguments, both of exported type
//   - the second argument is a pointer
//   - one return value, of type error
//
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
func (s *Service) Register(name string, svc interface{}) error {
	if s.rpc == nil {
		return errors.New("RPC service is not configured")
	}

	return s.rpc.RegisterName(name, svc)
}

// Client creates new RPC client
func (s *Service) Client() (*rpc.Client, error) {
	if s.options == nil {
		return nil, errors.New("RPC service is not configured")
	}

	conn, err := s.Dialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(codec.NewClient(conn)), nil
}

// Listener creates new rpc socket Listener
func (s *Service) Listener() (net.Listener, error) {
	if s.options.Protocol != SocketTypeTCP && s.options.Protocol != SocketTypeUNIX {
		return nil, errors.Errorf("[%s] Invalid socket type \"%v\"", s.Name, s.options.Protocol)
	}

	if s.options.Protocol == SocketTypeUNIX {
		syscall.Unlink(s.options.Address())
	}

	s.throw(EventInfo, service.InfoEvent(fmt.Sprintf("[%s] Starting: Listening on %s:%s...", s.Name, s.options.Protocol, s.options.Address())))
	return net.Listen(s.options.Protocol, s.options.Address())
}

// Dialer creates rpc socket Dialer
func (s *Service) Dialer() (net.Conn, error) {
	if s.options.Protocol != SocketTypeTCP && s.options.Protocol != SocketTypeUNIX {
		return nil, errors.Errorf("Invalid socket type \"%v\"", s.options.Protocol)
	}

	return net.Dial(s.options.Protocol, s.options.Host)
}

// NewService creates a new service of type RPC
func NewService(options config.Config) (service.Interface, error) {
	return &Service{
		Name:     "RPC Server",
		serving:  false,
		priority: 1,
	}, nil
}
