package log

import (
	"time"
	"wsf/config"
	"wsf/log/writer"
)

// Config represents logger configuration
type Config struct {
	Priority        int
	Enable          bool
	Verbose         bool
	TimestampFormat string
	Writers         map[string]*writer.Config
	Filters         map[string]map[string]interface{}
	Extras          map[string]string
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if wscfg := cfg.Get("writers"); wscfg != nil {
		for _, k := range wscfg.GetKeys() {
			writerCfg := &writer.Config{}
			writerCfg.Defaults()
			writerCfg.Populate(wscfg.Get(k))

			c.Writers[k] = writerCfg
		}
	} else {
		c.Writers["default"] = &writer.Config{}
		c.Writers["default"].Defaults()
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Priority = 1
	c.Enable = true
	c.Verbose = false
	c.TimestampFormat = time.RFC3339

	c.Writers = make(map[string]*writer.Config)
	c.Filters = make(map[string]map[string]interface{}, 0)
	c.Extras = make(map[string]string)
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
