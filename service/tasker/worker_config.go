package tasker

import (
	"wsf/config"
)

// WorkerConfig defines Worker config
type WorkerConfig struct {
	Instances             int64
	Precicion             int64
	MaxTasks              int
	MaxConsequetiveErrors int
	MaxHandlerRetryes     int
	RetryTimeout          int
	Persistent            bool
	AutoStart             bool
	AutoRestart           bool
}

// Populate must populate Config values using given Config source. Must return error if Config is not valid
func (c *WorkerConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults allows to init blank config with pre-defined set of default values.
func (c *WorkerConfig) Defaults() error {
	c.Instances = 1
	c.Precicion = 100000
	c.MaxTasks = 0
	c.MaxConsequetiveErrors = 1
	c.MaxHandlerRetryes = 0
	c.RetryTimeout = 1000000000
	c.Persistent = false
	c.AutoStart = true
	c.AutoRestart = false

	return nil
}

// Valid returns nil if config is valid
func (c *WorkerConfig) Valid() error {
	return nil
}
