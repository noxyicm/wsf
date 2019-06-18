package storage

import (
	"sync"
	"time"
)

// Memory is a in memory session storage
type Memory struct {
	sid                  string
	timeAccessed         time.Time
	writable             bool
	readable             bool
	writeClosed          bool
	sessionCookieDeleted bool
	destroyed            bool
	value                map[interface{}]interface{}
	lock                 sync.RWMutex
}
