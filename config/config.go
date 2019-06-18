package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"wsf/errors"

	"github.com/spf13/viper"
)

var (
	// AppRootPath holds application root folder
	AppRootPath = "/"

	// AppPath holds application folder
	AppPath = "application/"

	// BasePath is the absolute path to the app
	BasePath = "/"

	// StaticPath is the work path to the app
	StaticPath = "public/"

	// CachePath is the cashe path to the app
	CachePath = "cache/"

	// AppEnv represents application environment
	AppEnv string

	// App is a general application config
	App Config
)

// Config is general config interface
type Config interface {
	Get(name string) Config
	GetString(name string) string
	GetStringDefault(name string, def string) string
	GetBool(key string) bool
	GetBoolDefault(key string, def bool) bool
	GetKeys() []string
	GetAll() map[string]interface{}
	Merge(map[string]interface{}) error
	Unmarshal(out interface{}) error
}

// PopulatableConfig is an interface for populationg config with another config
type PopulatableConfig interface {
	Populate(Config) error
}

// DefaultConfig is an interface for initializing default config values
type DefaultConfig interface {
	Defaults() error
}

// Bridge provides interface bridge between viper configs and config.Config
type Bridge struct {
	v *viper.Viper
}

// Set sets value in config
func (c *Bridge) Set(key string, value interface{}) error {
	c.v.Set(key, value)
	return nil
}

// Get nested config section (sub-map), returns nil if section not found
func (c *Bridge) Get(key string) Config {
	sub := c.v.Sub(key)
	if sub == nil {
		return nil
	}

	return &Bridge{sub}
}

// GetString returns a string value
func (c *Bridge) GetString(key string) string {
	return c.v.GetString(key)
}

// GetStringDefault returns a string value or default value if empty
func (c *Bridge) GetStringDefault(key string, def string) string {
	str := c.v.Get(key)
	if str == nil {
		return def
	}

	return str.(string)
}

// GetBool returns a boolean value
func (c *Bridge) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetBoolDefault returns a boolean value or default value if empty
func (c *Bridge) GetBoolDefault(key string, def bool) bool {
	b := c.v.Get(key)
	if b == nil {
		return def
	}

	return b.(bool)
}

// GetKeys returns config keys
func (c *Bridge) GetKeys() []string {
	settings := c.v.AllSettings()
	s := make([]string, len(settings))
	i := 0
	for key := range settings {
		s[i] = key
		i++
	}

	return s
}

// GetAll returns a map
func (c *Bridge) GetAll() map[string]interface{} {
	return c.v.AllSettings()
}

// Merge merges a new configuration with an existing config
func (c *Bridge) Merge(cfg map[string]interface{}) error {
	return c.v.MergeConfigMap(cfg)
}

// Unmarshal unmarshals config data into given struct
func (c *Bridge) Unmarshal(out interface{}) error {
	return c.v.Unmarshal(out)
}

// NewBridge creates new bridge
func NewBridge() *Bridge {
	cfg := viper.New()
	return &Bridge{cfg}
}

// LoadConfig loads config file and merge it's values with set of flags
func LoadConfig(file string, path []string, name string, flags []string) (*Bridge, error) {
	cfg := viper.New()

	if file != "" {
		if absPath, err := filepath.Abs(file); err == nil {
			file = absPath

			if _, err := os.Stat(file); err != nil {
				return nil, err
			}
		}

		cfg.SetConfigFile(file)

		if dir, err := filepath.Abs(file); err == nil {
			if _, err := os.Stat(filepath.Dir(dir)); err != nil {
				return nil, err
			}
		}
	} else {
		for _, p := range path {
			cfg.AddConfigPath(p)
		}

		cfg.SetConfigName(name)
	}

	cfg.AutomaticEnv()
	if err := cfg.ReadInConfig(); err != nil {
		if len(flags) == 0 {
			err = errors.Wrap(err, "Read in config faild")
			return nil, err
		}
	}

	dcfg := defaults()
	if err := dcfg.MergeConfigMap(cfg.AllSettings()); err != nil {
		return nil, err
	}

	if len(flags) != 0 {
		for _, f := range flags {
			k, v, err := parseFlag(f)
			if err != nil {
				return nil, err
			}

			dcfg.Set(k, v)
		}

		merged := viper.New()
		if err := merged.MergeConfigMap(dcfg.AllSettings()); err != nil {
			return nil, err
		}

		return &Bridge{merged}, nil
	}

	return &Bridge{dcfg}, nil
}

func defaults() *viper.Viper {
	cfg := viper.New()

	dcfg := map[string]interface{}{
		"resources": map[string]interface{}{
			"log": map[string]interface{}{
				"log": map[string]interface{}{
					"enable": true,
					"writers": map[string]interface{}{
						"default": map[string]interface{}{
							"params": map[string]interface{}{
								"type": "null",
							},
							"formatter": map[string]interface{}{
								"type": "simple",
							},
						},
					},
				},
			},
			"controller": map[string]interface{}{
				"maincontroller": map[string]interface{}{
					"type":            "default",
					"throwExceptions": true,
					"errorHandling":   true,
				},
			},
			"view": map[string]interface{}{
				"view": map[string]interface{}{
					"type":        "default",
					"doctype":     "html5",
					"charset":     "utf-8",
					"contentType": "text/html",
				},
			},
		},
	}

	cfg.MergeConfigMap(dcfg)
	return cfg
}

func parseFlag(flag string) (string, string, error) {
	if !strings.Contains(flag, "=") {
		return "", "", errors.Errorf("invalid flag `%s`", flag)
	}

	parts := strings.SplitN(strings.TrimLeft(flag, " \"'`"), "=", 2)

	return strings.Trim(parts[0], " \n\t"), parseValue(strings.Trim(parts[1], " \n\t")), nil
}

func parseValue(value string) string {
	escape := []rune(value)[0]

	if escape == '"' || escape == '\'' || escape == '`' {
		value = strings.Trim(value, string(escape))
		value = strings.Replace(value, fmt.Sprintf("\\%s", string(escape)), string(escape), -1)
	}

	return value
}
