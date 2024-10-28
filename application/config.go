package application

import (
	"os"
	"path/filepath"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/log"
)

// Config defines set of environment variables
type Config struct {
	Environment string
	RootPath    string
	BasePath    string
	AppPath     string
	CachePath   string
	StaticPath  string
	DesignPath  string

	Offline      bool
	Domain       string
	Host         string
	Prefix       string
	Multitanency bool
	Instance     int
	Hub          bool
	Name         string
	Lang         string

	Log *log.Config
}

// Populate populates Config values using given Config source
func (c *Config) Populate(cfg config.Config) error {
	return cfg.Unmarshal(c)
}

// Defaults sets configuration default values
func (c *Config) Defaults() error {
	c.Environment = EnvDEV
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	c.RootPath = filepath.FromSlash(dir + "/")
	c.BasePath = filepath.FromSlash(dir + "/")
	c.AppPath = filepath.FromSlash(dir + "/application/")
	c.CachePath = filepath.FromSlash(dir + "/cache/")
	c.StaticPath = filepath.FromSlash(dir + "/public/")
	c.DesignPath = filepath.FromSlash(dir + "/public/design/")
	return nil
}
