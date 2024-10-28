package controller

import "wsf/config"

// RouteConfig represents router configuration
type RouteConfig struct {
	Type              string
	URIDelimiter      string
	URIVariable       string
	URIRegexDelimiter string
	ModulePrefix      string
	Path              string
	Module            string
	Controller        string
	Action            string
	Default           map[string]interface{}
	Locale            string
}

// Populate populates Config values using given Config source
func (c *RouteConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *RouteConfig) Defaults() error {
	c.Type = "default"
	c.Default = make(map[string]interface{})
	c.URIDelimiter = "/"
	c.URIVariable = ":"
	c.URIRegexDelimiter = ""
	c.ModulePrefix = ""
	return nil
}

// Valid validates the configuration
func (c *RouteConfig) Valid() error {
	return nil
}
