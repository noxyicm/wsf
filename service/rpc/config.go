package rpc

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
)

const (
	// SocketTypeTCP is a tcp socket
	SocketTypeTCP = "tcp"

	// SocketTypeUNIX is a unix socket
	SocketTypeUNIX = "unix"
)

// Config defines RPC service config
type Config struct {
	Enable   bool
	Protocol string
	Host     string
	Port     string
}

// Populate must populate Config values using given Config source. Must return error if Config is not valid
func (c *Config) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults allows to init blank config with pre-defined set of default values.
func (c *Config) Defaults() error {
	c.Enable = true
	c.Protocol = SocketTypeTCP
	c.Host = "127.0.0.1"
	c.Port = "6001"

	return nil
}

// Valid returns nil if config is valid
func (c *Config) Valid() error {
	if c.Protocol != SocketTypeTCP && c.Protocol != SocketTypeUNIX {
		return errors.Errorf("Invalid socket type \"%v\"", c.Protocol)
	}

	if c.Host == "" {
		errors.New("Socket address must be set")
	}

	return nil
}

// ListenTo returns listen whole string
func (c *Config) ListenTo() string {
	return c.Protocol + "://" + c.Address()
}

// Address returns listen whole string
func (c *Config) Address() string {
	s := ""
	if c.Host != "127.0.0.1" && c.Host != "localhost" {
		s = c.Host
	}

	if c.Port != "" {
		s += ":" + c.Port
	}

	return s
}
