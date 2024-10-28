package acl

import (
	"github.com/noxyicm/wsf/config"
)

// RoleRegistryConfig defines set of acl role registry variables
type RoleRegistryConfig struct {
	Type string
}

// Populate populates Config values using given Config source
func (c *RoleRegistryConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *RoleRegistryConfig) Defaults() error {
	c.Type = "default"
	return nil
}

// Valid validates the configuration
func (c *RoleRegistryConfig) Valid() error {
	return nil
}
