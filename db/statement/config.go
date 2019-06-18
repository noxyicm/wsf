package statement

import "wsf/config"

// Config defines set of statement variables
type Config struct {
	Params map[string]interface{}
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(c)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	return nil
}

// SupportsParameters returns true if adapter supports
func (c *Config) SupportsParameters(param string) bool {
	if v, ok := c.Params[param]; ok {
		return v.(bool)
	}

	return false
}
