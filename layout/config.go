package layout

import (
	"wsf/config"
)

// Config represents layout configuration
type Config struct {
	Type            string
	Priority        int
	Enabled         bool
	ContentKey      string
	InflectorTarget string
	Layout          string
	ViewScriptPath  string
	ViewBasePath    string
	ViewSuffix      string
	HelperName      string
	PluginName      string
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
	c.Type = "default"
	c.Priority = 21
	c.Enabled = true
	c.ContentKey = "content"
	c.InflectorTarget = ":script.:suffix"
	c.Layout = "layout"
	c.ViewScriptPath = "layouts/"
	c.ViewSuffix = "gohtml"
	c.HelperName = TYPELayoutActionHelper
	c.PluginName = TYPELayoutPlugin

	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
