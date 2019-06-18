package view

import (
	"html/template"
	"path/filepath"
	"wsf/config"
	"wsf/errors"
	"wsf/log"
	"wsf/view/helper"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface is a view resource interface
type Interface interface {
	Render(script string) ([]byte, error)
	GetPaths() map[string]map[string]string
	AddBasePath(path string, prefix string) error
	AddTeplatePath(path string) error
	GetTemplatePaths() map[string]string
	RegisterHelper(name string, hlp helper.Interface) error
	Helper(name string) helper.Interface
	Assign(params map[string]interface{}) error
	SetParam(key string, value interface{}) error
	Param(key string) interface{}
	ParamBool(key string, def bool) bool
	ParamString(key string, def string) string
	ParamInt(key string, def int) int
	PrepareTemplates() error
	Priority() int
}

type view struct {
	options   *Config
	logger    *log.Log
	baseDir   string
	paths     map[string]map[string]string
	helpers   map[string]helper.Interface
	params    map[string]interface{}
	templates map[string]*template.Template
}

// GetPaths returns all registered script pathes
func (v *view) GetPaths() map[string]map[string]string {
	return v.paths
}

//setBasePath
// AddBasePath registers a new base script path
func (v *view) AddBasePath(path string, prefix string) error {
	v.AddTeplatePath(filepath.FromSlash(path) + "templates")
	//$this->addHelperPath($path . 'helpers', $classPrefix . 'Helper');
	//$this->addFilterPath($path . 'filters', $classPrefix . 'Filter');
	return nil
}

// AddTeplatePath adds a path to templates
func (v *view) AddTeplatePath(path string) error {
	if _, ok := v.paths["template"]; !ok {
		v.paths["template"] = make(map[string]string)
	}

	v.paths["template"][filepath.FromSlash(path)] = filepath.FromSlash(path)
	return nil
}

// GetTemplatePaths returns registered template paths
func (v *view) GetTemplatePaths() map[string]string {
	return v.paths["template"]
}

// AddHelperPath adds a path to helpers
func (v *view) AddHelperPath(path string, prefix string) error {
	v.paths["helper"][filepath.FromSlash(path)] = prefix
	return nil
}

//setEscape
//setEncoding
//addFilter
//strictVars
//setLfiProtection

// Assign assigns variables to the view script
func (v *view) Assign(params map[string]interface{}) error {
	v.params = params
	return nil
}

// SetParam assigns variable to the view script by key
func (v *view) SetParam(key string, value interface{}) error {
	v.params[key] = value
	return nil
}

// Param returns a view parameter
func (v *view) Param(key string) interface{} {
	if v, ok := v.params[key]; ok {
		return v
	}

	return nil
}

// ParamBool returns a view parameter as bool or def
func (v *view) ParamBool(key string, def bool) bool {
	if v, ok := v.params[key]; ok {
		if v, ok := v.(bool); ok {
			return v
		}

		return def
	}

	return def
}

// ParamString returns a view parameter as string or def
func (v *view) ParamString(key string, def string) string {
	if v, ok := v.params[key]; ok {
		if v, ok := v.(string); ok {
			return v
		}

		return def
	}

	return def
}

// ParamInt returns a view parameter as int or def
func (v *view) ParamInt(key string, def int) int {
	if v, ok := v.params[key]; ok {
		if v, ok := v.(int); ok {
			return v
		}

		return def
	}

	return def
}

// RegisterHelper registers a view helper
func (v *view) RegisterHelper(name string, hlp helper.Interface) error {
	if _, ok := v.helpers[name]; ok {
		return errors.Errorf("View helper by name '%s' is already registered", name)
	}

	v.helpers[name] = hlp
	return nil
}

// Helper returns view helper by its name
func (v *view) Helper(name string) helper.Interface {
	if hlp, ok := v.helpers[name]; ok {
		return hlp
	}

	return nil
}

// Priority returns resource initialization priority
func (v *view) Priority() int {
	return v.options.Priority
}

// NewView creates a new controller specified by type
func NewView(viewType string, options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[viewType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized view type \"%v\"", viewType)
}

// Register registers a handler for controller creation
func Register(viewType string, handler func(*Config) (Interface, error)) {
	buildHandlers[viewType] = handler
}
