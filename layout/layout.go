package layout

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/filter"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view"
	"github.com/noxyicm/wsf/view/helper/placeholder"
	"github.com/noxyicm/wsf/view/helper/placeholder/container"
)

// Public constants
const (
	TYPELayoutDefault = "default"
)

var (
	buildHandlers = map[string]func(*Config) (Interface, error){}

	mvc Interface
)

func init() {
	Register(TYPELayoutDefault, NewLayoutDefault)
}

// Interface interface
type Interface interface {
	Init(cfg *Config) (bool, error)
	Priority() int
	SetView(v view.Interface) error
	GetView() view.Interface
	SetViewBasePath(path string) error
	GetViewBasePath() string
	SetViewScriptPath(path string) error
	GetViewScriptPath() string
	SetContentKey(key string) error
	GetContentKey() string
	SetHelperName(name string) error
	GetHelperName() string
	SetPluginName(name string) error
	GetPluginName() string
	SetViewSuffix(suffix string) error
	GetViewSuffix() string
	SetInflectorTarget(target string) error
	GetInflectorTarget() string
	SetInflector(inf *filter.Inflector) error
	GetInflector() *filter.Inflector
	IsEnabled() bool
	Assign(key string, value interface{}) error
	Get(key string) interface{}
	Populate(data map[string]interface{})
	Render(data map[string]interface{}, script string) ([]byte, error)
	GetOptions() *Config
}

// NewLayout creates a new layout
func NewLayout(layoutType string, options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildHandlers[layoutType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized layout type \"%v\"", layoutType)
}

// Register registers a handler for layout creation
func Register(layoutType string, handler func(*Config) (Interface, error)) {
	buildHandlers[layoutType] = handler
}

// SetInstance sets a layout instance
func SetInstance(lt Interface) {
	mvc = lt
}

// Instance returns a layout instance
func Instance() Interface {
	return mvc
}

// DefaultLayout is a default github.com/noxyicm/wsf layout
type DefaultLayout struct {
	Options         *Config
	Enabled         bool
	Container       container.Interface
	Inflector       *filter.Inflector
	View            view.Interface
	ContentKey      string
	ViewScriptPath  string
	ViewBasePath    string
	ViewSuffix      string
	InflectorTarget string
	HelperName      string
	PluginName      string
	Values          map[string]interface{}
	Templates       map[string]*template.Template
	paths           map[string]map[string]string
}

// Init the layout
func (l *DefaultLayout) Init(options *Config) (b bool, err error) {
	l.Options = options
	l.Enabled = options.Enabled
	l.ViewBasePath = options.ViewBasePath
	l.ViewScriptPath = options.ViewScriptPath
	l.ViewSuffix = options.ViewSuffix
	l.ContentKey = options.ContentKey
	l.InflectorTarget = options.InflectorTarget
	l.Container = placeholder.GetRegistry().GetContainer(l.Options.ContentKey)

	SetInstance(l)
	return true, nil
}

// Setup resource
func (l *DefaultLayout) Setup() (bool, error) {
	viewResource := registry.GetResource("view")
	if v, ok := viewResource.(view.Interface); ok {
		if err := l.SetView(v); err != nil {
			return false, err
		}
	} else {
		return false, errors.New("[Layout] View resource is not configured")
	}

	if err := l.AddLayoutPath(l.ViewScriptPath); err != nil {
		return false, errors.Wrapf(err, "[Layout] Unable to add layout path")
	}

	err := l.prepareLayouts()
	if err != nil {
		return false, err
	}

	return true, nil
}

// Priority of the resource
func (l *DefaultLayout) Priority() int {
	return l.Options.Priority
}

func (l *DefaultLayout) GetOptions() *Config {
	return l.Options
}

// prepareLayouts parses a layout templates files
func (l *DefaultLayout) prepareLayouts() error {
	for _, path := range l.paths["layouts"] {
		err := utils.WalkDirectoryDeep(filepath.Join(config.AppPath, filepath.FromSlash(path)), filepath.Join(config.AppPath, filepath.FromSlash(path)), l.readLayouts)
		if err != nil {
			switch err.(type) {
			case *os.PathError:
				return errors.Wrap(err, "[Layout] Unable to read layout directory")

			default:
				return err
			}
		}
	}

	return nil
}

// readLayouts loads and parses layout template into memory, extending
// all existing templates
func (l *DefaultLayout) readLayouts(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("Scanning source '%s' failed: %v", path, err)
	}

	if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
		return nil
	}

	tplFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer tplFile.Close()

	tplRaw := make([]byte, info.Size())
	_, err = tplFile.Read(tplRaw)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(config.AppPath, path)
	if err != nil {
		return err
	}

	l.Templates[relPath], err = template.New(l.ContentKey).Funcs(template.FuncMap(l.View.TemplateFunctions())).Parse(string(tplRaw))
	if err != nil {
		return err
	}

	return nil
}

// SetView sets a reference for view to layout
func (l *DefaultLayout) SetView(v view.Interface) error {
	l.View = v
	return nil
}

// GetView returns assosiated view
func (l *DefaultLayout) GetView() view.Interface {
	return l.View
}

// InitHelper initialize layout helper
func (l *DefaultLayout) InitHelper() error {
	return nil
	//if !controller. hasHelper()) {
	//	Zend_Controller_Action_HelperBroker::getStack()->offsetSet(-90, new $helperClass($this));
	//}
}

// SetViewBasePath sets a base view path
func (l *DefaultLayout) SetViewBasePath(path string) error {
	l.ViewBasePath = path
	return nil
}

// GetViewBasePath returns a base view path
func (l *DefaultLayout) GetViewBasePath() string {
	return l.ViewBasePath
}

// SetViewScriptPath sets a script view path
func (l *DefaultLayout) SetViewScriptPath(path string) error {
	l.ViewScriptPath = path
	return nil
}

// GetViewScriptPath returns a script view path
func (l *DefaultLayout) GetViewScriptPath() string {
	return l.ViewScriptPath
}

// SetContentKey sets a content key
func (l *DefaultLayout) SetContentKey(key string) error {
	l.ContentKey = key
	return nil
}

// GetContentKey returns a content key
func (l *DefaultLayout) GetContentKey() string {
	return l.ContentKey
}

// SetHelperName sets a name of action controller layout helper
func (l *DefaultLayout) SetHelperName(name string) error {
	l.HelperName = name
	return nil
}

// GetHelperName returns a name of action controller layout helper
func (l *DefaultLayout) GetHelperName() string {
	return l.HelperName
}

// SetPluginName sets a name of controller layout plugin
func (l *DefaultLayout) SetPluginName(name string) error {
	l.PluginName = name
	return nil
}

// GetPluginName returns a name of controller layout plugin
func (l *DefaultLayout) GetPluginName() string {
	return l.PluginName
}

// SetViewSuffix sets a suffix of template files
func (l *DefaultLayout) SetViewSuffix(suffix string) error {
	l.ViewSuffix = suffix
	return nil
}

// GetViewSuffix returns a template files suffix
func (l *DefaultLayout) GetViewSuffix() string {
	return l.ViewSuffix
}

// SetInflectorTarget sets an inflector target pattern
func (l *DefaultLayout) SetInflectorTarget(target string) error {
	l.InflectorTarget = target
	return nil
}

// GetInflectorTarget returns an inflector target pattern
func (l *DefaultLayout) GetInflectorTarget() string {
	return l.InflectorTarget
}

// SetInflector sets inflector
func (l *DefaultLayout) SetInflector(inf *filter.Inflector) error {
	l.Inflector = inf
	return nil
}

// GetInflector returns inflector
func (l *DefaultLayout) GetInflector() *filter.Inflector {
	if l.Inflector == nil {
		inf, err := filter.NewInflector()
		if err != nil {
			return nil
		}

		inf.AddRules(map[string]interface{}{
			":script": []interface{}{
				"Word_CamelCaseToDash",
				"StringToLower",
			},
		})
		inf.SetStaticRule("suffix", l.GetViewSuffix())
		inf.SetTarget(l.InflectorTarget)
		l.SetInflector(inf)
	}

	return l.Inflector
}

// IsEnabled returns true if layout is enabled
func (l *DefaultLayout) IsEnabled() bool {
	return l.Enabled
}

// GetMvcSuccessfulActionOnly returns true
func (l *DefaultLayout) GetMvcSuccessfulActionOnly() bool {
	return false
}

// AddLayoutPath adds a path to layout templates
func (l *DefaultLayout) AddLayoutPath(path string) error {
	if _, ok := l.paths["layouts"]; !ok {
		l.paths["layouts"] = make(map[string]string)
	}

	l.paths["layouts"][filepath.FromSlash(path)] = filepath.FromSlash(path)
	return nil
}

// GetLayoutPaths returns registered layout template paths
func (l *DefaultLayout) GetLayoutPaths() map[string]string {
	return l.paths["layouts"]
}

// Assign variable to layout
func (l *DefaultLayout) Assign(key string, value interface{}) error {

	return l.Container.Append(key, value)
}

// Get the stored value
func (l *DefaultLayout) Get(key string) interface{} {
	return l.Container.Get(key)
}

// Populate container with data
func (l *DefaultLayout) Populate(data map[string]interface{}) {
	for k, v := range data {
		l.Container.Append(k, v)
	}
}

// Render layout
func (l *DefaultLayout) Render(data map[string]interface{}, name string) ([]byte, error) {
	if name == "" {
		name = l.Options.Layout
	}

	if inf := l.GetInflector(); inf != nil {
		if str, err := inf.Filter(map[string]string{"script": name}); err == nil {
			name = str.(string)
		}
	}

	script := l.GetViewScriptPath() + name
	if t, ok := l.Templates[script]; ok {
		wr := &bytes.Buffer{}
		err := t.ExecuteTemplate(wr, l.ContentKey, data)
		if err != nil {
			return nil, errors.Wrapf(err, "[Layout] Unable to execute template '%s'", script)
		}

		b := make([]byte, wr.Len())
		_, err = wr.Read(b)
		if err != nil {
			return nil, errors.Wrapf(err, "[Layout] Unable to execute template '%s'", script)
		}

		return b, nil
	}

	return nil, errors.Errorf("[Layout] Template by name '%s' not found", script)
}

// NewLayoutDefault creates a new default layout
func NewLayoutDefault(options *Config) (Interface, error) {
	return &DefaultLayout{
		Options:   options,
		Values:    make(map[string]interface{}),
		Templates: make(map[string]*template.Template),
		paths:     make(map[string]map[string]string),
	}, nil
}
