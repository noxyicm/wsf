package file

import (
	"os"
	"path/filepath"
	"strings"
)

// Config represents file configuration
type Config struct {
	Dir     string
	allowed []string
}

// Defaults sets configuration default values
func (cfg *Config) Defaults() error {
	cfg.allowed = []string{}
	return nil
}

// TmpDir returns temporary directory
func (cfg *Config) TmpDir() string {
	if cfg.Dir != "" {
		return cfg.Dir
	}

	return os.TempDir()
}

// Allowed returns true if file allowed to be uploaded
func (cfg *Config) Allowed(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	for _, v := range cfg.allowed {
		if ext == v {
			return true
		}
	}

	return false
}
