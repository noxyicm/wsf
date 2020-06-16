package modules

import (
	"wsf/errors"
)

const (
	// TYPEDefault represents default dispatcher
	TYPEDefault = "default"
)

func init() {
	Register(TYPEDefault, NewDefaultModuleHandler)
}

// Default is a default dispatcher
type Default struct {
	handler
}

// Bootstrap loads and initializes modules
func (h *Default) Bootstrap() error {
	return nil
}

// RegisterModule registers a module in handler for initialization
func (h *Default) RegisterModule(order int, name string, constructor func(order int, name string) (Interface, error), callback func(Interface) error) (err error) {
	if _, ok := h.modules[name]; ok {
		return errors.Errorf("Module with name '%s' already registered", name)
	}

	if _, ok := h.moduleOrder[order]; ok {
		order = len(h.moduleOrder)
	}

	h.moduleOrder[order] = name
	h.modules[name], err = constructor(order, name)
	if err != nil {
		return errors.Wrapf(err, "Unable to create module '%s'", name)
	}

	return callback(h.modules[name])
}

// NewDefaultModuleHandler creates new default module handler
func NewDefaultModuleHandler(options *Config) (mi Handler, err error) {
	mh := &Default{}
	mh.options = options
	mh.modules = make(map[string]Interface)
	return mh, nil
}
