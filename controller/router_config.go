package controller

import "github.com/noxyicm/wsf/config"

// RouterConfig represents router configuration
type RouterConfig struct {
	Type              string
	File              string
	URIDelimiter      string
	URIVariable       string
	URIRegexDelimiter string
	UseDefaultRoutes  bool
}

// Populate populates Config values using given Config source
func (c *RouterConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *RouterConfig) Defaults() error {
	c.Type = "default"
	c.File = "routes.json"
	c.URIDelimiter = "/"
	c.URIVariable = ":"
	c.URIRegexDelimiter = ""
	return nil
}

// Valid validates the configuration
func (c *RouterConfig) Valid() error {
	return nil
}
