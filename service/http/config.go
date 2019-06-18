package http

import (
	"fmt"
	"os"
	"strconv"
	"wsf/application/file"
	"wsf/config"
	"wsf/errors"
)

// Config defines HTTP server configuration
type Config struct {
	Enable             bool
	Proxy              bool
	Host               string
	Port               int
	SSL                SSLConfig
	MaxRequestSize     int64
	MaxRequestTimeout  int64
	MaxResponseTimeout int64
	Uploads            *file.Config
	AccessLogger       config.Config
	Headers            map[string]string
}

// EnableTLS returns true if server must listen TLS connections
func (c *Config) EnableTLS() bool {
	return c.SSL.Key != "" || c.SSL.Cert != ""
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	if lcfg := cfg.Get("accesslog"); lcfg != nil {
		c.AccessLogger = lcfg
	}

	if c.AccessLogger == nil {
		c.AccessLogger = config.NewBridge()
	}

	if c.Uploads == nil {
		c.Uploads = &file.Config{}
	}

	if c.SSL.Port == 0 {
		c.SSL.Port = 443
	}

	c.Uploads.Defaults()

	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Enable = true
	c.Host = "127.0.0.1"
	c.Port = 8080
	c.MaxRequestSize = 1 << 26
	c.Headers = make(map[string]string)

	if c.AccessLogger == nil {
		c.AccessLogger = config.NewBridge()
		c.AccessLogger.Merge(map[string]interface{}{
			"enable": true,
			"writers": map[string]interface{}{
				"file": map[string]interface{}{
					"params": map[string]interface{}{
						"type":   "stream",
						"stream": "logs/access.log",
					},
					"formatter": map[string]interface{}{
						"type": "httpaccess",
					},
				},
			},
		})
	}
	return nil
}

// Valid validates the configuration
func (c *Config) Valid() error {
	if c.Uploads == nil {
		return errors.New("Invalid uploads configuration")
	}

	if c.EnableTLS() {
		if _, err := os.Stat(c.SSL.Key); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("SSL Certificate .key file '%s' does not exists", c.SSL.Key)
			}

			return err
		}

		if _, err := os.Stat(c.SSL.Cert); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("SSL Certificate .cert file '%s' does not exists", c.SSL.Cert)
			}

			return err
		}
	}

	return nil
}

// Address returns full address string
func (c *Config) Address() string {
	s := c.Host
	if c.EnableTLS() {
		s = s + ":" + strconv.Itoa(c.SSL.Port)
	} else if c.Port != 0 {
		s += ":" + strconv.Itoa(c.Port)
	}

	return s
}

// SSLConfig defines HTTPS server configuration
type SSLConfig struct {
	Port     int
	Redirect bool
	Key      string
	Cert     string
}
