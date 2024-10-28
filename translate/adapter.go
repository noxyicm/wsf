package translate

import (
	"strings"
	"github.com/noxyicm/wsf/cache"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/locale"
	"github.com/noxyicm/wsf/log"
)

const (
	// TYPEAdapterDefault resource type
	TYPEAdapterDefault = "default"
)

var (
	buildAdapterHandlers = map[string]func(*AdapterConfig) (Adapter, error){}
)

// Adapter defines the translate adapter
type Adapter interface {
	SetLogger(lg *log.Log) error
	SetLocale(lcl string) error
	Locale() *locale.Locale
	AddTranslationData(data map[string]*Entry, loc string) error
	LoadTranslationData(loc string, options map[string]interface{}) (map[string]*Entry, error)
	Translate(key string) string
	TranslateForLocale(key string, locale string) string
	Plural(singular string, double string, plural string, number int) string
	PluralForLocale(singular string, double string, plural string, number int, locale string) string
}

// NewAdapter creates a new translate adapter of provided type
func NewAdapter(adapterType string, options config.Config) (Adapter, error) {
	cfg := &AdapterConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildAdapterHandlers[adapterType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized translate adapter type \"%v\"", adapterType)
}

// NewAdapterFromConfig creates a new translate adapter of provided config
func NewAdapterFromConfig(cfg *AdapterConfig) (Adapter, error) {
	if f, ok := buildAdapterHandlers[cfg.Type]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized translate adapter type \"%v\"", cfg.Type)
}

// RegisterAdapter registers a handler for translate adapter creation
func RegisterAdapter(adapterType string, handler func(*AdapterConfig) (Adapter, error)) {
	buildAdapterHandlers[adapterType] = handler
}

// DefaultAdapter is a default translate adapter
type DefaultAdapter struct {
	Options         *AdapterConfig
	Logger          *log.Log
	Cache           cache.Interface
	Automatic       bool
	locale          *locale.Locale
	LocaleDirectory string
	LocaleFilename  string
	translate       map[string]map[string]*Entry
}

// SetLogger sets logger for adapter
func (a *DefaultAdapter) SetLogger(lg *log.Log) error {
	a.Logger = lg
	return nil
}

// SetLocale sets the default locale
func (a *DefaultAdapter) SetLocale(lcl string) error {
	if lcl == "auto" || lcl == "" {
		a.Automatic = true
	} else {
		a.Automatic = false
	}

	locale, err := locale.FindLocale(lcl)
	if err != nil {
		return errors.Wrapf(err, "The given Language '%s' does not exist", lcl)
	}

	if _, ok := a.translate[locale]; !ok {
		temp := strings.Split(locale, "_")
		if _, ok := a.translate[temp[0]]; !ok {
			a.Logger.Alertf("The language '%s' has to be added before it can be used", nil, locale)
			return errors.Wrapf(err, "The language '%s' has to be added before it can be used", locale)
		}

		locale = temp[0]
	}

	if len(a.translate[locale]) == 0 {
		a.Logger.Alertf("No translation for the language '%s' available", nil, locale)
		return errors.Wrapf(err, "No translation for the language '%s' available", locale)
	}

	if a.Options.Locale != locale {
		a.Options.Locale = locale

		/*if (isset(self::$_cache)) {
			$id = 'Zend_Translate_' . $this->toString() . '_Options';
			if (self::$_cacheTags) {
				self::$_cache->save($this->_options, $id, [$this->_options['tag']]);
			} else {
				self::$_cache->save($this->_options, $id);
			}
		}*/
	}

	return nil
}

// Locale returns preseted locale
func (a *DefaultAdapter) Locale() *locale.Locale {
	return a.locale
}

// Translate the provided message
func (a *DefaultAdapter) Translate(key string) string {
	return key
}

// TranslateForLocale provided message for specific locale
func (a *DefaultAdapter) TranslateForLocale(key string, locale string) string {
	return key
}

// Plural translates message depending on plurality of number
func (a *DefaultAdapter) Plural(singular string, double string, plural string, number int) string {
	return singular
}

// PluralForLocale translates message for specific locale depending on plurality of number
func (a *DefaultAdapter) PluralForLocale(singular string, double string, plural string, number int, locale string) string {
	return singular
}

// NewDefaultAdapter creates a new translate adapter of type default
func NewDefaultAdapter(cfg *AdapterConfig) (*DefaultAdapter, error) {
	d := &DefaultAdapter{
		Options:         cfg,
		LocaleDirectory: "directory",
		LocaleFilename:  "filename",
		translate:       make(map[string]map[string]*Entry),
	}

	lg, err := log.NewLogFromConfig(cfg.Logger)
	if err == nil {
		return nil, errors.Wrap(err, "Unable to create translate adapter: Unable to create logger")
	}
	d.Logger = lg

	if cfg.UseCache {
		ccfg := &cache.Config{}
		ccfg.Defaults()
		ccfg.Populate(cfg.Cache)

		cch, err := cache.NewCore(cfg.Cache.GetString("type"), cfg.Cache)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to create translate adapter: Unable to create cache")
		}
		d.Cache = cch

		if ok, err := d.Cache.Init(ccfg); !ok {
			return nil, errors.Wrap(err, "Unable to create translate adapter: Unable to initialize cache")
		}
	}

	return d, nil
}

// Entry is a translatable entry
type Entry struct {
	Single   string
	Double   string
	Plural   string
	Multiple bool
}
