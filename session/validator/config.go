package validator

import "wsf/config"

// Config defines set of session validator
type Config struct {
	Type string
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(c)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	return nil
}
