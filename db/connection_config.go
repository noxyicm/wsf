package db

import (
	"wsf/config"
)

// ConnectionConfig represents connection configuration
type ConnectionConfig struct {
	Type         string
	PingTimeout  int
	QueryTimeout int
	Transaction  *TransactionConfig
}

// Populate populates Config values using given Config source
func (c *ConnectionConfig) Populate(cfg config.Config) error {
	if c.Transaction == nil {
		c.Transaction = &TransactionConfig{}
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
func (c *ConnectionConfig) Defaults() error {
	c.Type = "default"
	c.PingTimeout = 1
	c.QueryTimeout = 0

	c.Transaction = &TransactionConfig{}
	c.Transaction.Defaults()

	return nil
}

// Valid validates the configuration
func (c *ConnectionConfig) Valid() error {
	return nil
}
