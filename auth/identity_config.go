package auth

import "wsf/config"

// IdentityConfig defines set of adapter variables
type IdentityConfig struct {
	Type string
}

// Populate populates Config values using given Config source
func (c *IdentityConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *IdentityConfig) Defaults() error {
	c.Type = "default"

	return nil
}

// Valid validates the configuration
func (c *IdentityConfig) Valid() error {
	return nil
}
