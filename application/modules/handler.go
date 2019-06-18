package modules

import (
	"go/build"
	"path/filepath"
	"wsf/config"
	"wsf/errors"
	"wsf/registry"
)

var (
	buildHandlers         = map[string]func(*Config) (Handler, error){}
	buildModules          = map[int]string{}
	buildModulesCallbacks = map[string]func(m *Module) error{}
)

// Handler represents module handler interface
type Handler interface {
	Bootstrap() error
	Modules() map[string]*Module
	Module(name string) *Module
	RegisterModule(order int, name string, callback func(*Module)) error
	Priority() int
}

type handler struct {
	options     *Config
	namespace   string
	moduleOrder map[int]string
	modules     map[string]*Module
}

// Init initializes handler and its modules
func (h *handler) Init(options config.Config) (bool, error) {
	h.moduleOrder = buildModules
	for mo, mn := range h.moduleOrder {
		m, err := NewModule(mn, mo)
		if err != nil {
			return false, errors.Errorf("Unable to create module '%s'\n", mn)
		}

		h.modules[mn] = m
	}

	for mn, md := range h.modules {
		if v, ok := buildModulesCallbacks[mn]; ok {
			err := v(md)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// Modules returns handler modules
func (h *handler) Modules() map[string]*Module {
	return h.modules
}

// Module returns specific module
func (h *handler) Module(name string) *Module {
	if v, ok := h.modules[name]; ok {
		return v
	}

	return nil
}

// Priority returns resource initialization priority
func (h *handler) Priority() int {
	return h.options.Priority
}

// NewHandler creates a new module handler specified by type
func NewHandler(moduleType string, options config.Config) (Handler, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[moduleType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized module handler type \"%v\"", moduleType)
}

// Register registers a handler for module handler creation
func Register(moduleType string, handler func(*Config) (Handler, error)) {
	buildHandlers[moduleType] = handler
}

// RegisterModule registers a module in handler
func RegisterModule(order int, name string, callback func(m *Module) error) error {
	//, callback func(*Module)
	//if handler := registry.Get("moduleHandler"); handler != nil {
	//	return handler.(Handler).RegisterModule(order, name, callback)
	//}

	if _, ok := buildModules[order]; ok {
		order = len(buildModules)
	}

	buildModules[order] = name
	buildModulesCallbacks[name] = callback
	return nil
}

// ResolveImportPath returns the filesystem path for the given import path
// Returns an error if the import path could not be found
func ResolveImportPath(importPath string) (string, error) {
	if registry.GetBool("packaged") {
		return filepath.Join(registry.GetString("SourcePath"), importPath), nil
	}

	modPkg, err := build.Import(importPath, registry.GetString("AppPath"), build.FindOnly)
	if err != nil {
		return "", err
	}

	return modPkg.Dir, nil
}
