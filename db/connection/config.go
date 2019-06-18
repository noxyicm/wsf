package connection

import (
	"wsf/config"
	"wsf/db/transaction"
)

// Config represents connection configuration
type Config struct {
	Type         string
	PingTimeout  int
	QueryTimeout int
	Transaction  *transaction.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if c.Transaction == nil {
		c.Transaction = &transaction.Config{}
	}

	c.Transaction.Defaults()
	if dcfg := cfg.Get("transaction"); dcfg != nil {
		c.Transaction.Populate(dcfg)
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = "default"
	c.PingTimeout = 1
	c.QueryTimeout = 0

	c.Transaction = &transaction.Config{}
	c.Transaction.Defaults()

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
