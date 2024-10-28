package backend

import (
	"github.com/noxyicm/wsf/config"
)

// FileConfig represents file backend cache configuration
type FileConfig struct {
	Type       string
	Dir        string
	Suffix     string
	TagsHolder string
	GC         int64
}

// Populate populates Config values using given Config source
func (c *FileConfig) Populate(cfg config.Config) error {
	if err := cfg.Unmarshal(c); err != nil {
		return err
	}

	return c.Valid()
}

// Defaults sets configuration default values
func (c *FileConfig) Defaults() error {
	c.Type = "file"
	c.Dir = config.CachePath
	c.GC = 3600
	return nil
}

// Valid validates the configuration
func (c *FileConfig) Valid() error {
	return nil
}
