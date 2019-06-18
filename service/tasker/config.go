package tasker

import (
	"wsf/config"
)

// Config defines RPC service config
type Config struct {
	Enable  bool
	Workers uint8
}

// Populate must populate Config values using given Config source. Must return error if Config is not valid
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults allows to init blank config with pre-defined set of default values.
func (c *Config) Defaults() error {
	c.Enable = true
	c.Workers = 1

	return nil
}

// Valid returns nil if config is valid
func (c *Config) Valid() error {
	return nil
}
