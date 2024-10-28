package db

import (
	"github.com/noxyicm/wsf/config"
)

// Config defines set of adapter variables
type Config struct {
	Priority       int
	Adapter        string
	DefaultAdapter string
	Select         *SelectConfig
	Table          *TableConfig
	Rowset         *RowsetConfig
	Row            *RowConfig
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Select == nil {
		c.Select = &SelectConfig{}
	}

	c.Select.Defaults()
	if scfg := cfg.Get("select"); scfg != nil {
		c.Select.Populate(scfg)
	}

	if c.Table == nil {
		c.Table = &TableConfig{}
	}

	c.Table.Defaults()
	if scfg := cfg.Get("table"); scfg != nil {
		c.Table.Populate(scfg)
	}

	if c.Rowset == nil {
		c.Rowset = &RowsetConfig{}
	}

	c.Rowset.Defaults()
	if rcfg := cfg.Get("rowset"); rcfg != nil {
		c.Rowset.Populate(rcfg)
	}

	if c.Row == nil {
		c.Row = &RowConfig{}
	}

	c.Row.Defaults()
	if rcfg := cfg.Get("row"); rcfg != nil {
		c.Row.Populate(rcfg)
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Priority = 3

	c.Select = &SelectConfig{}
	c.Select.Defaults()

	c.Table = &TableConfig{}
	c.Table.Defaults()

	c.Rowset = &RowsetConfig{}
	c.Rowset.Defaults()

	c.Row = &RowConfig{}
	c.Row.Defaults()

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
