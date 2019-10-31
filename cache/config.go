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
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if bcfg := cfg.Get("backend"); bcfg != nil {
		c.Backend = bcfg
	}

	if lcfg := cfg.Get("log"); lcfg != nil {
		c.Logger = lcfg
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Priority = 20
	c.Enable = true

	if c.Backend == nil {
		c.Backend = config.NewBridge()
	}

	if c.Logger == nil {
		c.Logger = config.NewBridge()
	}

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
