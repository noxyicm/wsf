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
	buildModules          = map[string]func(order int, name string) (Interface, error){}
	buildModulesOrder     = map[int]string{}
	buildModulesCallbacks = map[string]func(m Interface) error{}
)

// Handler represents module handler interface
type Handler interface {
	Bootstrap() error
	Modules() map[string]Interface
	Module(name string) Interface
	RegisterModule(order int, name string, constructor func(order int, name string) (Interface, error), callback func(Interface) error) (err error)
	Priority() int
}

type handler struct {
	options     *Config
	namespace   string
	moduleOrder map[int]string
	modules     map[string]Interface
}

// Init initializes handler and its modules
func (h *handler) Init(options config.Config) (bool, error) {
	h.moduleOrder = buildModulesOrder
	for mo, mn := range h.moduleOrder {
		if constr, ok := buildModules[mn]; ok {
			m, err := constr(mo, mn)
			if err != nil {
				return false, errors.Errorf("Unable to create module '%s'", mn)
			}

			h.modules[mn] = m
		} else {
			return false, errors.Errorf("No constructor for module '%s'", mn)
		}
	}

	for mn, md := range h.modules {
		if err := md.InitControllers(); err != nil {
			return false, errors.Wrapf(err, "Unabele to initialize module '%s'", mn)
		}

		if err := md.InitPlugins(); err != nil {
			return false, errors.Wrapf(err, "Unabele to initialize module '%s'", mn)
		}

		if err := md.InitRoutes(); err != nil {
			return false, errors.Wrapf(err, "Unabele to initialize module '%s'", mn)
		}

		if err := md.InitActionHelpers(); err != nil {
			return false, errors.Wrapf(err, "Unabele to initialize module '%s'", mn)
		}

		if v, ok := buildModulesCallbacks[mn]; ok && v != nil {
			if err := v(md); err != nil {
				return false, errors.Wrapf(err, "Unabele to initialize module '%s'", mn)
			}
		}
	}

	return true, nil
}

// Modules returns handler modules
func (h *handler) Modules() map[string]Interface {
	return h.modules
}

// Module returns specific module
func (h *handler) Module(name string) Interface {
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
func NewHandler(typ string, options config.Config) (Handler, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[typ]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized module handler type \"%v\"", typ)
}

// Register registers a handler constructor for module handler creation
func Register(typ string, handler func(*Config) (Handler, error)) {
	buildHandlers[typ] = handler
}

// RegisterModule registers a module in handler
func RegisterModule(order int, name string, constructor func(order int, name string) (Interface, error), callback func(m Interface) error) error {
	if _, ok := buildModulesOrder[order]; ok {
		order = len(buildModulesOrder)
	}

	buildModulesOrder[order] = name
	buildModules[name] = constructor
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
