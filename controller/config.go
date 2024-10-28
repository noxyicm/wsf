package controller

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
)

// Config represents controller configuration
type Config struct {
	Type            string
	Priority        int
	ThrowExceptions bool
	ErrorHandling   bool
	VerboseErrors   bool
	Logger          *log.Log
	Dispatcher      *DispatcherConfig
	Router          config.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if c.Dispatcher == nil {
		c.Dispatcher = &DispatcherConfig{}
	}

	c.Dispatcher.Defaults()
	if dcfg := cfg.Get("dispatcher"); dcfg != nil {
		c.Dispatcher.Populate(dcfg)
	}

	if c.Router == nil {
		c.Router = config.NewBridge()
	}

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = "default"
	c.Priority = 2
	c.ThrowExceptions = false
	c.ErrorHandling = false
	c.VerboseErrors = false

	c.Dispatcher = &DispatcherConfig{}
	c.Dispatcher.Defaults()

	c.Router = config.NewBridge()
	c.Router.Merge(map[string]interface{}{
		"type":              "default",
		"file":              "routes.json",
		"uridelimiter":      "/",
		"urivariable":       ":",
		"uriregexdelimiter": "",
		"moduleprefix":      "",
	})
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	if c.Dispatcher == nil {
		return errors.New("Invalid dispatcher configuration")
	}

	if c.Router == nil {
		return errors.New("Invalid router configuration")
	}

	return nil
}
