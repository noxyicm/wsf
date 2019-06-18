package cache

// CoreInterface represents a core cache
type CoreInterface interface {
	Load(id string, doNotTestCacheValidity bool, doNotUnserialize bool)
	Save(data interface{}, id string, tags []string, specificLifetime int, priority int)
}
