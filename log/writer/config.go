package writer

import (
	"github.com/noxyicm/wsf/config"
)

// Config represents writer configuration
type Config struct {
	Enable    bool
	Params    map[string]interface{}
	Formatter map[string]interface{}
	Filters   []map[string]interface{}
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(c)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Enable = true
	c.Params = map[string]interface{}{
		"type": "null",
	}

	c.Formatter = make(map[string]interface{})
	c.Filters = make([]map[string]interface{}, 0)
	return nil
}
