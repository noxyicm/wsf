package file

import (
	"os"
	"path/filepath"
	"strings"
	"wsf/config"
	"wsf/utils"
)

// Config represents file configuration
type Config struct {
	Dir               string
	AllowedExtensions []string
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Dir = ""
	c.AllowedExtensions = []string{}
	return nil
}

// TmpDir returns temporary directory
func (c *Config) TmpDir() string {
	if c.Dir != "" {
		return filepath.Join(config.StaticPath, c.Dir)
	}

	return os.TempDir()
}

// Allowed returns true if file allowed to be uploaded
func (c *Config) Allowed(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return utils.InSSlice(ext[1:], c.AllowedExtensions)
}
