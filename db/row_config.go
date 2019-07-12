package db

import "wsf/config"

// RowConfig defines set of row variables
type RowConfig struct {
	Type      string
	Table     string
	Data      map[string]interface{}
	Connected bool
	Stored    bool
	ReadOnly  bool
}

// Populate populates Config values using given Config source
func (c *RowConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *RowConfig) Defaults() error {
	c.Type = "default"
	c.Data = make(map[string]interface{})
	return nil
}

// Valid validates the configuration
func (c *RowConfig) Valid() error {
	return nil
}
