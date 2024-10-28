package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"wsf/errors"

	"github.com/spf13/viper"
)

var (
	// Verbose defines if
	Verbose = false

	// AppName holds application name
	AppName = ""

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

	defaults map[string]interface{}
)

// Config is general config interface
type Config interface {
	Get(name string) Config
	GetInt(name string) int
	GetIntDefault(name string, def int) int
	GetInt64(name string) int64
	GetInt64Default(name string, def int64) int64
	GetString(name string) string
	GetStringDefault(name string, def string) string
	GetBool(key string) bool
	GetBoolDefault(key string, def bool) bool
	GetTime(name string) time.Time
	GetTimeDefault(name string, def time.Time) time.Time
	GetStringMap(key string) map[string]interface{}
	GetStringSlice(key string) []string
	GetKeys() []string
	GetAll() map[string]interface{}
	Set(key string, value interface{}) error
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

// GetInt returns an int value
func (c *Bridge) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetIntDefault returns a n int value or default value if empty
func (c *Bridge) GetIntDefault(key string, def int) int {
	if !c.v.IsSet(key) {
		return def
	}

	return c.v.GetInt(key)
}

// GetInt64 returns an int64 value
func (c *Bridge) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

// GetInt64Default returns a n int value or default value if empty
func (c *Bridge) GetInt64Default(key string, def int64) int64 {
	if !c.v.IsSet(key) {
		return def
	}

	return c.v.GetInt64(key)
}

// GetString returns a string value
func (c *Bridge) GetString(key string) string {
	return c.v.GetString(key)
}

// GetStringDefault returns a string value or default value if empty
func (c *Bridge) GetStringDefault(key string, def string) string {
	if !c.v.IsSet(key) {
		return def
	}

	return c.v.GetString(key)
}

// GetBool returns a boolean value
func (c *Bridge) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetBoolDefault returns a boolean value or default value if empty
func (c *Bridge) GetBoolDefault(key string, def bool) bool {
	if !c.v.IsSet(key) {
		return def
	}

	return c.v.GetBool(key)
}

// GetTime returns a time.Time value
func (c *Bridge) GetTime(key string) time.Time {
	return c.v.GetTime(key)
}

// GetTimeDefault returns a time.Time value or default value if empty
func (c *Bridge) GetTimeDefault(key string, def time.Time) time.Time {
	if !c.v.IsSet(key) {
		return def
	}

	return c.v.GetTime(key)
}

// GetStringMap returns a map[string]interface{} value
func (c *Bridge) GetStringMap(key string) map[string]interface{} {
	return c.v.GetStringMap(key)
}

// GetStringSlice returns a []string value
func (c *Bridge) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
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

// NewDefaultBridge creates new bridge with defaults
func NewDefaultBridge() (*Bridge, error) {
	cfg := viper.New()
	if err := cfg.MergeConfigMap(defaults); err != nil {
		return nil, errors.Wrap(err, "Unable to create default bridge")
	}

	return &Bridge{cfg}, nil
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

	ext := filepath.Ext(file)
	if ext[1:] == "ini" {
		cfg.SetConfigType("properties")
	}

	cfg.AutomaticEnv()
	if err := cfg.ReadInConfig(); err != nil {
		if len(flags) == 0 {
			err = errors.Wrap(err, "Read in config faild")
			return nil, err
		}
	}

	dcfg := getDefaults()
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

// SetDefaults sets default config for bridge
func SetDefaults(def map[string]interface{}) {
	defaults = def
}

func getDefaults() *viper.Viper {
	cfg := viper.New()
	cfg.MergeConfigMap(defaults)
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
