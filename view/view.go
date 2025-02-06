package view

import (
	"html/template"
	"path/filepath"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}
)

// Interface is a view resource interface
type Interface interface {
	Priority() int
	GetOptions() *Config
	Render(data map[string]interface{}, script string, tpl string) ([]byte, error)
	GetPaths() map[string]map[string]string
	AddBasePath(path string, prefix string) error
	GetBasePath() string
	SetActionPath(path string) error
	GetActionPath() string
	SetActionPathNoController(path string) error
	GetActionPathNoController() string
	SetHelperPath(path string) error
	GetHelperPath() string
	SetSuffix(suffix string) error
	GetSuffix() string
	AddTemplatePath(path string) error
	GetTemplatePaths() map[string]string
	AddLayoutPath(path string) error
	GetLayoutPaths() map[string]string
	RegisterHelper(name string, hlp HelperInterface) error
	Helper(name string) HelperInterface
	Assign(params map[string]interface{}) error
	SetParam(key string, value interface{}) error
	Param(key string) interface{}
	ParamBool(key string, def bool) bool
	ParamString(key string, def string) string
	ParamInt(key string, def int) int
	AddTemplateFunc(string, interface{}) error
	SetTemplateFunc(string, interface{}) error
	TemplateFunctions() map[string]interface{}
	RemoveTemplateFunc(string) error
	PrepareLayouts() error
	PrepareTemplates() error
	GetTemplate(path string) *template.Template
}

type view struct {
	Options                        *Config
	Logger                         *log.Log
	BaseDir                        string
	ViewBasePathSpec               string
	ViewActionPathSpec             string
	ViewActionPathNoControllerSpec string
	ViewHelperPathSpec             string
	ViewSuffix                     string
	paths                          map[string]map[string]string
	helpers                        map[string]HelperInterface
	params                         map[string]interface{}
	templates                      map[string]*template.Template
	layouts                        map[string]*template.Template
	template                       *template.Template
}

// GetPaths returns all registered script pathes
func (v *view) GetPaths() map[string]map[string]string {
	return v.paths
}

// setBasePath
// AddBasePath registers a new base script path
func (v *view) AddBasePath(path string, prefix string) error {
	v.AddTemplatePath(filepath.FromSlash(path))
	//$this->addHelperPath($path . 'helpers', $classPrefix . 'Helper');
	//$this->addFilterPath($path . 'filters', $classPrefix . 'Filter');
	return nil
}

// GetBasePath returns base path pattern
func (v *view) GetBasePath() string {
	return v.ViewBasePathSpec
}

// SetActionPath sets script path pattern
func (v *view) SetActionPath(path string) error {
	v.ViewActionPathSpec = path
	return nil
}

// GetActionPath returns script path pattern
func (v *view) GetActionPath() string {
	return v.ViewActionPathSpec
}

// SetActionPathNoController sets script path pattern without controller specification
func (v *view) SetActionPathNoController(path string) error {
	v.ViewActionPathNoControllerSpec = path
	return nil
}

// SetActionPathNoController returns script path pattern
func (v *view) GetActionPathNoController() string {
	return v.ViewActionPathNoControllerSpec
}

// SetHelperPath sets view helper path pattern
func (v *view) SetHelperPath(path string) error {
	v.ViewHelperPathSpec = path
	return nil
}

// GetHelperPath returns view helper path pattern
func (v *view) GetHelperPath() string {
	return v.ViewHelperPathSpec
}

// SetSuffix sets path file suffix
func (v *view) SetSuffix(suffix string) error {
	v.ViewSuffix = suffix
	return nil
}

// GetBasePath returns path file suffix
func (v *view) GetSuffix() string {
	return v.ViewSuffix
}

// AddTemplatePath adds a path to templates
func (v *view) AddTemplatePath(path string) error {
	if _, ok := v.paths["templates"]; !ok {
		v.paths["templates"] = make(map[string]string)
	}

	v.paths["templates"][filepath.FromSlash(path)] = filepath.FromSlash(path)
	return nil
}

// GetTemplatePaths returns registered templates
func (v *view) GetTemplatePaths() map[string]string {
	return v.paths["templates"]
}

// AddLayoutPath adds a path to layout templates
func (v *view) AddLayoutPath(path string) error {
	if _, ok := v.paths["layouts"]; !ok {
		v.paths["layouts"] = make(map[string]string)
	}

	v.paths["layouts"][filepath.FromSlash(path)] = filepath.FromSlash(path)
	return nil
}

// GetLayoutPaths returns registered layout template paths
func (v *view) GetLayoutPaths() map[string]string {
	return v.paths["layouts"]
}

// AddHelperPath adds a path to helpers
func (v *view) AddHelperPath(path string, prefix string) error {
	if _, ok := v.paths["helpers"]; !ok {
		v.paths["helpers"] = make(map[string]string)
	}

	v.paths["helpers"][filepath.FromSlash(path)] = prefix
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
func (v *view) RegisterHelper(name string, hlp HelperInterface) error {
	if _, ok := v.helpers[name]; ok {
		return errors.Errorf("View helper by name '%s' is already registered", name)
	}

	v.helpers[name] = hlp
	return nil
}

// Helper returns view helper by its name
func (v *view) Helper(name string) HelperInterface {
	if hlp, ok := v.helpers[name]; ok {
		return hlp
	}

	return nil
}

// Priority returns resource initialization priority
func (v *view) Priority() int {
	return v.Options.Priority
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

// TemplateData holds information about template
type TemplateData struct {
	Name string
	Path string
	Data *template.Template
}
