package tasker

import (
	"wsf/config"
)

// Config defines RPC service config
type Config struct {
	Enable  bool
	Workers map[string]config.Config
}

// Populate must populate Config values using given Config source. Must return error if Config is not valid
func (c *Config) Populate(cfg config.Config) error {
	if c.Workers == nil {
		c.Workers = make(map[string]config.Config)
	}

	if wcfgs := cfg.Get("workers"); wcfgs != nil {
		for _, wrkType := range wcfgs.GetKeys() {
			c.Workers[wrkType] = wcfgs.Get(wrkType)
		}

		cfg.Set("workers", c.Workers)
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults allows to init blank config with pre-defined set of default values.
func (c *Config) Defaults() error {
	c.Enable = true
	//c.Workers = make(map[string]interface{})

	return nil
}

// Valid returns nil if config is valid
func (c *Config) Valid() error {
	return nil
}
