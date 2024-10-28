package db

import (
	"time"
	"github.com/noxyicm/wsf/config"
)

// AdapterConfig defines set of adapter variables
type AdapterConfig struct {
	Type                  string
	Username              string
	Password              string
	Protocol              string
	Host                  string
	Port                  int
	DBname                string
	Charset               string
	TimeFormat            *time.Location
	PingTimeout           int
	QueryTimeout          int
	ConnectionMaxLifeTime int
	MaxIdleConnections    int
	MaxOpenConnections    int
	AutoQuoteIdentifiers  bool
	Transaction           *TransactionConfig
	Connection            *ConnectionConfig
	Profiler              config.Config
	Logger                config.Config
}

// Populate populates Config values using given Config source
func (c *AdapterConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Transaction == nil {
		c.Transaction = &TransactionConfig{}
	}

	c.Transaction.Defaults()
	if dcfg := cfg.Get("transaction"); dcfg != nil {
		c.Transaction.Populate(dcfg)
	}

	if c.Connection == nil {
		c.Connection = &ConnectionConfig{}
	}

	c.Connection.Defaults()
	c.Connection.PingTimeout = c.PingTimeout
	c.Connection.QueryTimeout = c.QueryTimeout
	if ccfg := cfg.Get("connection"); ccfg != nil {
		c.Connection.Populate(ccfg)
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *AdapterConfig) Defaults() error {
	c.Type = ""
	c.PingTimeout = 1
	c.QueryTimeout = 5
	c.ConnectionMaxLifeTime = 10
	c.MaxIdleConnections = 100
	c.MaxOpenConnections = 100
	c.AutoQuoteIdentifiers = true

	c.Transaction = &TransactionConfig{}
	c.Transaction.Defaults()

	c.Connection = &ConnectionConfig{}
	c.Connection.Defaults()
	c.Connection.PingTimeout = c.PingTimeout
	c.Connection.QueryTimeout = c.QueryTimeout

	return nil
}

// Valid validates the configuration
func (c *AdapterConfig) Valid() error {
	return nil
}
