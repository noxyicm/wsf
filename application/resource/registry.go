package resource

import (
	"fmt"
	"reflect"
	"sync"
	"wsf/config"
	"wsf/errors"
	"wsf/registry"
)

// InitMethod Worker initialization function
const (
	InitMethod = "Init"

	// EventInfo thrown if there is something to say
	EventInfo = iota + 500

	// EventError thrown on any non job error provided
	EventError
)

var errNoConfig = errors.New("No configuration has been provided")

// Registry is a resource container interface
type Registry interface {
	Register(name string, typ string, resource Interface)
	Init(cfg config.Config) error
	Has(name string) bool
	Get(name string) (resource Interface, status int)
	Listen(l func(event int, ctx interface{}))
}

// NewRegistry creates new resource registry
func NewRegistry() Registry {
	return &resourceregistry{
		resources: make([]*bus, 0),
	}
}

type resourceregistry struct {
	resources []*bus
	lsn       func(event int, ctx interface{})
	mu        sync.Mutex
}

// Register add new resource to the registry under given name
func (r *resourceregistry) Register(name string, typ string, rsr Interface) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := len(r.resources)
	for k, v := range r.resources {
		if v.resource.Priority() > rsr.Priority() {
			key = k
			break
		}
	}

	b := &bus{
		name:     name,
		typ:      typ,
		resource: rsr,
		status:   StatusRegistered,
		order:    key,
	}

	if len(r.resources) == 0 {
		r.resources = []*bus{b}
	} else if key == 0 {
		r.resources = append([]*bus{b}, r.resources...)
	} else {
		r.resources = append(r.resources[:key], append([]*bus{b}, r.resources[key:]...)...)
	}

	for k, v := range r.resources {
		v.order = k
	}

	registry.Set(name, rsr)
}

// Init configures all underlying resources with given configuration
func (r *resourceregistry) Init(cfg config.Config) error {
	for _, resourceType := range cfg.GetKeys() {
		for _, resourceName := range cfg.Get(resourceType).GetKeys() {
			if !r.Has(resourceName) {
				rsr, err := NewResource(resourceType, cfg.Get(resourceType).Get(resourceName))
				if err != nil {
					return err
				}

				r.Register(resourceName, resourceType, rsr)
			}
		}
	}

	for _, rs := range r.resources {
		if rs.getStatus() >= StatusOK {
			return errors.Errorf("Resource [%s] has already been configured", rs.name)
		}

		if ok, err := r.initResource(rs.resource, cfg.Get(rs.typ).Get(rs.name)); err != nil {
			if err == errNoConfig {
				r.throw(EventError, "["+rs.name+"]: disabled. No configuration has been provided")
				continue
			}

			return errors.Wrap(err, fmt.Sprintf("[%s]", rs.name))
		} else if ok {
			rs.setStatus(StatusOK)
		} else {
			r.throw(EventError, "["+rs.name+"]: disabled")
		}
	}

	return nil
}

// Has cheks if specified resource registered or not
func (r *resourceregistry) Has(target string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rs := range r.resources {
		if rs.name == target {
			return true
		}
	}

	return false
}

// Get returns resource instance by it's name or nil if resource not found
func (r *resourceregistry) Get(target string) (rsr Interface, status int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rs := range r.resources {
		if rs.name == target {
			return rs.resource, rs.getStatus()
		}
	}

	return nil, StatusUndefined
}

// Listen attaches handler event watcher
func (r *resourceregistry) Listen(l func(event int, ctx interface{})) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lsn = l
}

// throw invokes event handler if any
func (r *resourceregistry) throw(event int, ctx interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.lsn != nil {
		r.lsn(event, ctx)
	}
}

// calls Init method with automatically resolved arguments
func (r *resourceregistry) initResource(rs interface{}, segment config.Config) (bool, error) {
	rf := reflect.TypeOf(rs)

	m, ok := rf.MethodByName(InitMethod)
	if !ok {
		return true, nil
	}

	if err := r.verifySignature(m); err != nil {
		return false, err
	}

	values, err := r.resolveValues(rs, m, segment)
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
func (r *resourceregistry) resolveValues(w interface{}, m reflect.Method, cfg config.Config) (values []reflect.Value, err error) {
	for i := 0; i < m.Type.NumIn(); i++ {
		v := m.Type.In(i)

		switch {
		case v.ConvertibleTo(reflect.ValueOf(w).Type()):
			values = append(values, reflect.ValueOf(w))

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
			value, err := r.resolveValue(v)
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}
	}

	return
}

// verifySignature checks if Init method has valid signature
func (r *resourceregistry) verifySignature(m reflect.Method) error {
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

func (r *resourceregistry) resolveValue(v reflect.Type) (reflect.Value, error) {
	value := reflect.Value{}
	for _, rs := range r.resources {
		if !rs.hasStatus(StatusOK) {
			continue
		}

		if v.Kind() == reflect.Interface && reflect.TypeOf(rs.resource).Implements(v) {
			if value.IsValid() {
				return value, errors.Errorf("Disambiguous dependency `%s`", v)
			}

			value = reflect.ValueOf(rs.resource)
		}

		if v.ConvertibleTo(reflect.ValueOf(rs.resource).Type()) {
			if value.IsValid() {
				return value, errors.Errorf("Disambiguous dependency `%s`", v)
			}

			value = reflect.ValueOf(rs.resource)
		}
	}

	if !value.IsValid() {
		value = reflect.New(v).Elem()
	}

	return value, nil
}
