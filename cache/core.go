package cache

import (
	"encoding/json"
	"math/rand"
	"regexp"
	"sync"
	"time"
	"wsf/cache/backend"
	"wsf/config"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
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
	Init(options *Config) (bool, error)
	Enabled() bool
	Load(id string, testCacheValidity bool) ([]byte, bool)
	Read(id string, object interface{}, testCacheValidity bool) bool
	Test(id string) bool
	Save(data []byte, id string, tags []string, specificLifetime int64) bool
	Remove(id string) bool
	Clear(mode int64, tags []string) bool
	Error() error
}

// Core cache
type Core struct {
	Options         *Config
	Logger          *log.Log
	Backend         backend.Interface
	ExtendedBackend bool
	lastError       error
	lastID          string
	mur             sync.RWMutex
}

// Priority returns resource initialization priority
func (c *Core) Priority() int {
	return c.Options.Priority
}

// Init resource
func (c *Core) Init(options *Config) (bool, error) {
	c.Options = options

	if options.Logger == nil {
		lg := registry.GetResource("syslog")
		if lg == nil {
			return false, errors.New("Log resource is not configured")
		}
		c.Logger = lg.(*log.Log)
	}

	return c.Backend.Init(c.Options.Backend)
}

// Enabled returns true if cache is enabled
func (c *Core) Enabled() bool {
	return c.Options.Enable
}

// Load loads data from cache
func (c *Core) Load(id string, testCacheValidity bool) ([]byte, bool) {
	if !c.Options.Enable {
		return nil, false
	}

	id = c.prepareID(id)

	c.mur.Lock()
	c.lastID = id
	c.mur.Unlock()

	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return nil, false
	}

	c.Logger.Debugf("[WSF Cache]: Load item '%s'", nil, id)
	data, err := c.Backend.Load(id, testCacheValidity)
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
func (c *Core) Read(id string, object interface{}, testCacheValidity bool) bool {
	if !c.Options.Enable {
		return false
	}

	id = c.prepareID(id)

	c.mur.Lock()
	c.lastID = id
	c.mur.Unlock()

	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return false
	}

	c.Logger.Debugf("[WSF Cache]: Load item '%s'", nil, id)
	data, err := c.Backend.Load(id, testCacheValidity)
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
	if !c.Options.Enable {
		return false
	}

	id = c.prepareID(id)

	c.mur.Lock()
	c.lastID = id
	c.mur.Unlock()

	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return false
	}

	c.Logger.Debugf("[WSF Cache]: Test item '%s'", nil, id)
	return c.Backend.Test(id)
}

// Save saves data into cache
func (c *Core) Save(data []byte, id string, tags []string, specificLifetime int64) bool {
	if !c.Options.Enable {
		return false
	}

	if id == "" {
		c.mur.RLock()
		id = c.lastID
		c.mur.RUnlock()
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
	if c.Options.AutomaticCleaningFactor > 0 {
		rand.Seed(time.Now().UnixNano())
		rand := rand.Int63n(c.Options.AutomaticCleaningFactor)
		if rand == 0 {
			if c.Options.ExtendedBackend {
				c.Logger.Debug("[WSF Cache]::save(): Automatic cleaning running", nil)
				c.Clear(CleaningModeOld, []string{})
			} else {
				c.Logger.Warning("[WSF Cache]::save(): Automatic cleaning is not available/necessary with current backend", nil)
			}
		}
	}

	if specificLifetime == 0 {
		specificLifetime = int64(c.Options.Backend.GetInt("lifetime"))
	}

	c.Logger.Debugf("[WSF Cache]: Save item '%s'", nil, id)
	if err := c.Backend.Save(data, id, tags, specificLifetime); err != nil {
		c.Logger.Warningf("[WSF Cache]::save(): Failed to save item '%s' -> removing it", nil, id)
		c.Remove(id)
		c.lastError = err
		return false
	}

	if c.Options.WriteControl {
		dataCheck, err := c.Backend.Load(id, true)
		if err != nil {
			c.Logger.Warningf("[WSF Cache]::save(): Write control of item '%s' failed -> removing it", nil, id)
			c.Remove(id)
			c.lastError = err
			return false
		}

		if !utils.EqualBSlice(data, dataCheck) {
			c.Logger.Warningf("[WSF Cache]::save(): Write control of item '%s' failed -> removing it", nil, id)
			c.Remove(id)
			return false
		}
	}

	return true
}

// Remove the cache
func (c *Core) Remove(id string) bool {
	if !c.Options.Enable {
		return false
	}

	id = c.prepareID(id)
	if err := c.validateIDOrTag(id); err != nil {
		c.lastError = err
		return false
	}

	c.Logger.Debugf("[WSF Cache]: Remove item '%s'", nil, id)
	if err := c.Backend.Remove(id); err != nil {
		c.lastError = err
		return false
	}

	return true
}

// Clear cache
func (c *Core) Clear(mode int64, tags []string) bool {
	if !c.Options.Enable {
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

	if err := c.Backend.Clear(mode, tags); err != nil {
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
	if id != "" && c.Options.CacheIDPrefix != "" {
		return c.Options.CacheIDPrefix + id
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
		Options: cfg,
	}

	adp, err := backend.NewBackendCache(cfg.Backend.GetString("type"), cfg.Backend)
	if err != nil {
		return nil, errors.Wrap(err, "[Core] Unable to create underliyng backend")
	}
	cc.Backend = adp

	return cc, nil
}
