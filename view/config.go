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
	ViewActionPathSpec             string
	ViewActionPathNoControllerSpec string
	ViewHelperPathSpec             string
	ViewSuffix                     string
	SegmentContentKey              string
	DefaultLayout                  string
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
	c.ViewBasePathSpec = "views/actions/:module/"
	c.ViewActionPathSpec = "views/actions/:module/:controller/:action.:suffix"
	c.ViewActionPathNoControllerSpec = "views/actions/:module/:action.:suffix"
	c.ViewHelperPathSpec = "views/helpers/"
	c.ViewSuffix = "gohtml"
	c.SegmentContentKey = "content"
	c.DefaultLayout = "default"
	c.Assign = make(map[string]interface{})
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	return nil
}
