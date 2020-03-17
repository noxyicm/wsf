package application

import (
	"os"
	"path/filepath"
	"sync"
	"wsf/application/bootstrap"
	"wsf/config"
	"wsf/context"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
)

const (
	// EventDebug thrown if there is something insegnificant to say
	EventDebug = iota + 500

	// EventInfo thrown if there is something to say
	EventInfo

	// EventError thrown on any non job error provided
	EventError

	// VERSION represent version.
	VERSION = "0.0.0.0"

	// EnvDEV is a development mode
	EnvDEV = "development"

	// EnvPROD is a production mode
	EnvPROD = "production"

	// EnvLOC is a local mode
	EnvLOC = "local"
)

// Application struct
type Application struct {
	options     *Config
	environment string
	logger      *log.Log
	RootPath    string
	bootstrap   bootstrap.Interface
	ctx         context.Context
	lsns        []func(event int, ctx interface{})
	mu          sync.Mutex
}

// Context returns Application context
func (a *Application) Context() context.Context {
	return a.ctx
}

// Options returns config for Application
func (a *Application) Options() *Config {
	return a.options
}

// SetOptions Sets config for Application
func (a *Application) SetOptions(options *Config) error {
	if options == nil {
		return errors.New("Invalid options provided; options can't be empty")
	}

	a.options = options
	return nil
}

// Init Initializes the application
func (a *Application) Init() (err error) {
	a.ctx, err = context.NewContext(context.Background())
	if err != nil {
		return errors.New("Initialization failed. Unable to create application context")
	}

	return a.bootstrap.Init()
}

// Run serves the application
func (a *Application) Run() error {
	return a.bootstrap.Run(a.ctx)
}

// Stop shuts down the application
func (a *Application) Stop() {
	a.bootstrap.Stop()
}

// Resource returns resource registered in the application if exists
func (a *Application) Resource(name string) (interface{}, int) {
	return a.bootstrap.Resource(name)
}

// AddListener attaches event watcher
func (a *Application) AddListener(l func(event int, ctx interface{})) {
	a.lsns = append(a.lsns, l)
}

// throw handles events
func (a *Application) throw(event int, ctx interface{}) {
	for _, l := range a.lsns {
		l(event, ctx)
	}
}

// SetRootPath sets a root path of application
func SetRootPath(path string) {
	config.AppRootPath = filepath.FromSlash(path)
}

// SetAppPath sets a application path
func SetAppPath(path string) error {
	config.AppPath = filepath.Join(config.AppRootPath, filepath.FromSlash(path))
	return nil
}

// SetBasePath sets the absolute path to the app
func SetBasePath(path string) error {
	config.BasePath = filepath.Join(config.AppRootPath, filepath.FromSlash(path))
	return nil
}

// SetStaticPath sets an application static folder path
func SetStaticPath(path string) error {
	config.StaticPath = filepath.Join(config.AppRootPath, filepath.FromSlash(path))
	return nil
}

// SetCachePath sets an application cache folder path
func SetCachePath(path string) error {
	config.CachePath = filepath.Join(config.AppRootPath, filepath.FromSlash(path))
	return nil
}

// NewApplication Creates new Application struct
func NewApplication(environment string, options interface{}, override []string) (App *Application, err error) {
	app := &Application{
		environment: environment,
	}
	app.mu.Lock()
	defer app.mu.Unlock()

	logOptions := config.NewBridge()
	logOptions.Merge(map[string]interface{}{
		"writers": map[string]interface{}{
			"default": map[string]interface{}{
				"params": map[string]interface{}{
					"type": "stdout",
				},
				"formatter": map[string]interface{}{
					"type": "colorized",
				},
			},
		},
	})
	lg, err := log.NewLog(logOptions)
	if err != nil {
		return nil, err
	}
	app.logger = lg

	var cfg config.Config
	switch o := options.(type) {
	case string:
		dir := filepath.Dir(filepath.FromSlash(o))
		filename := filepath.Base(o)
		filename = filename[:len(filename)-len(filepath.Ext(filename))]
		cfg, err = config.LoadConfig(o, []string{dir}, filename, override)

	case map[string]interface{}:
		err = errors.New("Unsupported yet")
		//cfg, err = config.NewConfig(options.(map[string]interface{}), false)

	case config.Config:
		cfg = o

	default:
		err = errors.New("Invalid options provided; must be location of config file or config object")
	}

	if err != nil {
		return nil, err
	}

	config.App = cfg
	acfg := &Config{}
	acfg.Defaults()
	acfg.Environment = environment

	if appcfg := cfg.Get("application"); appcfg != nil {
		acfg.Populate(appcfg)
	}
	app.options = acfg

	config.AppEnv = acfg.Environment
	SetRootPath(acfg.RootPath)
	err = os.Chdir(acfg.RootPath)
	if err != nil {
		return nil, err
	}

	if err := SetAppPath(acfg.AppPath); err != nil {
		return nil, err
	}
	if err := SetBasePath(acfg.BasePath); err != nil {
		return nil, err
	}
	if err := SetStaticPath(acfg.StaticPath); err != nil {
		return nil, err
	}

	app.bootstrap, err = bootstrap.NewBootstrap(cfg)
	if err != nil {
		return nil, err
	}
	app.bootstrap.AddListener(app.throw)
	app.AddListener(func(event int, ctx interface{}) {
		switch event {
		case EventDebug:
			app.logger.Debug(ctx.(string), nil)

		case EventInfo:
			app.logger.Info(ctx.(string), nil)

		case EventError:
			app.logger.Error(ctx.(string), nil)
		}
	})

	registry.Set("Application", app)
	return app, nil
}
