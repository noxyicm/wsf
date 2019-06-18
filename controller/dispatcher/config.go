package dispatcher

import (
	"wsf/config"
)

// Config represents dispatcher configuration
type Config struct {
	Type                       string
	defaultModule              string
	defaultController          string
	defaultAction              string
	useDefaultControllerAlways bool
	plugins                    []string
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = "default"
	c.defaultModule = "default"
	c.defaultController = "front"
	c.defaultAction = "index"
	c.useDefaultControllerAlways = true
	c.plugins = make([]string, 0)
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
