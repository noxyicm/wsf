package auth

import "github.com/noxyicm/wsf/config"

// AdapterConfig defines set of adapter variables
type AdapterConfig struct {
	Type                string
	Source              string
	IdentityColumn      string
	CredentialColumn    string
	CredentialTreatment string
	AmbiguityIdentity   bool
}

// Populate populates Config values using given Config source
func (c *AdapterConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *AdapterConfig) Defaults() error {
	c.Type = "default"

	return nil
}

// Valid validates the configuration
func (c *AdapterConfig) Valid() error {
	return nil
}
