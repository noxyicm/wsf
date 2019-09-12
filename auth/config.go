package auth

import "wsf/config"

// Config defines set of adapter variables
type Config struct {
	Priority int
	Type     string
	Adapter  *AdapterConfig
	Storage  *StorageConfig
	Identity *IdentityConfig
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	if c.Adapter == nil {
		c.Adapter = &AdapterConfig{}
	}

	c.Adapter.Defaults()
	if acfg := cfg.Get("adapter"); acfg != nil {
		c.Adapter.Populate(acfg)
	}

	if c.Storage == nil {
		c.Storage = &StorageConfig{}
	}

	c.Storage.Defaults()
	if scfg := cfg.Get("storage"); scfg != nil {
		c.Storage.Populate(scfg)
	}

	if c.Identity == nil {
		c.Identity = &IdentityConfig{}
	}

	c.Identity.Defaults()
	if icfg := cfg.Get("identity"); icfg != nil {
		c.Identity.Populate(icfg)
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Priority = 11

	c.Adapter = &AdapterConfig{}
	c.Adapter.Defaults()

	c.Storage = &StorageConfig{}
	c.Storage.Defaults()

	c.Identity = &IdentityConfig{}
	c.Identity.Defaults()

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
