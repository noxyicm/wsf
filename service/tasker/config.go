package tasker

import (
	"wsf/config"
)

// Config defines RPC service config
type Config struct {
	Enable             bool
	Persistent         bool
	TryStartNewWorkers bool
	Workers            map[string]config.Config
	Tasks              []config.Config
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

	if c.Tasks == nil {
		c.Tasks = make([]config.Config, 0)
	}

	if tcfgs := cfg.Get("tasks"); tcfgs != nil {
		for _, key := range tcfgs.GetKeys() {
			c.Tasks = append(c.Tasks, tcfgs.Get(key))
		}

		cfg.Set("tasks", c.Tasks)
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults allows to init blank config with pre-defined set of default values.
func (c *Config) Defaults() error {
	c.Enable = true
	c.Persistent = true
	c.TryStartNewWorkers = true
	c.Workers = make(map[string]config.Config)
	c.Tasks = make([]config.Config, 0)

	return nil
}

// Valid returns nil if config is valid
func (c *Config) Valid() error {
	return nil
}
