package acl

import (
	"wsf/config"
)

// Config defines set of acl variables
type Config struct {
	Type     string
	Enable   bool
	Priority int
	Role     config.Config
	Resource config.Config
	Rule     config.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	c.Role = cfg.Get("role")
	c.Resource = cfg.Get("resource")
	c.Rule = cfg.Get("rule")

	if c.Role == nil {
		c.Role = config.NewBridge()
	}

	if c.Resource == nil {
		c.Resource = config.NewBridge()
	}

	if c.Rule == nil {
		c.Rule = config.NewBridge()
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = "default"
	c.Enable = true
	c.Priority = 3

	c.Role = config.NewBridge()
	c.Role.Merge(map[string]interface{}{
		"type": "default",
	})

	c.Resource = config.NewBridge()
	c.Resource.Merge(map[string]interface{}{
		"type": "default",
	})

	c.Rule = config.NewBridge()
	c.Rule.Merge(map[string]interface{}{
		"type": "default",
	})
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
