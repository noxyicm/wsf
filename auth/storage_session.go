package auth

import (
	"sync"
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
func (s *SessionStorage) IsEmpty(idnt string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.Session.SessionGet(idnt); ok {
		return !sess.Has(s.Member)
	}

	return true
}

// Read data from storage
func (s *SessionStorage) Read(idnt string) (Identity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.Session.SessionGet(idnt); ok {
		contents := sess.Get(s.Member)
		if contents == nil {
			return nil, errors.New("Empty identity")
		}

		return contents.(Identity), nil
	}

	return nil, errors.Errorf("Session '%s' does not exist", idnt)
}

// Write data to storage
func (s *SessionStorage) Write(idnt string, contents Identity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.Session.SessionGet(idnt); ok {
		return sess.Set(s.Member, contents)
	}

	return errors.Errorf("Session '%s' does not exist", idnt)
}

// Clear storage
func (s *SessionStorage) Clear(idnt string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.Session.SessionGet(idnt); ok {
		return sess.Unset(s.Member)
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
