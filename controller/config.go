package controller

import (
	"wsf/config"
	"wsf/controller/dispatcher"
	"wsf/controller/router"
	"wsf/log"

	"github.com/pkg/errors"
)

// Config represents controller configuration
type Config struct {
	Type            string
	Priority        int
	ThrowExceptions bool
	ErrorHandling   bool
	Logger          *log.Log
	Dispatcher      *dispatcher.Config
	Router          *router.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if c.Dispatcher == nil {
		c.Dispatcher = &dispatcher.Config{}
	}

	c.Dispatcher.Defaults()
	if dcfg := cfg.Get("dispatcher"); dcfg != nil {
		c.Dispatcher.Populate(dcfg)
	}

	if c.Router == nil {
		c.Router = &router.Config{}
	}

	c.Router.Defaults()
	if rcfg := cfg.Get("router"); rcfg != nil {
		c.Router.Populate(rcfg)
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
	c.ThrowExceptions = true
	c.ErrorHandling = true

	c.Dispatcher = &dispatcher.Config{}
	c.Dispatcher.Defaults()

	c.Router = &router.Config{}
	c.Router.Defaults()
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
