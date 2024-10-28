package translate

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/log"
)

// AdapterConfig defines translate adapter configuration
type AdapterConfig struct {
	Type            string
	Enable          bool
	Priority        int
	Clear           bool
	Params          map[string]interface{}
	Content         interface{}
	Ignore          map[string]string
	Locale          string
	Logger          *log.Config
	LogUntranslated bool
	Reload          bool
	Route           interface{}
	Scan            string
	UseCache        bool
	Cache           config.Config
	Tag             string
}

// Populate populates Config values using given Config source
func (c *AdapterConfig) Populate(cfg config.Config) error {
	if scfg := cfg.Get("cache"); scfg != nil {
		c.Cache = scfg
	}

	if c.Cache == nil {
		c.Cache = config.NewBridge()
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Logger == nil {
		c.Logger = &log.Config{}
	}

	c.Logger.Defaults()
	if lcfg := cfg.Get("log"); lcfg != nil {
		c.Logger.Populate(lcfg)
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *AdapterConfig) Defaults() error {
	c.Type = "default"
	c.Enable = true
	c.Priority = 3
	c.Clear = false
	c.Params = make(map[string]interface{})
	c.Content = nil
	c.Ignore = map[string]string{"root": "."}
	c.Locale = "auto"
	c.LogUntranslated = false
	c.Reload = false
	c.Route = nil
	c.Scan = "directory"
	c.Tag = "WSFTranslate"
	c.UseCache = false
	c.Cache = nil

	c.Logger = &log.Config{}
	c.Logger.Defaults()

	if c.Cache == nil {
		c.Cache = config.NewBridge()
	}

	return nil
}

// Valid validates the configuration
func (c *AdapterConfig) Valid() error {
	return nil
}
