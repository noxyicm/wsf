package rowset

import (
	"wsf/config"
	"wsf/db/table/row"
)

// Config defines set of statement variables
type Config struct {
	Type string
	Row  *row.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if c.Row == nil {
		c.Row = &row.Config{}
	}

	c.Row.Defaults()
	if rcfg := cfg.Get("row"); rcfg != nil {
		c.Row.Populate(rcfg)
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = "default"

	c.Row = &row.Config{}
	c.Row.Defaults()

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
