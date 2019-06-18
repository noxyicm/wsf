package bootstrap

import (
	"wsf/config"
)

// Config defines set of bootstrap variables
type Config struct {
	Type      string
	AppConfig config.Config
}

// Get nested config section (sub-map), returns nil if section not found
func (c *Config) Get(key string) config.Config {
	return c.AppConfig.Get(key)
}

// GetString returns a string value
func (c *Config) GetString(key string) string {
	return c.AppConfig.GetString(key)
}

// GetKeys returns config keys
func (c *Config) GetKeys() []string {
	return c.AppConfig.GetKeys()
}

// GetAll returns config map
func (c *Config) GetAll() map[string]interface{} {
	return c.AppConfig.GetAll()
}

// Unmarshal unmarshals config data into given struct
func (c *Config) Unmarshal(out interface{}) error {
	return c.AppConfig.Unmarshal(out)
}

// Merge merges a new configuration with an existing config
func (c *Config) Merge(cfg map[string]interface{}) error {
	return c.AppConfig.Merge(cfg)
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(c)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = "default"
	return nil
}
