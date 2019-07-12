package db

import (
	"wsf/config"
)

// RowsetConfig defines set of rowset variables
type RowsetConfig struct {
	Type      string
	Tbl       string
	Connected bool
	Pointer   uint32
	Cnt       uint32
	Pointing  bool
	Stored    bool
	ReadOnly  bool
	Row       *RowConfig
}

// Populate populates Config values using given Config source
func (c *RowsetConfig) Populate(cfg config.Config) error {
	if c.Row == nil {
		c.Row = &RowConfig{}
	}

	c.Row.Defaults()
	if rcfg := cfg.Get("row"); rcfg != nil {
		c.Row.Populate(rcfg)
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *RowsetConfig) Defaults() error {
	c.Type = "default"

	c.Row = &RowConfig{}
	c.Row.Defaults()

	return nil
}

// Valid validates the configuration
func (c *RowsetConfig) Valid() error {
	return nil
}
