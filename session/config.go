package session

import (
	"github.com/noxyicm/wsf/config"
)

// Config defines set of session variables
type Config struct {
	Type string `json:"type"`
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(c)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Type = "default"

	return nil
}
