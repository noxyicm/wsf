package translate

import (
	"wsf/config"
	"wsf/errors"
	"wsf/locale"
	"wsf/log"
)

const (
	// TYPEDefault resource name
	TYPEDefault = "default"
)

var (
	buildHandlers = map[string]func(config.Config) (Interface, error){}

	inst Interface
)

func init() {
	Register(TYPEDefault, NewDefault)
}

// Interface defines the translate resource
type Interface interface {
	Priority() int
	Enabled() bool
	SetLocale(lcl string) error
	Locale() *locale.Locale
	SetAdapter(adp Adapter) error
	Adapter() Adapter
	Translate(key string) string
	TranslateForLocale(key string, locale string) string
	Plural(singular string, double string, plural string, number int) string
	PluralForLocale(singular string, double string, plural string, number int, locale string) string
}

// Default is a default translate
type Default struct {
	Options         *Config
	Logger          *log.Log
	AdapterInstance Adapter
	Automatic       bool
	translate       map[string]map[string]*Entry
}

// Init resource
func (t *Default) Init(options *Config) (bool, error) {
	t.Options = options

	lg, err := log.NewLogFromConfig(t.Options.Logger)
	if err == nil {
		return false, errors.Wrap(err, "Unable to initialize translate: Unable to create logger")
	}
	t.Logger = lg

	adp, err := NewAdapterFromConfig(t.Options.Adapter)
	if err != nil {
		return false, errors.Wrap(err, "Unable to initialize translate")
	}
	t.AdapterInstance = adp
	t.AdapterInstance.SetLogger(t.Logger)

	return true, nil
}

// Setup setups handler and its modules
func (t *Default) Setup() (bool, error) {
	for _, loc := range t.Options.Locales {
		data, err := t.AdapterInstance.LoadTranslationData(loc, nil)
		if err != nil {
			return false, errors.Wrapf(err, "Unable to load translation data for '%s' language", loc)
		}

		if err := t.AdapterInstance.AddTranslationData(data, loc); err != nil {
			return false, errors.Wrapf(err, "Unable to add translation data for '%s' language", loc)
		}
	}

	return true, nil
}

// Priority returns resource initialization priority
func (t *Default) Priority() int {
	return t.Options.Priority
}

// Enabled returns true if ACL is enabled
func (t *Default) Enabled() bool {
	return t.Options.Enable
}

// AddTranslation adds new translations
// This may be a new language or additional content for an existing language
// If the key 'clear' is true, then translations for the specified
// language will be replaced and added otherwise
func (t *Default) AddTranslation() error {
	return nil
}

// SetLocale sets the default locale
func (t *Default) SetLocale(lcl string) error {
	if err := t.AdapterInstance.SetLocale(lcl); err != nil {
		return errors.Wrapf(err, "Unable to set locale")
	}

	return nil
}

// Locale returns preseted locale
func (t *Default) Locale() *locale.Locale {
	return t.AdapterInstance.Locale()
}

// SetAdapter sets the translate adapter
func (t *Default) SetAdapter(adp Adapter) error {
	t.AdapterInstance = adp
	return nil
}

// Adapter returns translate adapter
func (t *Default) Adapter() Adapter {
	return t.AdapterInstance
}

// Translate the seted key
func (t *Default) Translate(key string) string {
	return t.AdapterInstance.Translate(key)
}

// TranslateForLocale translate the seted key for specifyed locale
func (t *Default) TranslateForLocale(key string, locale string) string {
	return t.AdapterInstance.TranslateForLocale(key, locale)
}

// Plural translate the seted key
func (t *Default) Plural(singular string, double string, plural string, number int) string {
	return t.AdapterInstance.Plural(singular, double, plural, number)
}

// PluralForLocale translate the seted key for specifyed locale
func (t *Default) PluralForLocale(singular string, double string, plural string, number int, locale string) string {
	return t.AdapterInstance.PluralForLocale(singular, double, plural, number, locale)
}

// NewDefault creates a new acl of type default
func NewDefault(options config.Config) (Interface, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	t := &Default{
		Options:   cfg,
		translate: make(map[string]map[string]*Entry),
	}

	return t, nil
}

// NewTranslate creates a new Translate resource of type typ
func NewTranslate(typ string, options config.Config) (Interface, error) {
	if f, ok := buildHandlers[typ]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized Translate type '%s'", typ)
}

// Register registers a handler for acl creation
func Register(typ string, handler func(config.Config) (Interface, error)) {
	buildHandlers[typ] = handler
}

// SetInstance sets global instance
func SetInstance(a Interface) {
	inst = a
}

// Instance returns global instance
func Instance() Interface {
	return inst
}
