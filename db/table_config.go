package db

import (
	"wsf/config"
)

// TableConfig defines set of table variables
type TableConfig struct {
	Type            string
	Adapter         string
	Definition      map[string]interface{}
	DefinitionName  string
	Primary         []string
	Identity        int64
	Schema          string
	Name            string
	ReferenceMap    map[string]interface{}
	DependentTables map[string]interface{}
	DefaultSource   string
	DefaultValues   map[string]interface{}
	RowsetType      string
	RowType         string
}

// Populate populates Config values using given Config source
func (c *TableConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *TableConfig) Defaults() error {
	c.Type = ""
	c.DefaultSource = DefaultNone
	c.Definition = make(map[string]interface{})
	c.ReferenceMap = make(map[string]interface{})
	c.DependentTables = make(map[string]interface{})
	c.DefaultValues = make(map[string]interface{})
	return nil
}

// Valid validates the configuration
func (c *TableConfig) Valid() error {
	return nil
}
