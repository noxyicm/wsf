package backend

import (
	"container/list"
	"sync"
)

// FileGC is a ttl watcher
type FileGC struct {
	List   *list.List
	Table  map[string]*list.Element
	MaxLen int

	mu sync.Mutex
}
