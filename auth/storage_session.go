package auth

import (
	"sync"
	"wsf/context"
	"wsf/errors"
	"wsf/session"
)

// Public constants
const (
	TYPEStorageSession = "session"
)

func init() {
	RegisterStorage(TYPEStorageSession, NewStorageSession)
}

// SessionStorage is an auth storage in session
type SessionStorage struct {
	Options   *StorageConfig
	Session   session.ManagerInterface
	Namespace string
	Member    string
	mu        sync.Mutex
}

// Setup the object
func (s *SessionStorage) Setup() error {
	s.Session = session.Instance()
	s.Namespace = s.Options.Namespace
	s.Member = s.Options.Member
	return nil
}

// IsEmpty returns true if storage is empty
func (s *SessionStorage) IsEmpty(ctx context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	untypedSessID := ctx.Value(context.SessionIDKey)
	if sessID, ok := untypedSessID.(string); ok {
		if sess, ok := s.Session.SessionGet(sessID); ok {
			return !sess.Has(s.Member)
		}
	}

	return true
}

// Read data from storage
func (s *SessionStorage) Read(ctx context.Context) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	untypedSessID := ctx.Value(context.SessionIDKey)
	if sessID, ok := untypedSessID.(string); ok {
		if sess, ok := s.Session.SessionGet(sessID); ok {
			contents := sess.Get(s.Member)
			if contents == nil {
				return nil, errors.New("Empty identity")
			} else if v, ok := contents.(map[string]interface{}); ok {
				return v, nil
			}
		}
	}

	return nil, errors.New("Identity does not exist")
}

// Write data to storage
func (s *SessionStorage) Write(ctx context.Context, contents map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	untypedSessID := ctx.Value(context.SessionIDKey)
	if sessID, ok := untypedSessID.(string); ok {
		if sess, ok := s.Session.SessionGet(sessID); ok {
			return sess.Set(s.Member, contents)
		}
	}

	return errors.New("Identity does not exist")
}

// Clear storage
func (s *SessionStorage) Clear(ctx context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	untypedSessID := ctx.Value(context.SessionIDKey)
	if sessID, ok := untypedSessID.(string); ok {
		if sess, ok := s.Session.SessionGet(sessID); ok {
			return sess.Unset(s.Member)
		}
	}

	return false
}

// ClearAll storage
func (s *SessionStorage) ClearAll() bool {
	return false
}

// NewStorageSession creates a new auth storage of type session
func NewStorageSession(options *StorageConfig) (Storage, error) {
	s := &SessionStorage{}
	s.Options = options
	s.Setup()

	return s, nil
}
