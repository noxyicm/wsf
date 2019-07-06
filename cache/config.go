package cache

import (
	"wsf/config"
)

// Config represents dispatcher configuration
type Config struct {
	Priority                int
	Enable                  bool
	AutomaticSerialization  bool
	AutomaticCleaningFactor int64
	ExtendedBackend         bool
	WriteControl            bool
	CacheIDPrefix           string
	Backend                 config.Config
	Logger                  config.Config
	//Frontend         map[string]*frontend.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	c.Backend = cfg.Get("backend")
	c.Logger = cfg.Get("log")

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Priority = 20
	c.Enable = true

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
