package auth

import (
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
)

var (
	buildStorageHandlers = map[string]func(*StorageConfig) (Storage, error){}
)

// Storage represents auth storage interface
type Storage interface {
	Setup() error
	IsEmpty(ctx context.Context) bool
	Read(ctx context.Context) (map[string]interface{}, error)
	Write(ctx context.Context, contents map[string]interface{}) error
	Clear(ctx context.Context) bool
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
