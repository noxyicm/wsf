package environment

import "wsf/config"

// Config defines set of environment variables
type Config struct {
	Values map[string]string
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(&c.Values)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Values = make(map[string]string)
	return nil
}
