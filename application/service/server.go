package service

import (
	"fmt"
	"reflect"
	"sync"
	"wsf/config"
	"wsf/context"
	"wsf/errors"
	"wsf/registry"
	"wsf/service"
)

// InitMethod Worker initialization function
const (
	// EventDebug thrown if there is something insegnificant to say
	EventDebug = iota + 500

	// EventInfo thrown if there is something to say
	EventInfo

	// EventError thrown on any non job error provided
	EventError

	InitMethod = "Init"
)

var errNoConfig = errors.New("No configuration has been provided")

// Server interface exposing its services
type Server interface {
	Register(name string, typ string, worker service.Interface)
	Init(cfg config.Config) error
	Has(service string) bool
	Get(service string) (svc interface{}, status int)
	Serve(ctx context.Context) error
	Stop()
	Listen(l func(event int, ctx interface{}))
}

// NewServer creates new server
func NewServer() Server {
	return &server{
		services: make([]*bus, 0),
	}
}

type server struct {
	services []*bus
	lsn      func(event int, ctx interface{})
	mu       sync.Mutex
}

// Register add new service to the server under given name
func (s *server) Register(name string, typ string, svc service.Interface) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := len(s.services)
	for k, v := range s.services {
		if v.service.Priority() > svc.Priority() {
			key = k
			break
		}
	}

	b := &bus{
		name:    name,
		typ:     typ,
		service: svc,
		status:  StatusRegistered,
		order:   key,
	}

	if len(s.services) == 0 {
		s.services = []*bus{b}
	} else if key == 0 {
		s.services = append([]*bus{b}, s.services...)
	} else {
		s.services = append(s.services[:key], append([]*bus{b}, s.services[key:]...)...)
	}

	for k, v := range s.services {
		v.order = k
	}

	registry.Set("service."+name, svc)
	s.throw(EventDebug, fmt.Sprintf("Service '%s' registered", name))
}

// Init configures all underlying services with given configuration
func (s *server) Init(cfg config.Config) error {
	for _, serviceType := range cfg.GetKeys() {
		for _, serviceName := range cfg.Get(serviceType).GetKeys() {
			if !s.Has(serviceName) {
				svc, err := NewService(serviceType, cfg.Get(serviceType).Get(serviceName))
				if err != nil {
					return err
				}

				s.Register(serviceName, serviceType, svc)
			}
		}
	}

	for _, b := range s.services {
		if b.getStatus() >= StatusOK {
			return errors.Errorf("Service [%s] has already been configured", b.name)
		}

		if ok, err := s.initService(b.service, cfg.Get(b.typ).Get(b.name)); err != nil {
			if err == errNoConfig {
				s.throw(EventError, fmt.Sprintf("Service '%s' disabled: %v\n", b.name, errNoConfig))
				continue
			}

			return err
		} else if ok {
			b.setStatus(StatusOK)
			b.service.AddListener(s.throw)
		} else {
			s.throw(EventError, "Service '"+b.name+"' disabled. No configuration has been provided")
		}
	}

	return nil
}

// Has cheks if specified service registered or not
func (s *server) Has(target string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, b := range s.services {
		if b.name == target {
			return true
		}
	}

	return false
}

// Get returns service instance by it's name or nil if service not found
func (s *server) Get(target string) (svc interface{}, status int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, b := range s.services {
		if b.name == target {
			return b.service, b.getStatus()
		}
	}

	return nil, StatusUndefined
}

// Serve all configured services
func (s *server) Serve(ctx context.Context) error {
	var (
		numServing = 0
		done       = make(chan interface{}, len(s.services))
	)

	for _, b := range s.services {
		if b.hasStatus(StatusOK) && b.canServe() {
			numServing++
		} else {
			continue
		}

		s.throw(EventDebug, fmt.Sprintf("Service '%s' started", b.name))
		go func(b *bus) {
			b.setStatus(StatusServing)
			defer b.setStatus(StatusStopped)

			if err := b.service.Serve(ctx); err != nil {
				done <- err
			} else {
				done <- nil
			}
		}(b)
	}

	for i := 0; i < numServing; i++ {
		result := <-done

		if result == nil {
			continue
		}

		// found an error in one of the services, stopping the rest of running services
		if err := result.(error); err != nil {
			s.Stop()
			return err
		}
	}

	return nil
}

// Stop sends stop command to all running services
func (s *server) Stop() {
	for _, b := range s.services {
		if b.hasStatus(StatusServing) {
			b.service.Stop()
			b.setStatus(StatusStopped)

			s.throw(EventDebug, fmt.Sprintf("Service '%s' stopped", b.name))
		}
	}
}

// Listen attaches handler event watcher
func (s *server) Listen(l func(event int, ctx interface{})) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lsn = l
}

// throw invokes event handler if any
func (s *server) throw(event int, ctx interface{}) {
	if s.lsn != nil {
		s.lsn(event, ctx)
	}
}

// calls Init method with automatically resolved arguments
func (s *server) initService(b interface{}, segment config.Config) (bool, error) {
	r := reflect.TypeOf(b)

	m, ok := r.MethodByName(InitMethod)
	if !ok {
		return true, nil
	}

	if err := s.verifySignature(m); err != nil {
		return false, err
	}

	values, err := s.resolveValues(b, m, segment)
	if err != nil {
		return false, err
	}

	out := m.Func.Call(values)
	if out[1].IsNil() {
		return out[0].Bool(), nil
	}

	return out[0].Bool(), out[1].Interface().(error)
}

// resolveValues returns slice of call arguments for service Init method
func (s *server) resolveValues(b interface{}, m reflect.Method, cfg config.Config) (values []reflect.Value, err error) {
	for i := 0; i < m.Type.NumIn(); i++ {
		v := m.Type.In(i)

		switch {
		case v.ConvertibleTo(reflect.ValueOf(b).Type()):
			values = append(values, reflect.ValueOf(b))

		case v.Implements(reflect.TypeOf((*Server)(nil)).Elem()):
			values = append(values, reflect.ValueOf(s))

			//case v.Implements(reflect.TypeOf((*logrus.StdLogger)(nil)).Elem()),
			//	v.Implements(reflect.TypeOf((*logrus.FieldLogger)(nil)).Elem()),
			//	v.ConvertibleTo(reflect.ValueOf(s.log).Type()): // logger
			//	values = append(values, reflect.ValueOf(s.log))

		case v.Implements(reflect.TypeOf((*config.PopulatableConfig)(nil)).Elem()):
			sc := reflect.New(v.Elem())

			if dsc, ok := sc.Interface().(config.DefaultConfig); ok {
				dsc.Defaults()
				if cfg == nil {
					values = append(values, sc)
					continue
				}
			} else if cfg == nil {
				return nil, errNoConfig
			}

			if err := sc.Interface().(config.PopulatableConfig).Populate(cfg); err != nil {
				return nil, err
			}

			values = append(values, sc)

		case v.Implements(reflect.TypeOf((*config.Config)(nil)).Elem()):
			if cfg == nil {
				return nil, errNoConfig
			}

			values = append(values, reflect.ValueOf(cfg))

		default:
			value, err := s.resolveValue(v)
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}
	}

	return
}

// verifySignature checks if Init method has valid signature
func (s *server) verifySignature(m reflect.Method) error {
	if m.Type.NumOut() != 2 {
		return errors.New("Method Init must have exact 2 return values")
	}

	if m.Type.Out(0).Kind() != reflect.Bool {
		return errors.New("First return value of Init method must be bool type")
	}

	if !m.Type.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return errors.New("Second return value of Init method value must be error type")
	}

	return nil
}

func (s *server) resolveValue(v reflect.Type) (reflect.Value, error) {
	value := reflect.Value{}
	for _, b := range s.services {
		if !b.hasStatus(StatusOK) {
			continue
		}

		if v.Kind() == reflect.Interface && reflect.TypeOf(b.service).Implements(v) {
			if value.IsValid() {
				return value, errors.Errorf("Disambiguous dependency `%s`", v)
			}

			value = reflect.ValueOf(b.service)
		}

		if v.ConvertibleTo(reflect.ValueOf(b.service).Type()) {
			if value.IsValid() {
				return value, errors.Errorf("Disambiguous dependency `%s`", v)
			}

			value = reflect.ValueOf(b.service)
		}
	}

	if !value.IsValid() {
		value = reflect.New(v).Elem()
	}

	return value, nil
}
