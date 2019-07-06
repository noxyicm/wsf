package cache

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"time"
	"wsf/cache/backend"
	"wsf/config"
	"wsf/errors"
	"wsf/log"
	"wsf/utils"
)

// Public constants
const (
	CleaningModeAll = iota
	CleaningModeOld
	CleaningModeMatchingTag
	CleaningModeNotMatchingTag
	CleaningModeMatchingAnyTag
)

var (
	allowedSymbolsForIdsAndTags = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// Interface represents a core cache
type Interface interface {
	Load(id string, testCacheValidity bool) ([]byte, bool)
	Read(id string, object *interface{}, testCacheValidity bool) bool
	Test(id string) bool
	Save(data []byte, id string, tags []string, specificLifetime int64) bool
	Remove(id string) bool
	Clear(mode int64, tags []string) bool
	Error() error
}

// Core cache
type Core struct {
	options         *Config
	logger          *log.Log
	backend         backend.Interface
	extendedBackend bool
	lastError       error
	lastID          string
}

// Priority returns resource initialization priority
func (c *Core) Priority() int {
	return c.options.Priority
}

// Load loads data from cache
func (c *Core) Load(id string, testCacheValidity bool) ([]byte, bool) {
	if !c.options.Enable {
		return nil, false
	}

	id = c.prepareID(id)
	c.lastID = id
	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return nil, false
	}

	c.logger.Debugf("[WSF Cache]: Load item '%s'", nil, id)
	data, err := c.backend.Load(id, testCacheValidity)
	if err != nil {
		c.lastError = err
		return nil, false
	}

	if len(data) == 0 {
		return nil, false
	}

	return data, true
}

// Read loads data from cache and unmarshals it into object
func (c *Core) Read(id string, object *interface{}, testCacheValidity bool) bool {
	if !c.options.Enable {
		return false
	}

	id = c.prepareID(id)
	c.lastID = id
	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return false
	}

	c.logger.Debugf("[WSF Cache]: Load item '%s'", nil, id)
	data, err := c.backend.Load(id, testCacheValidity)
	if err != nil {
		c.lastError = err
		return false
	}

	if len(data) == 0 {
		return false
	}

	if err := json.Unmarshal(data, object); err != nil {
		c.lastError = errors.Wrap(err, "Unable to deserialize data")
		return false
	}

	return true
}

// Test returns true if chache exists
func (c *Core) Test(id string) bool {
	if !c.options.Enable {
		return false
	}

	id = c.prepareID(id)
	c.lastID = id
	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return false
	}

	c.logger.Debugf("[WSF Cache]: Load item '%s'", nil, id)
	return c.backend.Test(id)
}

// Save saves data into cache
func (c *Core) Save(data []byte, id string, tags []string, specificLifetime int64) bool {
	if !c.options.Enable {
		return false
	}

	if id == "" {
		id = c.lastID
	} else {
		id = c.prepareID(id)
	}

	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return false
	}

	if err := c.validateTags(tags); err != nil {
		c.lastError = err
		return false
	}

	// automatic cleaning
	if c.options.AutomaticCleaningFactor > 0 {
		rand.Seed(time.Now().UnixNano())
		rand := rand.Int63n(c.options.AutomaticCleaningFactor)
		if rand == 0 {
			if c.options.ExtendedBackend {
				c.logger.Debug("[WSF Cache]::save(): Automatic cleaning running", nil)
				c.Clear(CleaningModeOld, []string{})
			} else {
				c.logger.Warning("[WSF Cache]::save(): Automatic cleaning is not available/necessary with current backend", nil)
			}
		}
	}

	c.logger.Debugf("[WSF Cache]: Save item '%s'", nil, id)
	if err := c.backend.Save(data, id, tags, specificLifetime); err != nil {
		c.logger.Warningf("[WSF Cache]::save(): Failed to save item '%s' -> removing it", nil, id)
		c.Remove(id)
		c.lastError = err
		return false
	}

	if c.options.WriteControl {
		dataCheck, err := c.backend.Load(id, true)
		if err != nil {
			c.logger.Warningf("[WSF Cache]::save(): Write control of item '%s' failed -> removing it", nil, id)
			c.Remove(id)
			c.lastError = err
			return false
		}

		if !utils.EqualBSlice(data, dataCheck) {
			c.logger.Warningf("[WSF Cache]::save(): Write control of item '%s' failed -> removing it", nil, id)
			c.Remove(id)
			return false
		}
	}

	return true
}

// Remove the cache
func (c *Core) Remove(id string) bool {
	if !c.options.Enable {
		return false
	}

	id = c.prepareID(id)
	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return false
	}

	c.logger.Debugf("[WSF Cache]: Remove item '%s'", nil, id)
	if err := c.backend.Remove(id); err != nil {
		c.lastError = err
		return false
	}

	return true
}

// Clear cache
func (c *Core) Clear(mode int64, tags []string) bool {
	if !c.options.Enable {
		return false
	}

	if !utils.InI64Slice(mode, []int64{CleaningModeAll, CleaningModeOld, CleaningModeMatchingTag, CleaningModeNotMatchingTag, CleaningModeMatchingAnyTag}) {
		c.lastError = errors.New("Invalid cleaning mode")
		return false
	}

	if err := c.validateTags(tags); err != nil {
		c.lastError = err
		return false
	}

	if err := c.backend.Clear(mode, tags); err != nil {
		c.lastError = err
		return false
	}

	return true
}

// Error returns the last accuired error
func (c *Core) Error() error {
	return c.lastError
}

func (c *Core) prepareID(id string) string {
	if id != "" && c.options.CacheIDPrefix != "" {
		return c.options.CacheIDPrefix + id
	}

	return id
}

func (c *Core) validateIDOrTag(s string) error {
	if string(s[0:9]) == "internal-" {
		return errors.New("'internal-*' ids or tags are reserved")
	}

	if !allowedSymbolsForIdsAndTags.MatchString(s) {
		return errors.Errorf("Invalid id or tag '%s': must use only [a-zA-Z0-9_]", s)
	}

	return nil
}

func (c *Core) validateTags(tags []string) error {
	for _, tag := range tags {
		if err := c.validateIDOrTag(tag); err != nil {
			return err
		}
	}

	return nil
}

// NewCore creates a new cache adapter specified by type
func NewCore(cacheType string, options config.Config) (*Core, error) {
	cfg := &Config{}
	cfg.Defaults()
	cfg.Populate(options)

	cc := &Core{
		options: cfg,
	}

	adp, err := backend.NewBackendCache(cfg.Backend.GetString("type"), cfg.Backend)
	if err != nil {
		return nil, err
	}
	cc.backend = adp

	lg, err := log.NewLog(cfg.Logger)
	if err != nil {
		return nil, err
	}
	cc.logger = lg

	return cc, nil
}
