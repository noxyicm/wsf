package translate

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/log"
)

// Config defines translate configuration
type Config struct {
	Type     string
	Enable   bool
	Priority int
	Locale   string
	Locales  []string
	Logger   *log.Config
	Adapter  *AdapterConfig
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Adapter == nil {
		c.Adapter = &AdapterConfig{}
	}

	c.Adapter.Defaults()
	if acfg := cfg.Get("adapter"); acfg != nil {
		c.Adapter.Populate(acfg)
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
func (c *Config) Defaults() error {
	c.Type = "default"
	c.Enable = true
	c.Priority = 3
	c.Locale = "uk_UA"
	c.Locales = []string{"uk"}

	c.Adapter = &AdapterConfig{}
	c.Adapter.Defaults()

	c.Logger = &log.Config{}
	c.Logger.Defaults()

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
