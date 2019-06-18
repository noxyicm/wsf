package static

import (
	"os"
	"path"
	"strings"
	"wsf/config"

	"github.com/pkg/errors"
)

// Config defines Static server configuration
type Config struct {
	Dir    string
	Forbid []string
	Always []string
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
	c.Dir = config.StaticPath
	c.Forbid = make([]string, 0)
	c.Always = make([]string, 0)
	return nil
}

// Valid returns nil if config is valid
func (c *Config) Valid() error {
	st, err := os.Stat(c.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("Root directory '%s' does not exists", c.Dir)
		}

		return err
	}

	if !st.IsDir() {
		return errors.Errorf("Invalid root directory '%s'", c.Dir)
	}

	return nil
}

// AlwaysForbid must return true if file extension is not allowed for the upload
func (c *Config) AlwaysForbid(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))

	for _, v := range c.Forbid {
		if ext == v {
			return true
		}
	}

	return false
}

// AlwaysServe must indicate that file is expected to be served by static service
func (c *Config) AlwaysServe(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))
	if ext != "" {
		ext = ext[1:len(ext)]
	}

	for _, v := range c.Always {
		if ext == v {
			return true
		}
	}

	return false
}
