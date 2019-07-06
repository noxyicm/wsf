package table

import (
	"wsf/config"
)

// Config defines set of table variables
type Config struct {
	Type            string
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
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = ""
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
