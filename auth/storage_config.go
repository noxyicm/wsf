package auth

import "wsf/config"

// StorageConfig defines set of adapter variables
type StorageConfig struct {
	Type      string
	Namespace string
	Member    string
}

// Populate populates Config values using given Config source
func (c *StorageConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *StorageConfig) Defaults() error {
	c.Type = "default"
	c.Namespace = "WSFAuth"
	c.Member = "storage"

	return nil
}

// Valid validates the configuration
func (c *StorageConfig) Valid() error {
	return nil
}
