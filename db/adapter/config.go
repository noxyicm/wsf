package adapter

import (
	"time"
	"wsf/config"
	"wsf/db/connection"
	"wsf/db/dbselect"
	"wsf/db/table/row"
	"wsf/db/table/rowset"
	"wsf/db/transaction"
)

// Config defines set of adapter variables
type Config struct {
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
	Connection            *connection.Config
	Transaction           *transaction.Config
	Select                *dbselect.Config
	Rowset                *rowset.Config
	Row                   *row.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Transaction == nil {
		c.Transaction = &transaction.Config{}
	}

	c.Transaction.Defaults()
	if dcfg := cfg.Get("transaction"); dcfg != nil {
		c.Transaction.Populate(dcfg)
	}

	if c.Connection == nil {
		c.Connection = &connection.Config{}
	}

	c.Connection.Defaults()
	c.Connection.PingTimeout = c.PingTimeout
	c.Connection.QueryTimeout = c.QueryTimeout
	if ccfg := cfg.Get("connection"); ccfg != nil {
		c.Connection.Populate(ccfg)
	}

	if c.Select == nil {
		c.Select = &dbselect.Config{}
	}

	c.Select.Defaults()
	if scfg := cfg.Get("select"); scfg != nil {
		c.Select.Populate(scfg)
	}

	if c.Rowset == nil {
		c.Rowset = &rowset.Config{}
	}

	c.Rowset.Defaults()
	if rcfg := cfg.Get("rowset"); rcfg != nil {
		c.Rowset.Populate(rcfg)
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = ""
	c.PingTimeout = 1
	c.QueryTimeout = 0
	c.ConnectionMaxLifeTime = 0
	c.MaxIdleConnections = 100
	c.MaxOpenConnections = 100
	c.AutoQuoteIdentifiers = true

	c.Transaction = &transaction.Config{}
	c.Transaction.Defaults()

	c.Connection = &connection.Config{}
	c.Connection.Defaults()
	c.Connection.PingTimeout = c.PingTimeout
	c.Connection.QueryTimeout = c.QueryTimeout

	c.Select = &dbselect.Config{}
	c.Select.Defaults()

	c.Rowset = &rowset.Config{}
	c.Rowset.Defaults()

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
