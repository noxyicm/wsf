package view

import (
	"github.com/noxyicm/wsf/config"
)

// Config represents view configuration
type Config struct {
	Type                           string
	Priority                       int
	BaseDir                        string
	ViewBasePathSpec               string
	ViewScriptPathSpec             string
	ViewScriptPathNoControllerSpec string
	ViewSuffix                     string
	LayoutContentKey               string
	Doctype                        string
	Charset                        string
	ContentType                    string
	Assign                         map[string]interface{}
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
	c.Priority = 5
	c.BaseDir = ""
	c.ViewBasePathSpec = "views/:module/"
	c.ViewScriptPathSpec = "views/:module/:controller/:action.:suffix"
	c.ViewScriptPathNoControllerSpec = "views/:module/:action.:suffix"
	c.ViewSuffix = "gohtml"
	c.LayoutContentKey = "content"
	c.Assign = make(map[string]interface{})
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
