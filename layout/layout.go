package layout

import (
	"html/template"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/controller"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/filter"
	"github.com/noxyicm/wsf/registry"
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
	InitMvc() error
	InitPlugin() error
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
	Render(ctx context.Context, script string) ([]byte, error)
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
}

// Init the layout
func (l *DefaultLayout) Init(options *Config) (b bool, err error) {
	l.Options = options
	l.Enabled = options.Enabled
	l.ViewBasePath = options.ViewBasePath
	l.ViewScriptPath = options.ViewScriptPath
	l.ViewSuffix = options.ViewSuffix
	l.ContentKey = options.ContentKey
	l.HelperName = options.HelperName
	l.PluginName = options.PluginName
	l.InflectorTarget = options.InflectorTarget
	l.Container = placeholder.GetRegistry().GetContainer(l.Options.ContentKey)

	if err := l.InitMvc(); err != nil {
		return false, err
	}

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

	return true, nil
}

// Priority of the resource
func (l *DefaultLayout) Priority() int {
	return l.Options.Priority
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

// InitMvc initialize MVC
func (l *DefaultLayout) InitMvc() error {
	viewResource := registry.GetResource("view")
	if v, ok := viewResource.(view.Interface); ok {
		if err := v.AddLayoutPath(l.GetViewScriptPath()); err != nil {
			return err
		}
	}

	// if err := l.InitPlugin(); err != nil {
	// 	return err
	// }

	if err := l.InitHelper(); err != nil {
		return err
	}

	return nil
}

// InitPlugin initialize layout plugin
func (l *DefaultLayout) InitPlugin() error {
	if !controller.HasPlugin(l.Options.PluginName) {
		plg, err := NewLayoutPlugin()
		if err != nil {
			return err
		}

		plg.(*Plugin).SetLayout(l)
		controller.RegisterPlugin(plg, 99)
	}

	return nil
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
func (l *DefaultLayout) Render(ctx context.Context, name string) ([]byte, error) {
	if name == "" {
		name = l.Options.Layout
	}

	if inf := l.GetInflector(); inf != nil {
		if str, err := inf.Filter(map[string]string{"script": name}); err == nil {
			name = str.(string)
		}
	}

	view := l.GetView()
	if view == nil {
		return nil, errors.New("[Layout] view is not set")
	}

	return view.Render(ctx, l.GetViewScriptPath()+name, name)
}

// NewLayoutDefault creates a new default layout
func NewLayoutDefault(options *Config) (Interface, error) {
	return &DefaultLayout{
		Options:   options,
		Values:    make(map[string]interface{}),
		Templates: make(map[string]*template.Template),
	}, nil
}
