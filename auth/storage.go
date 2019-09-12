package auth

import (
	"wsf/config"
	"wsf/errors"
)

var (
	buildStorageHandlers = map[string]func(*StorageConfig) (Storage, error){}
)

// Storage represents auth storage interface
type Storage interface {
	Setup() error
	IsEmpty(idnt string) bool
	Read(idnt string) (Identity, error)
	Write(idnt string, contents Identity) error
	Clear(idnt string) bool
	ClearAll() bool
}

// NewStorage creates a new auth storage from given type and options
func NewStorage(storageType string, options config.Config) (Storage, error) {
	cfg := &StorageConfig{}
	cfg.Defaults()
	cfg.Populate(options)

	if f, ok := buildStorageHandlers[storageType]; ok {
		return f(cfg)
	}

	return nil, errors.Errorf("Unrecognized auth storage type \"%v\"", storageType)
}

// NewStorageFromConfig creates a new auth storage from given type and StorageConfig
func NewStorageFromConfig(storageType string, options *StorageConfig) (Storage, error) {
	if f, ok := buildStorageHandlers[storageType]; ok {
		return f(options)
	}

	return nil, errors.Errorf("Unrecognized auth storage type \"%v\"", storageType)
}

// RegisterStorage registers a handler for auth storage creation
func RegisterStorage(storageType string, handler func(*StorageConfig) (Storage, error)) {
	buildStorageHandlers[storageType] = handler
}
