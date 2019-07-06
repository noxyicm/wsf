package resource

import (
	"wsf/cache"
	"wsf/config"
)

// TYPECache id of resource
const TYPECache = "cache"

func init() {
	Register(TYPECache, NewCacheResource)
}

// NewCacheResource creates a new resource of type Cache
func NewCacheResource(cfg config.Config) (Interface, error) {
	cacheType := cfg.GetString("type")
	cch, err := cache.NewCore(cacheType, cfg)
	if err != nil {
		return nil, err
	}

	return cch, nil
}
