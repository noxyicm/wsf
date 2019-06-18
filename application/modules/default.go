package modules

import (
	"github.com/pkg/errors"
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
func (h *Default) RegisterModule(order int, name string, callback func(*Module)) (err error) {
	if _, ok := h.moduleOrder[order]; ok {
		order = len(h.moduleOrder)
	}

	h.moduleOrder[order] = name

	if _, ok := h.modules[name]; ok {
		return errors.Errorf("Module with name '%s' already registered", name)
	}

	h.modules[name], err = NewModule(name, order)
	if err != nil {
		return err
	}

	callback(h.modules[name])
	return nil
}

// NewDefaultModuleHandler creates new default module handler
func NewDefaultModuleHandler(options *Config) (mi Handler, err error) {
	mh := &Default{}
	mh.options = options
	mh.modules = make(map[string]*Module)
	return mh, nil
}
